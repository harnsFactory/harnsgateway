package opcua

import (
	"container/list"
	"context"
	"github.com/gopcua/opcua"
	opcuaruntime "harnsgateway/pkg/protocol/opcua/runtime"
	"k8s.io/klog/v2"
	"sync"
)

type Tunnels struct {
	newClient    func() (*opcua.Client, error)
	clients      *list.List
	Max          int
	Idle         int
	Mux          *sync.Mutex
	ConnRequests map[uint64]chan *opcua.Client
	NextRequest  uint64
}

func (t *Tunnels) getTunnel(ctx context.Context) (*opcua.Client, error) {
	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	t.Mux.Lock()
	if t.Idle > 0 {
		t.Idle = t.Idle - 1
		front := t.clients.Front()
		client := front.Value.(*opcua.Client)
		t.clients.Remove(front)
		if client.State() == opcua.Closed || client.State() == opcua.Disconnected {
			client.Close(ctx)
			newClient, err := t.newClient()
			if err != nil {
				klog.V(2).InfoS("Failed to connect opc ua server", "error", err)
				return nil, opcuaruntime.ErrConnectOpuServer
			}
			return newClient, nil
		}
		t.Mux.Unlock()
		return client, nil
	}

	cCh := make(chan *opcua.Client, 1)
	key := t.nextRequestKey()
	t.ConnRequests[key] = cCh
	t.Mux.Unlock()

	select {
	case <-ctx.Done():
		t.Mux.Lock()
		delete(t.ConnRequests, key)
		t.Mux.Unlock()
		select {
		default:
		case c, ok := <-cCh:
			if ok && c.State() != opcua.Closed {
				t.clients.PushBack(c)
			}
		}
		return nil, ctx.Err()
	case c, ok := <-cCh:
		if !ok {
			return nil, opcuaruntime.ErrTcpClosed
		}
		return c, nil
	}
}

func (t *Tunnels) releaseTunnel(client *opcua.Client) {
	t.Mux.Lock()
	defer t.Mux.Unlock()
	if t.Idle == 0 && len(t.ConnRequests) > 0 {
		var cCh chan *opcua.Client
		var key uint64
		for key, cCh = range t.ConnRequests {
			break
		}
		delete(t.ConnRequests, key)
		cCh <- client
	} else {
		t.clients.PushBack(client)
		t.Idle = t.Idle + 1
	}
}

func (t *Tunnels) Destroy(ctx context.Context) {
	t.Mux.Lock()
	defer t.Mux.Unlock()
	for t.clients.Len() > 0 {
		e := t.clients.Front()
		c := e.Value.(*opcua.Client)
		c.Close(ctx)
		t.clients.Remove(e)
	}

	for _, clientsRequest := range t.ConnRequests {
		close(clientsRequest)
	}
}

func (t *Tunnels) nextRequestKey() uint64 {
	next := t.NextRequest
	t.NextRequest++
	return next
}
