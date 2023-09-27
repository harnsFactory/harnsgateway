package config

import (
	"harnsgateway/pkg/collector"
	modbusruntime "harnsgateway/pkg/protocol/modbus/runtime"
	opcuaruntime "harnsgateway/pkg/protocol/opcua/runtime"
	s7runtime "harnsgateway/pkg/protocol/s7/runtime"
	"harnsgateway/pkg/runtime"
)

var DeviceTypeObjectMap = map[string]runtime.RunObject{
	"modbusTcp": &modbusruntime.ModBusDevice{},
	"opcUa":     &opcuaruntime.OpcUaDevice{},
	"s71500":    &s7runtime.S7Device{},
}

type Config struct {
	CollectorMgr *collector.Manager
}
