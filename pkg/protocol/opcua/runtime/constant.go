package runtime

import "errors"

var ErrBadConn = errors.New("Tcp bad connection\n")
var ErrTcpClosed = errors.New("Tcp closed\n")
var ErrManyRetry = errors.New("Tcp connect retry more than three times\n")
var ErrDeviceType = errors.New("Error device type\n")
var ErrConnectOpuServer = errors.New("Can not Connect to opu server\n")
