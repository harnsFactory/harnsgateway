package runtime

import (
	"harnsgateway/pkg/runtime"
)

type Variable struct {
	DataType     runtime.DataType `json:"dataType"`               // bool、int16、float32、float64、int32、int64、uint16
	Name         string           `json:"name"`                   // 变量名称
	Address      uint             `json:"address"`                // 变量地址
	Bits         uint8            `json:"bits"`                   // 位
	FunctionCode uint8            `json:"functionCode"`           // 功能码 1、2、3、4
	Rate         float64          `json:"rate"`                   // 比率
	Amount       uint             `json:"amount"`                 // 数量
	DefaultValue interface{}      `json:"defaultValue,omitempty"` // 默认值
	Value        interface{}      `json:"value,omitempty"`        // 值
}

func (v *Variable) SetValue(value interface{}) {
	v.Value = value
}

func (v *Variable) GetValue() interface{} {
	return v.Value
}

func (v *Variable) GetVariableName() string {
	return v.Name
}

func (v *Variable) SetVariableName(name string) {
	v.Name = name
}

type ModBusDevice struct {
	runtime.DeviceMeta
	CollectorCycle   uint                 `json:"collectorCycle"`                    // 采集周期
	VariableInterval uint                 `json:"variableInterval"`                  // 变量间隔
	Address          string               `json:"address"`                           // IP地址\串口地址
	Port             uint                 `json:"port"`                              // 端口号
	Slave            uint                 `json:"slave"`                             // 下位机号
	MemoryLayout     runtime.MemoryLayout `json:"memoryLayout"`                      // 内存布局 DCBA CDAB BADC ABCD
	PositionAddress  uint                 `json:"positionAddress"`                   // 起始地址
	Variables        []*Variable          `json:"variables" binding:"required,dive"` // 自定义变量
}

type VariableSlice []*Variable

type ParseVariableResult struct {
	VariableSlice VariableSlice
	Err           []error
}

func (vs VariableSlice) Len() int {
	return len(vs)
}

func (vs VariableSlice) Less(i, j int) bool {
	return vs[i].Address < vs[j].Address
}

func (vs VariableSlice) Swap(i, j int) {
	vs[i], vs[j] = vs[j], vs[i]
}
