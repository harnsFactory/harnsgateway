package runtime

import "errors"

var ErrManyRetry = errors.New("Opc server connect retry more than three times\n")
var ErrConnectOpuServer = errors.New("Can not Connect to opu server\n")
