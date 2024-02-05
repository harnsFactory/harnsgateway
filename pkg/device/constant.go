package device

import (
	"harnsgateway/pkg/protocol/modbus"
	"harnsgateway/pkg/protocol/opcua"
	"harnsgateway/pkg/protocol/s7"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"time"
)

var DeviceManagers = map[string]DeviceManager{
	"modbus": &modbus.ModbusDeviceManager{},
	"opcUa":  &opcua.OpcUaDeviceManager{},
	"s7":     &s7.S7DeviceManager{},
}

var patchTypes = sets.NewString(string(types.JSONPatchType), string(types.MergePatchType))

const (
	maxJSONPatchOperations = 1000
	mqttTimeout            = 1 * time.Second
	heartBeatTimeInterval  = 15 * time.Second
)
