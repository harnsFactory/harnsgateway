package generic

import (
	// "harnsgateway/pkg/collector"
	// "harnsgateway/pkg/collector"
	"harnsgateway/pkg/protocol/modbus"
	modbusruntime "harnsgateway/pkg/protocol/modbus/runtime"
	"harnsgateway/pkg/protocol/opcua"
	opcuaruntime "harnsgateway/pkg/protocol/opcua/runtime"
	"harnsgateway/pkg/protocol/s7"
	s7runtime "harnsgateway/pkg/protocol/s7/runtime"
	"harnsgateway/pkg/runtime"
	v1 "harnsgateway/pkg/v1"
)

var DeviceTypeMap = map[string]func() v1.DeviceType{
	"modbusTcp": func() v1.DeviceType { return &v1.ModBusDevice{} },
	"opcUa":     func() v1.DeviceType { return &v1.OpcUaDevice{} },
	"s71500":    func() v1.DeviceType { return &v1.S7Device{} },
}

var DeviceTypeObjectMap = map[string]runtime.RunObject{
	"modbusTcp": &modbusruntime.ModBusDevice{},
	"opcUa":     &opcuaruntime.OpcUaDevice{},
	"s71500":    &s7runtime.S7Device{},
}

type NewCollector func(object runtime.Device) (runtime.Collector, chan *runtime.ParseVariableResult, error)

var DeviceTypeCollectorMap = map[string]NewCollector{
	"modbusTcp": modbus.NewCollector,
	"opcUa":     opcua.NewCollector,
	"s71500":    s7.NewCollector,
}
