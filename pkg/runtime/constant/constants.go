package constant

import "errors"

var (
	ErrDeviceType    = errors.New("unsupported device type")
	ErrConnectDevice = errors.New("unable to connect to device")
)
