package modbusrtu

import (
	modbusrturuntime "harnsgateway/pkg/protocol/modbusrtu/runtime"
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/utils/randutil"
	"harnsgateway/pkg/utils/uuidutil"
	v1 "harnsgateway/pkg/v1"
	"strconv"
	"time"
)

type ModbusRtuDeviceManager struct {
}

func (m *ModbusRtuDeviceManager) CreateDevice(deviceType v1.DeviceType) (runtime.Device, error) {
	modbusRtuDevice, ok := deviceType.(*v1.ModBusRtuDevice)
	if !ok {
		return nil, modbusrturuntime.ErrDeviceType
	}

	d := &modbusrturuntime.ModBusRtuDevice{
		DeviceMeta: runtime.DeviceMeta{
			ObjectMeta: runtime.ObjectMeta{
				Name:    modbusRtuDevice.Name,
				ID:      uuidutil.UUID(),
				Version: strconv.FormatUint(randutil.Uint64n(), 10),
				ModTime: time.Now(),
			},
			DeviceCode:    modbusRtuDevice.DeviceCode,
			DeviceType:    modbusRtuDevice.DeviceType,
			CollectStatus: false,
		},
		CollectorCycle:   modbusRtuDevice.CollectorCycle,
		VariableInterval: modbusRtuDevice.VariableInterval,
		Address:          modbusRtuDevice.Address,
		BaudRate:         modbusRtuDevice.BaudRate,
		DataBits:         modbusRtuDevice.DataBits,
		Parity:           runtime.StringToParity[modbusRtuDevice.Parity],
		StopBits:         runtime.StringToStopBits[modbusRtuDevice.StopBits],
		Slave:            modbusRtuDevice.Slave,
		MemoryLayout:     runtime.StringToMemoryLayout[modbusRtuDevice.MemoryLayout],
		PositionAddress:  modbusRtuDevice.PositionAddress,
	}
	if len(modbusRtuDevice.Variables) > 0 {
		for _, variable := range modbusRtuDevice.Variables {
			d.Variables = append(d.Variables, &modbusrturuntime.Variable{
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

func (m *ModbusRtuDeviceManager) DeleteDevice(device runtime.Device) (runtime.Device, error) {
	return &modbusrturuntime.ModBusRtuDevice{DeviceMeta: runtime.DeviceMeta{
		ObjectMeta: runtime.ObjectMeta{ID: device.GetID(), Version: device.GetVersion()},
		DeviceType: device.GetDeviceType(),
		DeviceCode: device.GetDeviceCode(),
	}}, nil
}
