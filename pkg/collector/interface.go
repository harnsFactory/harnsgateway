package collector

import (
	"harnsgateway/pkg/runtime"
	v1 "harnsgateway/pkg/v1"
)

type DeviceManager interface {
	CreateDevice(deviceType v1.DeviceType) (runtime.Device, error)
	DeleteDevice(device runtime.Device) (runtime.Device, error)
}
