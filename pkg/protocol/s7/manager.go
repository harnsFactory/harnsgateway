package s7

import (
	s7runtime "harnsgateway/pkg/protocol/s7/runtime"
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/runtime/constant"
	"harnsgateway/pkg/utils/differenceutil"
	"harnsgateway/pkg/utils/randutil"
	"harnsgateway/pkg/utils/uuidutil"
	v1 "harnsgateway/pkg/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"strconv"
	"strings"
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

func (m *S7DeviceManager) UpdateValidation(deviceType v1.DeviceType, device runtime.Device) error {
	return nil
}

func (m *S7DeviceManager) UpdateDevice(id string, deviceType v1.DeviceType, device runtime.Device) (runtime.Device, error) {
	s7Device, ok := deviceType.(*v1.S7Device)
	if !ok {
		klog.V(2).InfoS("Unsupported device,type not OpcUa")
		return nil, constant.ErrDeviceType
	}

	copyDevice, _ := device.(*s7runtime.S7Device)
	copyDevice.DeviceMeta.PublishMeta.Topic = s7Device.Topic
	copyDevice.DeviceMeta.ObjectMeta.Name = s7Device.Name
	copyDevice.DeviceMeta.DeviceCode = s7Device.DeviceCode
	copyDevice.DeviceMeta.DeviceType = s7Device.DeviceType
	copyDevice.DeviceMeta.DeviceModel = s7Device.DeviceModel
	// todo should add enum to desc device has been updated
	// copyDevice.DeviceMeta.CollectStatus = runtime.CollectStatusToString[runtime.Stopped]

	copyDevice.CollectorCycle = s7Device.CollectorCycle
	copyDevice.VariableInterval = s7Device.VariableInterval
	copyDevice.Address.Location = s7Device.Address.Location
	copyDevice.Address.Option.Port = s7Device.Address.Option.Port
	copyDevice.Address.Option.Rack = s7Device.Address.Option.Rack
	copyDevice.Address.Option.Slot = s7Device.Address.Option.Slot

	delChars, _, _ := differenceutil.DifferenceAndIntersectionObjects(copyDevice.Variables, s7Device.Variables,
		func(value interface{}) string { return value.(*s7runtime.Variable).Name },
		func(value interface{}) string { return value.(*v1.S7Variable).Name })

	i := 0
	delCharSet := sets.NewString(delChars...)
	for _, c := range copyDevice.Variables {
		if !delCharSet.Has(c.Name) {
			copyDevice.Variables[i] = c
			i++
		} else {
			delete(copyDevice.VariablesMap, c.Name)
		}
	}
	for j := i; j < len(copyDevice.Variables); j++ {
		copyDevice.Variables[j] = nil
	}
	copyDevice.Variables = copyDevice.Variables[:i]

	// upsert
	for _, ndv := range s7Device.Variables {
		name := strings.TrimSpace(ndv.Name)
		if v, ok := copyDevice.VariablesMap[name]; ok {
			v.DataType = constant.StringToDataType[ndv.DataType]
			v.Name = ndv.Name
			v.Address = ndv.Address
			v.Rate = ndv.Rate
			v.DefaultValue = ndv.DefaultValue
			v.AccessMode = ndv.AccessMode
		} else {
			v := &s7runtime.Variable{
				DataType:     constant.StringToDataType[ndv.DataType],
				Name:         ndv.Name,
				Address:      ndv.Address,
				Rate:         ndv.Rate,
				DefaultValue: ndv.DefaultValue,
				AccessMode:   ndv.AccessMode,
			}
			copyDevice.Variables = append(copyDevice.Variables, v)
			copyDevice.VariablesMap[v.Name] = v

		}
	}

	return copyDevice, nil
}
