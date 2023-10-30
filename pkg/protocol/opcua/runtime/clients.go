package runtime

import (
	"container/list"
	"context"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"k8s.io/klog/v2"
	"sync"
)

type Clients struct {
	NewMessenger func() (Messenger, error)
	Messengers   *list.List
	Max          int
	Idle         int
	Mux          *sync.Mutex
	ConnRequests map[uint64]chan Messenger
	NextRequest  uint64
}

type Messenger interface {
	Read(ctx context.Context, req *ua.ReadRequest) (*ua.ReadResponse, error)
	Close(ctx context.Context)
	Available() bool
	Reset(messenger Messenger)
}

type UaClient struct {
	Timeout int
	Client  *opcua.Client
}

func (u *UaClient) Read(ctx context.Context, req *ua.ReadRequest) (*ua.ReadResponse, error) {
	return u.Client.Read(ctx, req)
}

func (u *UaClient) Close(ctx context.Context) {
	_ = u.Client.Close(ctx)
}

func (u *UaClient) Available() bool {
	if u.Client.State() == opcua.Closed || u.Client.State() == opcua.Disconnected {
		return false
	}
	return true
}

func (u *UaClient) Reset(messenger Messenger) {
	nu := (messenger).(*UaClient)
	u.Client = nu.Client
}

func (cs *Clients) GetMessenger(ctx context.Context) (Messenger, error) {
	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	cs.Mux.Lock()
	if cs.Idle > 0 {
		cs.Idle = cs.Idle - 1
		front := cs.Messengers.Front()
		messenger := front.Value.(Messenger)
		cs.Messengers.Remove(front)
		if !messenger.Available() {
			newMessenger, err := cs.NewMessenger()
			if err != nil {
				klog.V(2).InfoS("Failed to connect opc ua server", "error", err)
				return nil, ErrConnectOpuServer
			}
			return newMessenger, nil
		}
		cs.Mux.Unlock()
		return messenger, nil
	}

	cCh := make(chan Messenger, 1)
	key := cs.nextRequestKey()
	cs.ConnRequests[key] = cCh
	cs.Mux.Unlock()

	select {
	case <-ctx.Done():
		cs.Mux.Lock()
		delete(cs.ConnRequests, key)
		cs.Mux.Unlock()
		select {
		default:
		case m, ok := <-cCh:
			if ok && m.Available() {
				cs.Messengers.PushBack(m)
			}
		}
		return nil, ctx.Err()
	case c, ok := <-cCh:
		if !ok {
			return nil, ErrTcpClosed
		}
		return c, nil
	}
}

func (cs *Clients) ReleaseMessenger(messenger Messenger) {
	cs.Mux.Lock()
	defer cs.Mux.Unlock()
	if cs.Idle == 0 && len(cs.ConnRequests) > 0 {
		var cCh chan Messenger
		var key uint64
		for key, cCh = range cs.ConnRequests {
			break
		}
		delete(cs.ConnRequests, key)
		cCh <- messenger
	} else {
		cs.Messengers.PushBack(messenger)
		cs.Idle = cs.Idle + 1
	}
}

func (cs *Clients) Destroy(ctx context.Context) {
	cs.Mux.Lock()
	defer cs.Mux.Unlock()
	for cs.Messengers.Len() > 0 {
		e := cs.Messengers.Front()
		messenger := e.Value.(Messenger)
		messenger.Close(ctx)
		cs.Messengers.Remove(e)
	}

	for _, clientsRequest := range cs.ConnRequests {
		close(clientsRequest)
	}
}

func (cs *Clients) nextRequestKey() uint64 {
	next := cs.NextRequest
	cs.NextRequest++
	return next
}
