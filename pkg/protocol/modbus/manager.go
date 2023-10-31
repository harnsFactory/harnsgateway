package modbus

import (
	modbus "harnsgateway/pkg/protocol/modbus/runtime"
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/utils/randutil"
	"harnsgateway/pkg/utils/uuidutil"
	v1 "harnsgateway/pkg/v1"
	"strconv"
	"time"
)

type ModbusDeviceManager struct {
}

func (m *ModbusDeviceManager) CreateDevice(deviceType v1.DeviceType) (runtime.Device, error) {
	modbusDevice, ok := deviceType.(*v1.ModBusDevice)
	if !ok {
		return nil, modbus.ErrDeviceType
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
			CollectStatus: false,
		},
		CollectorCycle:   modbusDevice.CollectorCycle,
		VariableInterval: modbusDevice.VariableInterval,
		Address: &modbus.Address{
			Location: modbusDevice.Address.Location,
			Option: &modbus.Option{
				Port:     modbusDevice.Address.Option.Port,
				BaudRate: modbusDevice.Address.Option.BaudRate,
				DataBits: modbusDevice.Address.Option.DataBits,
				Parity:   runtime.StringToParity[modbusDevice.Address.Option.Parity],
				StopBits: runtime.StringToStopBits[modbusDevice.Address.Option.StopBits],
			},
		},
		Slave:           modbusDevice.Slave,
		MemoryLayout:    runtime.StringToMemoryLayout[modbusDevice.MemoryLayout],
		PositionAddress: modbusDevice.PositionAddress,
	}
	if len(modbusDevice.Variables) > 0 {
		for _, variable := range modbusDevice.Variables {
			d.Variables = append(d.Variables, &modbus.Variable{
				DataType:     runtime.StringToDataType[variable.DataType],
				Name:         variable.Name,
				Address:      *variable.Address,
				Bits:         variable.Bits,
				FunctionCode: variable.FunctionCode,
				Rate:         variable.Rate,
				Amount:       variable.Amount,
				DefaultValue: variable.DefaultValue,
			})
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
