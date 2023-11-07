package model

import (
	modbus "harnsgateway/pkg/protocol/modbus/runtime"
	"harnsgateway/pkg/runtime"
)

type ModbusRtuOverTcp struct {
}

func (m *ModbusRtuOverTcp) NewClients(address *modbus.Address, dataFrameCount int) (*modbus.Clients, error) {
	// TODO implement me
	panic("implement me")
}

func (m *ModbusRtuOverTcp) GenerateReadMessage(slave uint, functionCode uint8, startAddress uint, maxDataSize uint, variables []*modbus.VariableParse, memoryLayout runtime.MemoryLayout) *modbus.ModBusDataFrame {
	// TODO implement me
	panic("implement me")
}

func (m *ModbusRtuOverTcp) ExecuteAction(messenger modbus.Messenger, variables []*modbus.Variable) error {
	// TODO implement me
	panic("implement me")
}
