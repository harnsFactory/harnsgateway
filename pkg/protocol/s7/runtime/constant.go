package runtime

import "errors"

var ErrBadConn = errors.New("Tcp bad connection\n")
var ErrTcpClosed = errors.New("Tcp closed\n")
var ErrManyRetry = errors.New("Tcp connect retry more than three times\n")
var ErrDeviceType = errors.New("Error device type\n")
var ErrConnectS7DeviceCotpMessage = errors.New("Error connect s7 passed cotp message\n")
var ErrConnectS7DeviceS7COMMMessage = errors.New("Error connect s7 passed s7comm message\n")
var ErrMessageDataLengthNotEnough = errors.New("S7 message data length not enough\n")
var ErrMessageS7Response = errors.New("S7 message response error\n")

type S7StoreAddress int8
type AddressType int8

const (
	I S7StoreAddress = iota
	Q
	M
	DB
	// todo
)

var StoreAddressToString = map[S7StoreAddress]string{
	I:  "I",
	Q:  "Q",
	M:  "M",
	DB: "DB",
}

var StringToStoreAddress = map[string]S7StoreAddress{
	"I":  I,
	"Q":  Q,
	"M":  M,
	"DB": DB,
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
