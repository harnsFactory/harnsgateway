package opcua

import (
	opcuaruntime "harnsgateway/pkg/protocol/opcua/runtime"
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
			CollectStatus: runtime.CollectStatusToString[runtime.Stopped],
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

func (m *OpcUaDeviceManager) UpdateValidation(deviceType v1.DeviceType, device runtime.Device) error {
	return nil
}

func (m *OpcUaDeviceManager) UpdateDevice(id string, deviceType v1.DeviceType, device runtime.Device) (runtime.Device, error) {
	opcUaDevice, ok := deviceType.(*v1.OpcUaDevice)
	if !ok {
		klog.V(2).InfoS("Unsupported device,type not OpcUa")
		return nil, constant.ErrDeviceType
	}

	copyDevice, _ := device.(*opcuaruntime.OpcUaDevice)
	copyDevice.DeviceMeta.PublishMeta.Topic = opcUaDevice.Topic
	copyDevice.DeviceMeta.ObjectMeta.Name = opcUaDevice.Name
	copyDevice.DeviceMeta.DeviceCode = opcUaDevice.DeviceCode
	copyDevice.DeviceMeta.DeviceType = opcUaDevice.DeviceType
	copyDevice.DeviceMeta.DeviceModel = opcUaDevice.DeviceModel
	// todo should add enum to desc device has been updated
	// copyDevice.DeviceMeta.CollectStatus = runtime.CollectStatusToString[runtime.Stopped]

	copyDevice.CollectorCycle = opcUaDevice.CollectorCycle
	copyDevice.VariableInterval = opcUaDevice.VariableInterval
	copyDevice.Address.Location = opcUaDevice.Address.Location
	copyDevice.Address.Option.Port = opcUaDevice.Address.Option.Port
	copyDevice.Address.Option.Username = opcUaDevice.Address.Option.Username
	copyDevice.Address.Option.Password = opcUaDevice.Address.Option.Password

	delChars, _, _ := differenceutil.DifferenceAndIntersectionObjects(copyDevice.Variables, opcUaDevice.Variables,
		func(value interface{}) string { return value.(*opcuaruntime.Variable).Name },
		func(value interface{}) string { return value.(*v1.OpcUaDevice).Name })

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
	for _, ndv := range opcUaDevice.Variables {
		name := strings.TrimSpace(ndv.Name)
		if v, ok := copyDevice.VariablesMap[name]; ok {
			v.DataType = constant.StringToDataType[ndv.DataType]
			v.Name = ndv.Name
			v.Address = ndv.Address
			v.Namespace = ndv.NameSpace
			v.DefaultValue = ndv.DefaultValue
			v.AccessMode = ndv.AccessMode
		} else {
			v := &opcuaruntime.Variable{
				DataType:     constant.StringToDataType[ndv.DataType],
				Name:         ndv.Name,
				Address:      ndv.Address,
				Namespace:    ndv.NameSpace,
				DefaultValue: ndv.DefaultValue,
				AccessMode:   ndv.AccessMode,
			}
			copyDevice.Variables = append(copyDevice.Variables, v)
			copyDevice.VariablesMap[v.Name] = v

		}
	}

	return copyDevice, nil
}
