package runtime

import "errors"

var ErrBadConn = errors.New("S7 bad connection\n")
var ErrCommandFailed = errors.New("Error command s7\n")
var ErrServerBadResp = errors.New("S7 server bad response\n")
var ErrTcpClosed = errors.New("S7 closed\n")
var ErrManyRetry = errors.New("S7 connect retry more than three times\n")
var ErrDeviceType = errors.New("Error device type\n")
var ErrConnectS7DeviceCotpMessage = errors.New("Error connect s7 passed cotp message\n")
var ErrConnectS7DeviceS7COMMMessage = errors.New("Error connect s7 passed s7comm message\n")
var ErrMessageDataLengthNotEnough = errors.New("S7 message data length not enough\n")
var ErrMessageS7Response = errors.New("S7 message response error\n")

type S7StoreArea int8
type AddressType int8

const (
	I S7StoreArea = iota
	Q
	M
	DB
)

var StoreAddressToString = map[S7StoreArea]string{
	I:  "I",
	Q:  "Q",
	M:  "M",
	DB: "DB",
}

var StringToStoreAddress = map[string]S7StoreArea{
	"I":  I,
	"Q":  Q,
	"M":  M,
	"DB": DB,
}

var StoreAreaCode = map[S7StoreArea]uint8{
	I:  129,
	Q:  130,
	M:  131,
	DB: 132,
}

var StoreAreaTransportSize = map[S7StoreArea]uint8{
	I:  2,
	Q:  2,
	M:  2,
	DB: 4,
	// DB: 6,
}

const (
	String AddressType = iota
	Bool
	Byte
	Word
	DWord
)

var AddressTypeToString = map[AddressType]string{
	String: "STRING",
	Bool:   "X",
	Byte:   "B",
	Word:   "W",
	DWord:  "D",
}

var StringToAddressType = map[string]AddressType{
	"STRING": String,
	"X":      Bool,
	"B":      Byte,
	"W":      Word,
	"D":      DWord,
}
