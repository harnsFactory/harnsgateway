package runtime

import "errors"

var ErrBadConn = errors.New("Tcp bad connection\n")
var ErrTcpClosed = errors.New("Tcp closed\n")
var ErrMessageTransaction = errors.New("Tcp message transaction not match\n")
var ErrMessageDataLengthNotEnough = errors.New("Tcp message data length not enough\n")
var ErrMessageFunctionCodeError = errors.New("Tcp message function code error\n")
var ErrManyRetry = errors.New("Tcp connect retry more than three times\n")
var ErrDeviceType = errors.New("Error device type\n")
