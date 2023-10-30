package collector

import (
	"harnsgateway/pkg/protocol/modbusall"
	"harnsgateway/pkg/protocol/opcua"
	"harnsgateway/pkg/protocol/s7"
	"time"
)

var DeviceManagers = map[string]DeviceManager{
	"modbus": &modbusall.ModbusDeviceManager{},
	"opcUa":  &opcua.OpcUaDeviceManager{},
	"s71500": &s7.S7DeviceManager{},
}

const (
	mqttTimeout = 3 * time.Second
)
