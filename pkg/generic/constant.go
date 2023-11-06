package generic

import (
	"harnsgateway/pkg/protocol/modbus"
	modbusallruntime "harnsgateway/pkg/protocol/modbus/runtime"
	"harnsgateway/pkg/protocol/opcua"
	opcuaruntime "harnsgateway/pkg/protocol/opcua/runtime"
	"harnsgateway/pkg/protocol/s7"
	s7runtime "harnsgateway/pkg/protocol/s7/runtime"
	"harnsgateway/pkg/runtime"
	v1 "harnsgateway/pkg/v1"
)

var DeviceTypeMap = map[string]func() v1.DeviceType{
	"modbus": func() v1.DeviceType { return &v1.ModBusDevice{} },
	"opcUa":  func() v1.DeviceType { return &v1.OpcUaDevice{} },
	"s7":     func() v1.DeviceType { return &v1.S7Device{} },
}

var DeviceTypeObjectMap = map[string]runtime.Device{
	"modbus": &modbusallruntime.ModBusDevice{},
	"opcUa":  &opcuaruntime.OpcUaDevice{},
	"s7":     &s7runtime.S7Device{},
}

type NewBroker func(object runtime.Device) (runtime.Broker, chan *runtime.ParseVariableResult, error)

var DeviceTypeBrokerMap = map[string]NewBroker{
	"modbus": modbus.NewBroker,
	"opcUa":  opcua.NewBroker,
	"s7":     s7.NewBroker,
}
