package model

import (
	"container/list"
	"fmt"
	s7 "harnsgateway/pkg/protocol/s7/runtime"
	"harnsgateway/pkg/utils/binutil"
	"io"
	"k8s.io/klog/v2"
	"net"
	"sync"
)

type S71500 struct {
}

func (s *S71500) GetS7DevicePDULength(address *s7.S7Address) (uint16, error) {
	addr := fmt.Sprintf("%s:%d", address.Location, address.Option.Port)
	tunnel, err := net.Dial("tcp", addr)
	defer tunnel.Close()
	if err != nil {
		klog.V(2).InfoS("Failed to connect s7 device", "address", addr, "error", err)
		return 0, err
	}

	_, err = tunnel.Write(newCOTPConnectMessage(address.Option.Rack, address.Option.Slot))
	if err != nil {
		klog.V(2).InfoS("Failed to connect s7 device passed COTP message", "error", err)
		return 0, err
	}
	cotpResponse := make([]byte, 22)
	_, err = io.ReadAtLeast(tunnel, cotpResponse, 22)
	if err != nil {
		klog.V(2).InfoS("Failed to connect s7 device passed COTP message", "error", err)
		return 0, err
	}
	if int(cotpResponse[5]) != 208 {
		return 0, s7.ErrConnectS7DeviceCotpMessage
	}
	_, err = tunnel.Write(newS7COMMSetupMessage())
	if err != nil {
		klog.V(2).InfoS("Failed to connect s7 device passed S7COMM message", "error", err)
		return 0, err
	}
	s7Response := make([]byte, 27)
	_, err = io.ReadAtLeast(tunnel, s7Response, 27)
	if err != nil {
		klog.V(2).InfoS("Failed to connect s7 device passed S7COMM message", "error", err)
		return 0, err
	}
	if s7Response[8] != 3 {
		errClass := int(s7Response[17])
		errCode := int(s7Response[18])
		klog.V(2).InfoS("Failed to connect s7 device passed S7COMM message", "errClass", errClass, "errCode", errCode, "error", err)
		return 0, s7.ErrConnectS7DeviceS7COMMMessage
	}

	responsePduLength := s7Response[25:]
	return binutil.ParseUint16(responsePduLength), nil

}

func (s *S71500) NewClients(address *s7.S7Address, dataFrameCount int) (*s7.Clients, error) {
	tcpChannel := dataFrameCount/5 + 1

	addr := fmt.Sprintf("%s:%d", address.Location, address.Option.Port)
	ms := list.New()
	for i := 0; i < tcpChannel; i++ {
		m, err := newS7Messenger(addr, address.Option.Rack, address.Option.Slot)
		if err != nil {
			klog.V(2).InfoS("Failed to connect s7 server", "error", err)
			return nil, err
		}
		ms.PushBack(m)
	}

	clients := &s7.Clients{
		Messengers:   ms,
		Max:          tcpChannel,
		Idle:         tcpChannel,
		Mux:          &sync.Mutex{},
		NextRequest:  1,
		ConnRequests: make(map[uint64]chan s7.Messenger, 0),
		NewMessenger: func() (s7.Messenger, error) {
			return newS7Messenger(addr, address.Option.Rack, address.Option.Slot)
		},
	}
	return clients, nil
}

func newS7Messenger(addr string, rack uint8, slot uint8) (s7.Messenger, error) {
	tunnel, err := net.Dial("tcp", addr)
	if err != nil {
		klog.V(2).InfoS("Failed to connect s7 device", "address", addr, "error", err)
		return nil, err
	}

	_, err = tunnel.Write(newCOTPConnectMessage(rack, slot))
	if err != nil {
		klog.V(2).InfoS("Failed to connect s7 device passed COTP message", "error", err)
		return nil, err
	}
	cotpResponse := make([]byte, 22)
	_, err = io.ReadAtLeast(tunnel, cotpResponse, 22)
	if err != nil {
		klog.V(2).InfoS("Failed to connect s7 device passed COTP message", "error", err)
		return nil, err
	}
	if int(cotpResponse[5]) != 208 {
		return nil, s7.ErrConnectS7DeviceCotpMessage
	}
	_, err = tunnel.Write(newS7COMMSetupMessage())
	if err != nil {
		klog.V(2).InfoS("Failed to connect s7 device passed S7COMM message", "error", err)
		return nil, err
	}
	s7Response := make([]byte, 27)
	_, err = io.ReadAtLeast(tunnel, s7Response, 27)
	if err != nil {
		klog.V(2).InfoS("Failed to connect s7 device passed S7COMM message", "error", err)
		return nil, err
	}
	if s7Response[8] != 3 {
		errClass := int(s7Response[17])
		errCode := int(s7Response[18])
		klog.V(2).InfoS("Failed to connect s7 device passed S7COMM message", "errClass", errClass, "errCode", errCode, "error", err)
		return nil, s7.ErrConnectS7DeviceS7COMMMessage
	}
	return &s7.TcpClient{
		Tunnel:  tunnel,
		Timeout: 1,
	}, nil
}

func newCOTPConnectMessage(rack uint8, slot uint8) []byte {
	rackSlotByte := (rack*2)<<4 + slot
	bytes := []byte{
		0x03, 0x00, 0x00, 0x16, // 总字节数 固定22
		0x11, 0xe0, 0x00, 0x00,
		0x00, 0x01, 0x00, 0xc0,
		0x01, 0x0a, 0xc1, 0x02,
		0x01, 0x02, 0xc2, 0x02,
		0x01,         // Destination TSAP connectionType   01PG  02OP 03s7单边 0x10s7双边
		rackSlotByte, // rack & solt
	}
	return bytes
}

func newS7COMMSetupMessage() []byte {
	setupBytes := []byte{
		// TPKT
		0x03, 0x00, 0x00, 0x19, // 总字节数
		// COTP
		0x02, 0xf0, 0x80,
		// S7 header
		0x32, 0x01, 0x00, 0x00, 0x04, 0x00, 0x00, 0x08, 0x00, 0x00,
		// S7 parameter
		0xf0, 0x00, 0x00, 0x01, 0x00, 0x01, 0x01, 0xe0,
	}
	return setupBytes
}
