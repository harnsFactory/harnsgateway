package s7

import (
	s7runtime "harnsgateway/pkg/protocol/s7/runtime"
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/runtime/constant"
	"harnsgateway/pkg/utils/randutil"
	"harnsgateway/pkg/utils/uuidutil"
	v1 "harnsgateway/pkg/v1"
	"k8s.io/klog/v2"
	"strconv"
	"time"
)

type S7DeviceManager struct {
}

func (m *S7DeviceManager) CreateDevice(deviceType v1.DeviceType) (runtime.Device, error) {
	s7Device, ok := deviceType.(*v1.S7Device)
	if !ok {
		klog.V(2).InfoS("Unsupported device,type not S7")
		return nil, constant.ErrDeviceType
	}

	d := &s7runtime.S7Device{
		DeviceMeta: runtime.DeviceMeta{
			PublishMeta: runtime.PublishMeta{Topic: s7Device.Topic},
			ObjectMeta: runtime.ObjectMeta{
				Name:    s7Device.Name,
				ID:      uuidutil.UUID(),
				Version: strconv.FormatUint(randutil.Uint64n(), 10),
				ModTime: time.Now(),
			},
			DeviceCode:    s7Device.DeviceCode,
			DeviceType:    s7Device.DeviceType,
			DeviceModel:   s7Device.DeviceModel,
			CollectStatus: runtime.CollectStatusToString[runtime.Stopped],
		},
		CollectorCycle:   s7Device.CollectorCycle,
		VariableInterval: s7Device.VariableInterval,
		Address: &s7runtime.S7Address{
			Location: s7Device.Address.Location,
			Option: &s7runtime.S7AddressOption{
				Port: s7Device.Address.Option.Port,
				Rack: s7Device.Address.Option.Rack,
				Slot: s7Device.Address.Option.Slot,
			},
		},
		VariablesMap: map[string]*s7runtime.Variable{},
	}
	if len(s7Device.Variables) > 0 {
		for _, variable := range s7Device.Variables {
			v := &s7runtime.Variable{
				DataType:     constant.StringToDataType[variable.DataType],
				Name:         variable.Name,
				Address:      variable.Address,
				Rate:         variable.Rate,
				DefaultValue: variable.DefaultValue,
				AccessMode:   variable.AccessMode,
			}
			d.Variables = append(d.Variables, v)
			d.VariablesMap[v.Name] = v
		}
	}
	return d, nil
}

func (m *S7DeviceManager) DeleteDevice(device runtime.Device) (runtime.Device, error) {
	return &s7runtime.S7Device{DeviceMeta: runtime.DeviceMeta{
		ObjectMeta:  runtime.ObjectMeta{ID: device.GetID(), Version: device.GetVersion()},
		DeviceType:  device.GetDeviceType(),
		DeviceCode:  device.GetDeviceCode(),
		DeviceModel: device.GetDeviceModel(),
	}}, nil
}
