package model

import (
	"container/list"
	"fmt"
	modbus "harnsgateway/pkg/protocol/modbus/runtime"
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/utils/binutil"
	"harnsgateway/pkg/utils/crcutil"
	"k8s.io/klog/v2"
	"net"
	"sync"
)

const RtuOverTcpNonDataLength = 5

type ModbusRtuOverTcp struct {
}

func (m *ModbusRtuOverTcp) NewClients(address *modbus.Address, dataFrameCount int) (*modbus.Clients, error) {
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

func (m *ModbusRtuOverTcp) GenerateReadMessage(slave uint, functionCode uint8, startAddress uint, maxDataSize uint, variables []*modbus.VariableParse, memoryLayout runtime.MemoryLayout) *modbus.ModBusDataFrame {
	// 01 03 00 00 00 0A C5 CD
	// 01  设备地址
	// 03  功能码
	// 00 00  起始地址
	// 00 0A  寄存器数量(word数量)/线圈数量
	// C5 CD  crc16检验码
	message := make([]byte, 6)
	message[0] = byte(slave)
	message[1] = functionCode
	binutil.WriteUint16BigEndian(message[2:], uint16(startAddress))
	binutil.WriteUint16BigEndian(message[4:], uint16(maxDataSize))
	crc16 := make([]byte, 2)
	binutil.WriteUint16BigEndian(crc16, crcutil.CheckCrc16sum(message))
	message = append(message, crc16...)

	bytesLength := 0
	switch modbus.FunctionCode(functionCode) {
	case modbus.ReadCoilStatus, modbus.ReadInputStatus:
		if maxDataSize%8 == 0 {
			bytesLength = int(maxDataSize/8 + RtuOverTcpNonDataLength)
		} else {
			bytesLength = int(maxDataSize/8 + 1 + RtuOverTcpNonDataLength)
		}
	case modbus.ReadHoldRegister, modbus.ReadInputRegister:
		bytesLength = int(maxDataSize*2 + RtuOverTcpNonDataLength)
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
