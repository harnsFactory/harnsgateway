package runtime

import (
	"harnsgateway/pkg/runtime"
)

type Variable struct {
	DataType     runtime.DataType `json:"dataType"`               // bool、int16、float32、float64、int32、int64、uint16
	Name         string           `json:"name"`                   // 变量名称
	Address      interface{}      `json:"address"`                // 变量地址
	Namespace    uint16           `json:"Namespace"`              // namespace
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

type OpcUaDevice struct {
	runtime.DeviceMeta
	CollectorCycle   uint        `json:"collectorCycle"`   // 采集周期
	VariableInterval uint        `json:"variableInterval"` // 变量间隔
	Address          string      `json:"address"`          // IP地址\串口地址
	Port             uint        `json:"port"`             // 端口号
	Username         string      `json:"username,omitempty"`
	Password         string      `json:"password,omitempty"`
	Variables        []*Variable `json:"variables" binding:"required,dive"` // 自定义变量
}

type VariableSlice []*Variable

type ParseVariableResult struct {
	VariableSlice VariableSlice
	Err           []error
}
