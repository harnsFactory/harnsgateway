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
	newMessenger func() (*Messenger, error)
	Messengers   *list.List
	Max          int
	Idle         int
	Mux          *sync.Mutex
	ConnRequests map[uint64]chan *Messenger
	NextRequest  uint64
}

type Messenger struct {
	Timeout int
	Tunnel  *opcua.Client
}

func (t *Tunnels) getTunnel(ctx context.Context) (*Messenger, error) {
	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	t.Mux.Lock()
	if t.Idle > 0 {
		t.Idle = t.Idle - 1
		front := t.Messengers.Front()
		messenger := front.Value.(*Messenger)
		t.Messengers.Remove(front)
		if messenger.Tunnel.State() == opcua.Closed || messenger.Tunnel.State() == opcua.Disconnected {
			messenger.Tunnel.Close(ctx)
			newMessenger, err := t.newMessenger()
			if err != nil {
				klog.V(2).InfoS("Failed to connect opc ua server", "error", err)
				return nil, opcuaruntime.ErrConnectOpuServer
			}
			return newMessenger, nil
		}
		t.Mux.Unlock()
		return messenger, nil
	}

	cCh := make(chan *Messenger, 1)
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
		case m, ok := <-cCh:
			if ok && m.Tunnel.State() != opcua.Closed {
				t.Messengers.PushBack(m)
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

func (t *Tunnels) releaseTunnel(messenger *Messenger) {
	t.Mux.Lock()
	defer t.Mux.Unlock()
	if t.Idle == 0 && len(t.ConnRequests) > 0 {
		var cCh chan *Messenger
		var key uint64
		for key, cCh = range t.ConnRequests {
			break
		}
		delete(t.ConnRequests, key)
		cCh <- messenger
	} else {
		t.Messengers.PushBack(messenger)
		t.Idle = t.Idle + 1
	}
}

func (t *Tunnels) Destroy(ctx context.Context) {
	t.Mux.Lock()
	defer t.Mux.Unlock()
	for t.Messengers.Len() > 0 {
		e := t.Messengers.Front()
		messenger := e.Value.(*Messenger)
		messenger.Tunnel.Close(ctx)
		t.Messengers.Remove(e)
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
