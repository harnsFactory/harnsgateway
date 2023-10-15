package s7

import (
	"container/list"
	"context"
	modbusruntime "harnsgateway/pkg/protocol/modbus/runtime"
	s7runtime "harnsgateway/pkg/protocol/s7/runtime"
	"io"
	"k8s.io/klog/v2"
	"net"
	"sync"
	"time"
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
		t.Mux.Unlock()
		return messenger, nil
	}

	mCh := make(chan *Messenger, 1)
	key := t.nextRequestKey()
	t.ConnRequests[key] = mCh
	t.Mux.Unlock()

	select {
	case <-ctx.Done():
		t.Mux.Lock()
		delete(t.ConnRequests, key)
		t.Mux.Unlock()
		select {
		default:
		case m, ok := <-mCh:
			if ok && m.Tunnel != nil {
				t.Messengers.PushBack(m)
			}
		}
		return nil, ctx.Err()
	case m, ok := <-mCh:
		if !ok {
			return nil, modbusruntime.ErrTcpClosed
		}
		return m, nil
	}
}

func (t *Tunnels) releaseTunnel(messenger *Messenger) {
	t.Mux.Lock()
	defer t.Mux.Unlock()
	if t.Idle == 0 && len(t.ConnRequests) > 0 {
		var mCh chan *Messenger
		var key uint64
		for key, mCh = range t.ConnRequests {
			break
		}
		delete(t.ConnRequests, key)
		mCh <- messenger
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
		m := e.Value.(*Messenger)
		m.Tunnel.Close()
		t.Messengers.Remove(e)
	}

	for _, messengersRequest := range t.ConnRequests {
		close(messengersRequest)
	}
}

func (t *Tunnels) nextRequestKey() uint64 {
	next := t.NextRequest
	t.NextRequest++
	return next
}

type Messenger struct {
	Timeout int
	Tunnel  net.Conn
}

func (m *Messenger) AskAtLeast(request []byte, response []byte, min int) (int, error) {
	_, err := m.Tunnel.Write(request)
	if err != nil {
		klog.V(2).InfoS("Failed to ask message", "error", err)
		return 0, s7runtime.ErrBadConn
	}
	// 设置读超时
	deadLineTime := time.Now().Add(time.Duration(m.Timeout) * time.Second)

	err = m.Tunnel.SetReadDeadline(deadLineTime)
	if err != nil {
		klog.V(2).InfoS("Tcp connect timeout", "error", err)
		return 0, err
	}
	return io.ReadAtLeast(m.Tunnel, response, min)
}
