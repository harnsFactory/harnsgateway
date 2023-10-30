package model

import (
	modbus "harnsgateway/pkg/protocol/modbus/runtime"
	"harnsgateway/pkg/runtime"
)

var _ ModbusModeler = (*ModbusTcp)(nil)
var _ ModbusModeler = (*ModbusRtu)(nil)
var _ ModbusModeler = (*ModbusRtuOverTcp)(nil)

var ModbusModelers = map[string]ModbusModeler{
	"modbusTcp":        &ModbusTcp{},
	"modbusRtu":        &ModbusRtu{},
	"modbusRtuOverTcp": &ModbusRtuOverTcp{},
}

type ModbusModeler interface {
	GenerateReadMessage(slave uint, functionCode uint8, startAddress uint, maxDataSize uint, variables []*modbus.VariableParse, memoryLayout runtime.MemoryLayout) *modbus.ModBusDataFrame
	NewClients(address *modbus.Address, dataFrameCount int) (*modbus.Clients, error)
}
