package s7

import (
	s7runtime "harnsgateway/pkg/protocol/s7/runtime"
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/utils/randutil"
	"harnsgateway/pkg/utils/uuidutil"
	v1 "harnsgateway/pkg/v1"
	"strconv"
	"time"
)

type S7DeviceManager struct {
}

func (m *S7DeviceManager) CreateDevice(deviceType v1.DeviceType) (runtime.Device, error) {
	s7Device, ok := deviceType.(*v1.S7Device)
	if !ok {
		return nil, s7runtime.ErrDeviceType
	}

	d := &s7runtime.S7Device{
		DeviceMeta: runtime.DeviceMeta{
			ObjectMeta: runtime.ObjectMeta{
				Name:    s7Device.Name,
				ID:      uuidutil.UUID(),
				Version: strconv.FormatUint(randutil.Uint64n(), 10),
				ModTime: time.Now(),
			},
			DeviceCode:    s7Device.DeviceCode,
			DeviceType:    s7Device.DeviceType,
			CollectStatus: false,
		},
		CollectorCycle:   s7Device.CollectorCycle,
		VariableInterval: s7Device.VariableInterval,
		Address:          s7Device.Address,
		Port:             s7Device.Port,
		Rack:             s7Device.Rack,
		Slot:             s7Device.Slot,
	}
	if len(s7Device.Variables) > 0 {
		for _, variable := range s7Device.Variables {
			d.Variables = append(d.Variables, &s7runtime.Variable{
				DataType:     runtime.StringToDataType[variable.DataType],
				Name:         variable.Name,
				Address:      variable.Address,
				Rate:         variable.Rate,
				DefaultValue: variable.DefaultValue,
			})
		}
	}
	return d, nil
}

func (m *S7DeviceManager) DeleteDevice(device runtime.Device) (runtime.Device, error) {
	return &s7runtime.S7Device{DeviceMeta: runtime.DeviceMeta{
		ObjectMeta: runtime.ObjectMeta{ID: device.GetID(), Version: device.GetVersion()},
		DeviceType: device.GetDeviceType(),
		DeviceCode: device.GetDeviceCode(),
	}}, nil
}
