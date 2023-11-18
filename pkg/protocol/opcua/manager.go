package opcua

import (
	opcuaruntime "harnsgateway/pkg/protocol/opcua/runtime"
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/runtime/constant"
	"harnsgateway/pkg/utils/randutil"
	"harnsgateway/pkg/utils/uuidutil"
	v1 "harnsgateway/pkg/v1"
	"k8s.io/klog/v2"
	"strconv"
	"time"
)

type OpcUaDeviceManager struct {
}

func (m *OpcUaDeviceManager) CreateDevice(deviceType v1.DeviceType) (runtime.Device, error) {
	opcUaDevice, ok := deviceType.(*v1.OpcUaDevice)
	if !ok {
		klog.V(2).InfoS("Unsupported device,type not OpcUa")
		return nil, constant.ErrDeviceType
	}

	d := &opcuaruntime.OpcUaDevice{
		DeviceMeta: runtime.DeviceMeta{
			PublishMeta: runtime.PublishMeta{Topic: opcUaDevice.Topic},
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
				DataType:     constant.StringToDataType[variable.DataType],
				Name:         variable.Name,
				Address:      variable.Address,
				Namespace:    variable.NameSpace,
				DefaultValue: variable.DefaultValue,
				AccessMode:   variable.AccessMode,
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
