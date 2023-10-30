package model

import (
	opc "harnsgateway/pkg/protocol/opcua/runtime"
)

var _ OpcUaModeler = (*OpcUa)(nil)

var OpcUaModelers = map[string]OpcUaModeler{
	"opcUa": &OpcUa{},
}

type OpcUaModeler interface {
	NewClients(address *opc.Address, dataFrameCount int) (*opc.Clients, error)
}
