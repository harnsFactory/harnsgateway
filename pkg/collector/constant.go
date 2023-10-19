package collector

import (
	"harnsgateway/pkg/protocol/modbus"
	"harnsgateway/pkg/protocol/modbusrtu"
	"harnsgateway/pkg/protocol/opcua"
	"harnsgateway/pkg/protocol/s7"
	"time"
)

var DeviceManagers = map[string]DeviceManager{
	"modbusTcp": &modbus.ModbusDeviceManager{},
	"opcUa":     &opcua.OpcUaDeviceManager{},
	"s71500":    &s7.S7DeviceManager{},
	"modbusRtu": &modbusrtu.ModbusRtuDeviceManager{},
}

const (
	mqttTimeout = 3 * time.Second
)
