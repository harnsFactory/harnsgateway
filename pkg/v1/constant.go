package v1

var DeviceTypeMap = map[string]func() DeviceType{
	"modbusTcp": func() DeviceType { return &ModBusDevice{} },
	"opcUa":     func() DeviceType { return &OpcUaDevice{} },
	"s71500":    func() DeviceType { return &S7Device{} },
	"modbusRtu": func() DeviceType { return &ModBusRtuDevice{} },
}
