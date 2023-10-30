package v1

// opcua
type OpcUaVariable struct {
	DataType     string      `json:"dataType" binding:"required"`                                   // bool、int16、float32、float64、int32、int64、uint16
	Name         string      `json:"name" binding:"required,min=1,max=64,excludesall=\u002F\u005C"` // 变量名称
	Address      interface{} `json:"address" binding:"required"`                                    // 变量地址
	NameSpace    uint16      `json:"Namespace" binding:"required"`                                  // 命名空间
	DefaultValue interface{} `json:"defaultValue,omitempty"`                                        // 默认值
}

type OpcUaDevice struct {
	DeviceMeta
	CollectorCycle   uint             `json:"collectorCycle" binding:"required"` // 采集周期
	VariableInterval uint             `json:"variableInterval,omitempty"`        // 变量间隔
	Address          *OpcAddress      `json:"address" binding:"required"`        // IP地址\串口地址
	Variables        []*OpcUaVariable `json:"variables" binding:"required,dive"` // 自定义变量
}

type OpcAddress struct {
	Location string            `json:"location"` // 地址路径
	Option   *OpcAddressOption `json:"option"`   // 地址其他参数
}

type OpcAddressOption struct {
	Port     int    `json:"port,omitempty"`     // 端口号
	Username string `json:"username,omitempty"` // 用户名
	Password string `json:"password,omitempty"` // 密码
}

// s7
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
	Address          string        `json:"address" binding:"required"`        // IP地址\串口地址
	Port             uint          `json:"port"`                              // 端口号
	Rack             uint8         `json:"rack,omitempty"`                    // 用户名
	Slot             uint8         `json:"slot,omitempty"`                    // 密码
	Variables        []*S7Variable `json:"variables" binding:"required,dive"` // 自定义变量
}
