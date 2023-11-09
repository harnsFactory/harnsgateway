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
	defer func() {
		if tunnel != nil {
			_ = tunnel.Close()
		}
	}()
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
		// TPKT 共4个字节
		0x03,       // 版本号
		0x00,       // 预留字段
		0x00, 0x16, // 报文总长度 这里固定22
		// COTP 共18个字节
		0x11,       // 该字节之后的报文总长度
		0xe0,       // PDU类型 [(0xe0连接确认),(0xd0连接确认),(0x80断开请求),(0xc0断开确认),(0x50拒绝),(0xf0数据)]
		0x00, 0x00, // Destination reference 目标引用 用来唯一标识目标
		0x00, 0x01, // Source reference 源的引用
		0x00,       // 前四位标识Class,倒数第二位对应Extended formats是否使用拓展样式,倒数第一位对应No explicit flow control是否有明确的指定流控制
		0xc0,       // parameter code: tpdu-size 参数代码 TPDU-SIZE
		0x01,       // parameter length 参数长度
		0x0a,       // TPDU size TPDU大小
		0xc1,       // parameter code:src-tsap
		0x02,       // parameter length
		0x01, 0x00, // source TSAP   SourceTSAP/Rack
		0xc2,         // parameter code:dst-tsap
		0x02,         // parameter length
		0x01,         // Destination TSAP connectionType   01PG  02OP 03s7单边 0x10s7双边
		rackSlotByte, // rack & solt
	}
	return bytes
}

func newS7COMMSetupMessage() []byte {
	setupBytes := []byte{
		// TPKT
		0x03,       // 协议号
		0x00,       // 预留字段
		0x00, 0x19, // 总字节数
		// COTP
		0x02, // 该字节之后的COTP报文长度
		0xf0, // PDU类型 [(0xe0连接确认),(0xd0连接确认),(0x80断开请求),(0xc0断开确认),(0x50拒绝),(0xf0数据)]
		0x80, // Destination reference 首位：是否最后一个数据 后7位： TPDU**编号
		// S7 header
		0x32,       // 协议id
		0x01,       // pdu类型 [{0x01-job},{0x02-ack},{0x03-ack-data},{0x07-Userdata}]
		0x00, 0x00, // 保留字段
		0x04, 0x00, // Protocol Data Unit Reference |pdu的参考–由主站生成，每次新传输递增（大端）
		0x00, 0x08, // 参数长度
		0x00, 0x00, // 数据长度
		// S7 parameter
		0xf0,       // [{0xF0设置通信}]
		0x00,       // 预留
		0x00, 0x01, // Ack队列的大小（主叫）（大端）
		0x00, 0x01, // Ack队列的大小（被叫）（大端）
		0x01, 0xe0, // pdu长度
	}
	return setupBytes
}
