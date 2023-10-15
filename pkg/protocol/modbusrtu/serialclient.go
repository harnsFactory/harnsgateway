package modbusrtu

import (
	"container/list"
	"context"
	"go.bug.st/serial"
	modbusrturuntime "harnsgateway/pkg/protocol/modbusrtu/runtime"
	"k8s.io/klog/v2"
	"sync"
	"time"
)

type SerialClients struct {
	newSerialClient func() (*SerialClient, error)
	Clients         *list.List
	Max             int
	Idle            int
	Mux             *sync.Mutex
	ConnRequests    map[uint64]chan *SerialClient
	NextRequest     uint64
}

func (scs *SerialClients) getClient(ctx context.Context) (*SerialClient, error) {
	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	scs.Mux.Lock()
	if scs.Idle > 0 {
		scs.Idle = scs.Idle - 1
		front := scs.Clients.Front()
		client := front.Value.(*SerialClient)
		scs.Clients.Remove(front)
		scs.Mux.Unlock()
		return client, nil
	}

	cCh := make(chan *SerialClient, 1)
	key := scs.nextRequestKey()
	scs.ConnRequests[key] = cCh
	scs.Mux.Unlock()

	select {
	case <-ctx.Done():
		scs.Mux.Lock()
		delete(scs.ConnRequests, key)
		scs.Mux.Unlock()
		select {
		default:
		case c, ok := <-cCh:
			if ok && c.Port != nil {
				scs.Clients.PushBack(c)
			}
		}
		return nil, ctx.Err()
	case m, ok := <-cCh:
		if !ok {
			return nil, modbusrturuntime.ErrSerialPortClosed
		}
		return m, nil
	}
}

func (scs *SerialClients) releaseClient(client *SerialClient) {
	scs.Mux.Lock()
	defer scs.Mux.Unlock()
	if scs.Idle == 0 && len(scs.ConnRequests) > 0 {
		var cCh chan *SerialClient
		var key uint64
		for key, cCh = range scs.ConnRequests {
			break
		}
		delete(scs.ConnRequests, key)
		cCh <- client
	} else {
		scs.Clients.PushBack(client)
		scs.Idle = scs.Idle + 1
	}
}

func (scs *SerialClients) Destroy(ctx context.Context) {
	scs.Mux.Lock()
	defer scs.Mux.Unlock()
	for scs.Clients.Len() > 0 {
		e := scs.Clients.Front()
		c := e.Value.(*SerialClient)
		c.Port.Close()
		scs.Clients.Remove(e)
	}

	for _, clientRequest := range scs.ConnRequests {
		close(clientRequest)
	}
}

type SerialClient struct {
	Timeout int
	Port    serial.Port
}

func (scs *SerialClients) nextRequestKey() uint64 {
	next := scs.NextRequest
	scs.NextRequest++
	return next
}

func (sc *SerialClient) AskAtLeast(request []byte, response []byte) (int, error) {
	rql, err := sc.Port.Write(request)
	if err != nil {
		klog.V(2).InfoS("Failed to write byte to series port", "error", err)
		return 0, modbusrturuntime.ErrBadConn
	}
	klog.V(5).InfoS("Succeed to write byte to series port", "bytes", request, "length", rql)
	// 设置读超时
	deadLineTime := time.Duration(sc.Timeout) * time.Second
	err = sc.Port.SetReadTimeout(deadLineTime)
	if err != nil {
		klog.V(2).InfoS("Serial port connect timeout", "error", err)
		return 0, err
	}

	buf := make([]byte, 300)
	responseBytesLength := len(response)
	bytesLength := 0
	currentIndex := 0

	for {
		n, err := sc.Port.Read(buf)
		if err != nil {
			klog.V(2).InfoS("Failed to read byte from series port", "error", err)
			return 0, err
		}
		if n == 0 {
			break
		}
		bytesLength += n

		for i := 0; i < n; i++ {
			response[currentIndex] = buf[i]
			currentIndex++
		}

		if bytesLength == responseBytesLength {
			break
		}
	}
	if responseBytesLength != bytesLength {
		klog.V(2).InfoS("Modbus rtu data length no enough", "bytesLength", bytesLength)
		return 0, modbusrturuntime.ErrModbusRtuDataLengthNotEnough
	}

	return bytesLength, nil
}
