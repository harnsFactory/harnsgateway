package model

import (
	"container/list"
	"context"
	"fmt"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	opc "harnsgateway/pkg/protocol/opcua/runtime"
	"k8s.io/klog/v2"
	"sync"
)

type OpcUa struct {
}

func (o *OpcUa) NewClients(address *opc.Address, dataFrameCount int) (*opc.Clients, error) {
	tcpChannel := dataFrameCount/5 + 1

	var endpoint string
	if address.Option.Port <= 0 {
		endpoint = address.Location
	} else {
		endpoint = fmt.Sprintf("%s:%d", address.Location, address.Option.Port)
	}

	ms := list.New()
	for i := 0; i < tcpChannel; i++ {
		c, err := opcua.NewClient(endpoint, opcua.SecurityMode(ua.MessageSecurityModeNone))
		if err != nil {
			klog.V(2).InfoS("Failed to get opc ua client")
		}
		if err := c.Connect(context.Background()); err != nil {
			klog.V(2).InfoS("Failed to connect opc ua server")
		}
		m := &opc.UaClient{
			Timeout: 1,
			Client:  c,
		}
		ms.PushBack(m)
	}

	clients := &opc.Clients{
		Messengers:   ms,
		Max:          tcpChannel,
		Idle:         tcpChannel,
		Mux:          &sync.Mutex{},
		NextRequest:  1,
		ConnRequests: make(map[uint64]chan opc.Messenger, 0),
		NewMessenger: func() (opc.Messenger, error) {
			c, err := opcua.NewClient(endpoint, opcua.SecurityMode(ua.MessageSecurityModeNone))
			if err != nil {
				klog.V(2).InfoS("Failed to get opc ua client")
			}
			if err := c.Connect(context.Background()); err != nil {
				klog.V(2).InfoS("Failed to connect opc ua server")
			}

			return &opc.UaClient{
				Timeout: 1,
				Client:  c,
			}, nil
		},
	}
	return clients, nil
}
