package constant

import "errors"

var (
	ErrDeviceType          = errors.New("unsupported device type")
	ErrConnectDevice       = errors.New("unable to connect to device")
	ErrDeviceServerClosed  = errors.New("device server closed")
	ErrDeviceEmptyVariable = errors.New("device variable emptied")
)
