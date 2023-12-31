package device

import (
	"harnsgateway/pkg/protocol/modbus"
	"harnsgateway/pkg/protocol/opcua"
	"harnsgateway/pkg/protocol/s7"
	"time"
)

var DeviceManagers = map[string]DeviceManager{
	"modbus": &modbus.ModbusDeviceManager{},
	"opcUa":  &opcua.OpcUaDeviceManager{},
	"s7":     &s7.S7DeviceManager{},
}

const (
	mqttTimeout           = 1 * time.Second
	heartBeatTimeInterval = 15 * time.Second
)
