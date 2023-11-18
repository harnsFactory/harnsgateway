package v1

import "harnsgateway/pkg/runtime/constant"

// modbus
type ModbusVariable struct {
	DataType     string              `json:"dataType" binding:"required"`                                   // bool、int16、float32、float64、int32、int64、uint16
	Name         string              `json:"name" binding:"required,min=1,max=64,excludesall=\u002F\u005C"` // 变量名称
	Address      *uint               `json:"address" binding:"required,number,gte=0"`                       // 变量地址
	Bits         uint8               `json:"bits" binding:"gte=0,lte=7"`                                    // 位
	FunctionCode uint8               `json:"functionCode" binding:"required,gte=1,lte=4"`                   // 功能码 1、2、3、4
	Rate         float64             `json:"rate,omitempty"`                                                // 比率
	Amount       uint                `json:"amount,omitempty"`                                              // 数量
	DefaultValue interface{}         `json:"defaultValue,omitempty"`                                        // 默认值
	AccessMode   constant.AccessMode `json:"accessMode" binding:"required"`                                 // 读写属性
}

type ModBusDevice struct {
	DeviceMeta
	CollectorCycle   uint              `json:"collectorCycle" binding:"required"`                         // 采集周期
	VariableInterval uint              `json:"variableInterval,omitempty"`                                // 变量间隔
	Address          *ModbusAddress    `json:"address" binding:"required"`                                // IP地址\串口地址
	Slave            uint              `json:"slave" binding:"required"`                                  // 下位机号
	MemoryLayout     string            `json:"memoryLayout" binding:"required,oneof=ABCD BADC CDAB DCBA"` // 内存布局 DCBA CDAB BADC ABCD
	PositionAddress  uint              `json:"positionAddress,omitempty"`                                 // 起始地址
	Variables        []*ModbusVariable `json:"variables" binding:"required,dive"`                         // 自定义变量
}

type ModbusAddress struct {
	Location string               `json:"location"` // 地址路径
	Option   *ModbusAddressOption `json:"option"`   // 地址其他参数
}

type ModbusAddressOption struct {
	Port     int    `json:"port,omitempty"`     // 端口号
	BaudRate int    `json:"baudRate,omitempty"` // 波特率
	DataBits int    `json:"dataBits,omitempty"` // 数据位
	Parity   string `json:"parity,omitempty"`   // 校验位
	StopBits string `json:"stopBits,omitempty"` // 停止位
}
