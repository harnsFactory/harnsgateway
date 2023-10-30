package v1

type S7Variable struct {
	DataType     string      `json:"dataType" binding:"required"`                                   // bool、int16、float32、float64、int32、int64、uint16
	Name         string      `json:"name" binding:"required,min=1,max=64,excludesall=\u002F\u005C"` // 变量名称
	Address      string      `json:"address" binding:"required"`                                    // 变量地址
	Rate         float64     `json:"rate,omitempty"`
	DefaultValue interface{} `json:"defaultValue,omitempty"` // 默认值
}

type S7Device struct {
	DeviceMeta
	CollectorCycle   uint          `json:"collectorCycle" binding:"required"` // 采集周期
	VariableInterval uint          `json:"variableInterval,omitempty"`        // 变量间隔
	Address          *S7Address    `json:"address" binding:"required"`        // IP地址\串口地址
	Variables        []*S7Variable `json:"variables" binding:"required,dive"` // 自定义变量
}

type S7Address struct {
	Location string           `json:"location"` // 地址路径
	Option   *S7AddressOption `json:"option"`   // 地址其他参数
}

type S7AddressOption struct {
	Port uint  `json:"port"`           // 端口号
	Rack uint8 `json:"rack,omitempty"` // rack
	Slot uint8 `json:"slot,omitempty"` // slot
}
