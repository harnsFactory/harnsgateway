package model

import (
	s7 "harnsgateway/pkg/protocol/s7/runtime"
)

var _ S7Modeler = (*S71500)(nil)

var S7Modelers = map[string]S7Modeler{
	"s71500": &S71500{},
}

type S7Modeler interface {
	NewClients(address *s7.S7Address, dataFrameCount int) (*s7.Clients, error)
	GetS7DevicePDULength(address *s7.S7Address) (uint16, error)
}
