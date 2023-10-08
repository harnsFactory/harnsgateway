package v1

type DeviceType interface {
	GetDeviceType() string
}

type DeviceMeta struct {
	Name       string `json:"name" binding:"required,min=1,max=64,excludesall=\u002F\u005C"`
	DeviceCode string `json:"deviceCode" binding:"required,min=1,max=32,excludesall=\u002F\u005C"`
	DeviceType string `json:"deviceType" binding:"required,min=1,max=32,excludesall=\u002F\u005C"`
}

func (d *DeviceMeta) GetDeviceType() string {
	return d.DeviceType
}

// modbus
type ModbusVariable struct {
	DataType     string      `json:"dataType" binding:"required"`                                   // bool、int16、float32、float64、int32、int64、uint16
	Name         string      `json:"name" binding:"required,min=1,max=64,excludesall=\u002F\u005C"` // 变量名称
	Address      *uint       `json:"address" binding:"required,number,gte=0"`                       // 变量地址
	Bits         uint16      `json:"bits" binding:"gte=0,lte=7"`                                    // 位
	FunctionCode uint16      `json:"functionCode" binding:"required,gte=1,lte=4"`                   // 功能码 1、2、3、4
	Rate         float64     `json:"rate,omitempty"`                                                // 比率
	Amount       uint        `json:"amount,omitempty"`                                              // 数量
	DefaultValue interface{} `json:"defaultValue,omitempty"`                                        // 默认值
}

type ModBusDevice struct {
	DeviceMeta
	CollectorCycle   uint              `json:"collectorCycle" binding:"required"`                         // 采集周期
	VariableInterval uint              `json:"variableInterval,omitempty"`                                // 变量间隔
	Address          string            `json:"address" binding:"required"`                                // IP地址\串口地址
	Port             uint              `json:"port" binding:"required"`                                   // 端口号
	Slave            uint              `json:"slave" binding:"required"`                                  // 下位机号
	MemoryLayout     string            `json:"memoryLayout" binding:"required,oneof=ABCD BADC CDAB DCBA"` // 内存布局 DCBA CDAB BADC ABCD
	PositionAddress  uint              `json:"positionAddress,omitempty"`                                 // 起始地址
	Variables        []*ModbusVariable `json:"variables" binding:"required,dive"`                         // 自定义变量
}

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
	Address          string           `json:"address" binding:"required"`        // IP地址\串口地址
	Port             uint             `json:"port"`                              // 端口号
	Username         string           `json:"username,omitempty"`                // 用户名
	Password         string           `json:"password,omitempty"`                // 密码
	Variables        []*OpcUaVariable `json:"variables" binding:"required,dive"` // 自定义变量
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

// modbus-rtu
type ModBusRtuDevice struct {
	DeviceMeta
	CollectorCycle   uint              `json:"collectorCycle" binding:"required"`                         // 采集周期
	VariableInterval uint              `json:"variableInterval,omitempty"`                                // 变量间隔
	Address          string            `json:"address" binding:"required"`                                // IP地址\串口地址
	Slave            uint              `json:"slave" binding:"required"`                                  // 下位机号
	MemoryLayout     string            `json:"memoryLayout" binding:"required,oneof=ABCD BADC CDAB DCBA"` // 内存布局 DCBA CDAB BADC ABCD
	PositionAddress  uint              `json:"positionAddress,omitempty"`                                 // 起始地址
	BaudRate         int               `json:"baudRate,omitempty"`                                        // 波特率 (1200,2400,4800,9600,19200,38400,57600,115200)
	DataBits         int               `json:"dataBits,omitempty"`                                        // 数据位 (must be 5, 6, 7 or 8)
	Parity           string            `json:"parity,omitempty"`                                          // 校验位 (无校验,奇校验,偶校验)
	StopBits         string            `json:"stopBits,omitempty"`                                        // 停止位 (must be 1, 1.5, 2)
	Variables        []*ModbusVariable `json:"variables" binding:"required,dive"`                         // 自定义变量
}
