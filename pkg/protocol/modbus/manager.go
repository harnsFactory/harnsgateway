package modbus

import (
	modbus "harnsgateway/pkg/protocol/modbus/runtime"
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

type ModbusDeviceManager struct {
}

func (m *ModbusDeviceManager) CreateDevice(deviceType v1.DeviceType) (runtime.Device, error) {
	modbusDevice, ok := deviceType.(*v1.ModBusDevice)
	if !ok {
		klog.V(2).InfoS("Unsupported device,type not Modbus")
		return nil, constant.ErrDeviceType
	}

	d := &modbus.ModBusDevice{
		DeviceMeta: runtime.DeviceMeta{
			PublishMeta: runtime.PublishMeta{Topic: modbusDevice.Topic},
			ObjectMeta: runtime.ObjectMeta{
				Name:    modbusDevice.Name,
				ID:      uuidutil.UUID(),
				Version: strconv.FormatUint(randutil.Uint64n(), 10),
				ModTime: time.Now(),
			},
			DeviceCode:    modbusDevice.DeviceCode,
			DeviceType:    modbusDevice.DeviceType,
			DeviceModel:   modbusDevice.DeviceModel,
			CollectStatus: runtime.CollectStatusToString[runtime.Stopped],
		},
		CollectorCycle:   modbusDevice.CollectorCycle,
		VariableInterval: modbusDevice.VariableInterval,
		Address: &modbus.Address{
			Location: modbusDevice.Address.Location,
			Option: &modbus.Option{
				Port:     modbusDevice.Address.Option.Port,
				BaudRate: modbusDevice.Address.Option.BaudRate,
				DataBits: modbusDevice.Address.Option.DataBits,
				Parity:   constant.StringToParity[modbusDevice.Address.Option.Parity],
				StopBits: constant.StringToStopBits[modbusDevice.Address.Option.StopBits],
			},
		},
		Slave:           modbusDevice.Slave,
		MemoryLayout:    constant.StringToMemoryLayout[modbusDevice.MemoryLayout],
		PositionAddress: modbusDevice.PositionAddress,
		VariablesMap:    map[string]*modbus.Variable{},
	}
	if len(modbusDevice.Variables) > 0 {
		for _, variable := range modbusDevice.Variables {
			v := &modbus.Variable{
				DataType:     constant.StringToDataType[variable.DataType],
				Name:         variable.Name,
				Address:      *variable.Address,
				Bits:         variable.Bits,
				FunctionCode: variable.FunctionCode,
				Rate:         variable.Rate,
				Amount:       variable.Amount,
				DefaultValue: variable.DefaultValue,
				AccessMode:   variable.AccessMode,
			}
			d.Variables = append(d.Variables, v)
			d.VariablesMap[v.Name] = v
		}
	}
	return d, nil
}

func (m *ModbusDeviceManager) DeleteDevice(device runtime.Device) (runtime.Device, error) {
	return &modbus.ModBusDevice{DeviceMeta: runtime.DeviceMeta{
		ObjectMeta:  runtime.ObjectMeta{ID: device.GetID(), Version: device.GetVersion()},
		DeviceType:  device.GetDeviceType(),
		DeviceCode:  device.GetDeviceCode(),
		DeviceModel: device.GetDeviceModel(),
	}}, nil
}

func (m *ModbusDeviceManager) UpdateValidation(deviceType v1.DeviceType, device runtime.Device) error {
	return nil
}

func (m *ModbusDeviceManager) UpdateDevice(id string, deviceType v1.DeviceType, device runtime.Device) (runtime.Device, error) {
	modbusDevice, ok := deviceType.(*v1.ModBusDevice)
	if !ok {
		klog.V(2).InfoS("Unsupported device,type not Modbus")
		return nil, constant.ErrDeviceType
	}

	copyDevice, _ := device.(*modbus.ModBusDevice)
	copyDevice.DeviceMeta.PublishMeta.Topic = modbusDevice.Topic
	copyDevice.DeviceMeta.ObjectMeta.Name = modbusDevice.Name
	copyDevice.DeviceMeta.DeviceCode = modbusDevice.DeviceCode
	copyDevice.DeviceMeta.DeviceType = modbusDevice.DeviceType
	copyDevice.DeviceMeta.DeviceModel = modbusDevice.DeviceModel
	// todo should add enum to desc device has been updated
	// copyDevice.DeviceMeta.CollectStatus = runtime.CollectStatusToString[runtime.Stopped]

	copyDevice.CollectorCycle = modbusDevice.CollectorCycle
	copyDevice.VariableInterval = modbusDevice.VariableInterval
	copyDevice.Address.Location = modbusDevice.Address.Location
	copyDevice.Address.Option.Port = modbusDevice.Address.Option.Port
	copyDevice.Address.Option.BaudRate = modbusDevice.Address.Option.BaudRate
	copyDevice.Address.Option.DataBits = modbusDevice.Address.Option.DataBits
	copyDevice.Address.Option.Parity = constant.StringToParity[modbusDevice.Address.Option.Parity]
	copyDevice.Address.Option.StopBits = constant.StringToStopBits[modbusDevice.Address.Option.StopBits]

	copyDevice.Slave = modbusDevice.Slave
	copyDevice.MemoryLayout = constant.StringToMemoryLayout[modbusDevice.MemoryLayout]
	copyDevice.PositionAddress = modbusDevice.PositionAddress

	delChars, _, _ := differenceutil.DifferenceAndIntersectionObjects(copyDevice.Variables, modbusDevice.Variables,
		func(value interface{}) string { return value.(*modbus.Variable).Name },
		func(value interface{}) string { return value.(*v1.ModbusVariable).Name })

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
	for _, ndv := range modbusDevice.Variables {
		name := strings.TrimSpace(ndv.Name)
		if v, ok := copyDevice.VariablesMap[name]; ok {
			v.DataType = constant.StringToDataType[ndv.DataType]
			v.Name = ndv.Name
			v.Address = *ndv.Address
			v.Bits = ndv.Bits
			v.FunctionCode = ndv.FunctionCode
			v.Rate = ndv.Rate
			v.Amount = ndv.Amount
			v.DefaultValue = ndv.DefaultValue
			v.AccessMode = ndv.AccessMode
		} else {
			v := &modbus.Variable{
				DataType:     constant.StringToDataType[ndv.DataType],
				Name:         ndv.Name,
				Address:      *ndv.Address,
				Bits:         ndv.Bits,
				FunctionCode: ndv.FunctionCode,
				Rate:         ndv.Rate,
				Amount:       ndv.Amount,
				DefaultValue: ndv.DefaultValue,
				AccessMode:   ndv.AccessMode,
			}
			copyDevice.Variables = append(copyDevice.Variables, v)
			copyDevice.VariablesMap[v.Name] = v

		}
	}

	return copyDevice, nil
}
