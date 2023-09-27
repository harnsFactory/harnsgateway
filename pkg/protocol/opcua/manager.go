package opcua

import (
	"fmt"
	opcuaruntime "harnsgateway/pkg/protocol/opcua/runtime"
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/utils/randutil"
	"harnsgateway/pkg/utils/uuidutil"
	v1 "harnsgateway/pkg/v1"
	"strconv"
	"time"
)

type OpcUaDeviceManager struct {
}

func (m *OpcUaDeviceManager) CreateDevice(deviceType v1.DeviceType) (runtime.RunObject, error) {
	opcUaDevice, ok := deviceType.(*v1.OpcUaDevice)
	if !ok {
		return nil, opcuaruntime.ErrDeviceType
	}

	d := &opcuaruntime.OpcUaDevice{
		DeviceMeta: runtime.DeviceMeta{
			ObjectMeta: runtime.ObjectMeta{
				Name:    opcUaDevice.Name,
				ID:      fmt.Sprintf("%s.%s", opcUaDevice.GetDeviceType(), uuidutil.UUID()),
				Version: strconv.FormatUint(randutil.Uint64n(), 10),
				ModTime: time.Now(),
			},
			DeviceCode:    opcUaDevice.DeviceCode,
			DeviceType:    opcUaDevice.DeviceType,
			CollectStatus: false,
		},
		CollectorCycle:   opcUaDevice.CollectorCycle,
		VariableInterval: opcUaDevice.VariableInterval,
		Address:          opcUaDevice.Address,
		Port:             opcUaDevice.Port,
		Username:         opcUaDevice.Username,
		Password:         opcUaDevice.Password,
	}
	if len(opcUaDevice.Variables) > 0 {
		for _, variable := range opcUaDevice.Variables {
			d.Variables = append(d.Variables, &opcuaruntime.Variable{
				DataType:     runtime.StringToDataType[variable.DataType],
				Name:         variable.Name,
				Address:      variable.Address,
				Namespace:    variable.NameSpace,
				DefaultValue: variable.DefaultValue,
			})
		}
	}
	return d, nil
}

func (m *OpcUaDeviceManager) DeleteDevice(device runtime.Device) (runtime.RunObject, error) {
	return &opcuaruntime.OpcUaDevice{DeviceMeta: runtime.DeviceMeta{
		ObjectMeta: runtime.ObjectMeta{ID: device.GetID(), Version: device.GetVersion()},
	}}, nil
}
