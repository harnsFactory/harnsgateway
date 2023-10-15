package runtime

import (
	"errors"
	"go.bug.st/serial"
	"harnsgateway/pkg/runtime"
)

var ErrBadConn = errors.New("Rtu bad connection\n")
var ErrServerBadResp = errors.New("Rtu server bad response\n")
var ErrSerialPortClosed = errors.New("Serial port closed\n")
var ErrModbusRtuDataLengthNotEnough = errors.New("Modbus rtu message data length not enough\n")
var ErrCRC16Error = errors.New("Rtu message crc16 error\n")
var ErrMessageFunctionCodeError = errors.New("Rtu message function code error\n")
var ErrManyRetry = errors.New("Rtu connect retry more than three times\n")
var ErrDeviceType = errors.New("Error device type\n")

var StopBitsToStopBits = map[runtime.StopBits]serial.StopBits{
	runtime.OneStopBit:           serial.OneStopBit,
	runtime.OnePointFiveStopBits: serial.OnePointFiveStopBits,
	runtime.TwoStopBits:          serial.TwoStopBits,
}

var ParityToParity = map[runtime.Parity]serial.Parity{
	runtime.NoParity:   serial.NoParity,
	runtime.OddParity:  serial.OddParity,
	runtime.EvenParity: serial.EvenParity,
}
