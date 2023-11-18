package runtime

import "errors"

var ErrBadConn = errors.New("Opc server bad connection\n")
var ErrServerBadResp = errors.New("Opc server bad response\n")
var ErrTcpClosed = errors.New("Opc server closed\n")
var ErrManyRetry = errors.New("Opc server connect retry more than three times\n")
var ErrConnectOpuServer = errors.New("Can not Connect to opu server\n")
