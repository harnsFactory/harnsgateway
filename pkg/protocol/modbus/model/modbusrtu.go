package model

import (
	"container/list"
	"go.bug.st/serial"
	modbus "harnsgateway/pkg/protocol/modbus/runtime"
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/utils/binutil"
	"harnsgateway/pkg/utils/crcutil"
	"k8s.io/klog/v2"
	"sync"
)

const RtuNonDataLength = 5

type ModbusRtu struct {
}

func (m *ModbusRtu) NewClients(address *modbus.Address, dataFrameCount int) (*modbus.Clients, error) {
	mode := &serial.Mode{
		BaudRate: address.Option.BaudRate,
		Parity:   modbus.ParityToParity[address.Option.Parity],
		DataBits: address.Option.DataBits,
		StopBits: modbus.StopBitsToStopBits[address.Option.StopBits],
	}
	port, err := serial.Open(address.Location, mode)
	if err != nil {
		klog.V(2).InfoS("Failed to connect serial port", "address", address.Location)
		return nil, err
	}

	cs := list.New()
	cs.PushBack(&modbus.SerialClient{
		Timeout: 1,
		Port:    port,
	})

	clients := &modbus.Clients{
		Messengers:   cs,
		Max:          1,
		Idle:         1,
		Mux:          &sync.Mutex{},
		NextRequest:  1,
		ConnRequests: make(map[uint64]chan modbus.Messenger, 0),
		NewMessenger: func() (modbus.Messenger, error) {
			newPort, err := serial.Open(address.Location, mode)
			if err != nil {
				klog.V(2).InfoS("Failed to connect serial port", "address", address.Location)
				return nil, err
			}
			return &modbus.SerialClient{
				Timeout: 1,
				Port:    newPort,
			}, nil
		},
	}
	return clients, nil
}

func (m *ModbusRtu) GenerateReadMessage(slave uint, functionCode uint8, startAddress uint, maxDataSize uint, variables []*modbus.VariableParse, memoryLayout runtime.MemoryLayout) *modbus.ModBusDataFrame {
	// 01 03 00 00 00 0A C5 CD
	// 01  设备地址
	// 03  功能码
	// 00 00  起始地址
	// 00 0A  寄存器数量(word数量)/线圈数量
	// C5 CD  crc16检验码
	message := make([]byte, 6)
	message[0] = byte(slave)
	message[1] = functionCode
	binutil.WriteUint16(message[2:], uint16(startAddress))
	binutil.WriteUint16(message[4:], uint16(maxDataSize))
	crc16 := make([]byte, 2)
	binutil.WriteUint16(crc16, crcutil.CheckCrc16sum(message))
	message = append(message, crc16...)

	bytesLength := 0
	switch modbus.FunctionCode(functionCode) {
	case modbus.ReadCoilStatus, modbus.ReadInputStatus:
		if maxDataSize%8 == 0 {
			bytesLength = int(maxDataSize/8 + RtuNonDataLength)
		} else {
			bytesLength = int(maxDataSize/8 + 1 + RtuNonDataLength)
		}
	case modbus.ReadHoldRegister, modbus.ReadInputRegister:
		bytesLength = int(maxDataSize*2 + RtuNonDataLength)
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

func (m *ModbusRtu) ExecuteAction(variables []*modbus.Variable) error {
	// TODO implement me
	panic("implement me")
}
