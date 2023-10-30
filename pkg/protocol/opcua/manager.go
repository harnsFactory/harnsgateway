package opcua

import (
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

func (m *OpcUaDeviceManager) CreateDevice(deviceType v1.DeviceType) (runtime.Device, error) {
	opcUaDevice, ok := deviceType.(*v1.OpcUaDevice)
	if !ok {
		return nil, opcuaruntime.ErrDeviceType
	}

	d := &opcuaruntime.OpcUaDevice{
		DeviceMeta: runtime.DeviceMeta{
			ObjectMeta: runtime.ObjectMeta{
				Name:    opcUaDevice.Name,
				ID:      uuidutil.UUID(),
				Version: strconv.FormatUint(randutil.Uint64n(), 10),
				ModTime: time.Now(),
			},
			DeviceCode:    opcUaDevice.DeviceCode,
			DeviceType:    opcUaDevice.DeviceType,
			DeviceModel:   opcUaDevice.DeviceModel,
			CollectStatus: false,
		},
		CollectorCycle:   opcUaDevice.CollectorCycle,
		VariableInterval: opcUaDevice.VariableInterval,
		Address: &opcuaruntime.Address{
			Location: opcUaDevice.Address.Location,
			Option: &opcuaruntime.Option{
				Port:     opcUaDevice.Address.Option.Port,
				Username: opcUaDevice.Address.Option.Username,
				Password: opcUaDevice.Address.Option.Password,
			},
		},
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

func (m *OpcUaDeviceManager) DeleteDevice(device runtime.Device) (runtime.Device, error) {
	return &opcuaruntime.OpcUaDevice{DeviceMeta: runtime.DeviceMeta{
		ObjectMeta:  runtime.ObjectMeta{ID: device.GetID(), Version: device.GetVersion()},
		DeviceType:  device.GetDeviceType(),
		DeviceCode:  device.GetDeviceCode(),
		DeviceModel: device.GetDeviceModel(),
	}}, nil
}
