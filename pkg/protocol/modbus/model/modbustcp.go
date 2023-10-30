package model

import (
	"container/list"
	"fmt"
	modbus "harnsgateway/pkg/protocol/modbus/runtime"
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/utils/binutil"
	"k8s.io/klog/v2"
	"net"
	"sync"
)

const TcpNonDataLength = 9

type ModbusTcp struct {
}

func (m *ModbusTcp) NewClients(address *modbus.Address, dataFrameCount int) (*modbus.Clients, error) {
	tcpChannel := dataFrameCount/5 + 1
	addr := fmt.Sprintf("%s:%d", address.Location, address.Option.Port)
	cs := list.New()
	for i := 0; i < tcpChannel; i++ {
		tunnel, err := net.Dial("tcp", addr)
		if err != nil {
			klog.V(2).InfoS("Failed to connect modbus server", "error", err)
			return nil, err
		}
		c := &modbus.TcpClient{
			Tunnel:  tunnel,
			Timeout: 1,
		}
		cs.PushBack(c)
	}

	clients := &modbus.Clients{
		Messengers:   cs,
		Max:          tcpChannel,
		Idle:         tcpChannel,
		Mux:          &sync.Mutex{},
		NextRequest:  1,
		ConnRequests: make(map[uint64]chan modbus.Messenger, 0),
		NewMessenger: func() (modbus.Messenger, error) {
			tunnel, err := net.Dial("tcp", addr)
			if err != nil {
				klog.V(2).InfoS("Failed to connect modbus server", "error", err)
				return nil, err
			}
			return &modbus.TcpClient{
				Tunnel:  tunnel,
				Timeout: 1,
			}, nil
		},
	}
	return clients, nil
}

func (m *ModbusTcp) GenerateReadMessage(slave uint, functionCode uint8, startAddress uint, maxDataSize uint, variables []*modbus.VariableParse, memoryLayout runtime.MemoryLayout) *modbus.ModBusDataFrame {
	// 00 01 00 00 00 06 18 03 00 02 00 02
	// 00 01  此次通信事务处理标识符，一般每次通信之后将被要求加1以区别不同的通信数据报文
	// 00 00  表示协议标识符，00 00为modbus协议
	// 00 06  数据长度，用来指示接下来数据的长度，单位字节
	// 18  设备地址，用以标识连接在串行线或者网络上的远程服务端的地址。以上七个字节也被称为modbus报文头
	// 03  功能码，此时代码03为读取保持寄存器数据
	// 00 02  起始地址
	// 00 02  寄存器数量(word数量)/线圈数量
	message := make([]byte, 12)

	binutil.WriteUint16(message[2:], 0) // 协议版本
	binutil.WriteUint16(message[4:], 6) // 剩余长度
	message[6] = byte(slave)
	message[7] = functionCode
	binutil.WriteUint16(message[8:], uint16(startAddress))
	binutil.WriteUint16(message[10:], uint16(maxDataSize))

	bytesLength := 0
	switch modbus.FunctionCode(functionCode) {
	case modbus.ReadCoilStatus, modbus.WriteCoilStatus:
		if maxDataSize%8 == 0 {
			bytesLength = int(maxDataSize/8 + TcpNonDataLength)
		} else {
			bytesLength = int(maxDataSize/8 + 1 + TcpNonDataLength)
		}
	case modbus.ReadHoldRegister, modbus.WriteHoldRegister:
		bytesLength = int(maxDataSize*2 + TcpNonDataLength)
	}

	df := &modbus.ModBusDataFrame{
		Slave:             slave,
		MemoryLayout:      memoryLayout,
		StartAddress:      startAddress,
		FunctionCode:      functionCode,
		MaxDataSize:       maxDataSize,
		TransactionId:     0,
		DataFrame:         message,
		ResponseDataFrame: make([]byte, bytesLength),
		Variables:         make([]*modbus.VariableParse, 0, len(variables)),
	}
	df.Variables = append(df.Variables, variables...)

	return df
}
