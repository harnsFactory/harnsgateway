package modbus

import (
	modbus "harnsgateway/pkg/protocol/modbus/runtime"
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/runtime/constant"
	"harnsgateway/pkg/utils/randutil"
	"harnsgateway/pkg/utils/uuidutil"
	v1 "harnsgateway/pkg/v1"
	"k8s.io/klog/v2"
	"strconv"
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
