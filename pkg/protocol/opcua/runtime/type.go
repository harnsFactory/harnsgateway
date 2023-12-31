package runtime

import (
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/runtime/constant"
)

var _ runtime.Device = (*OpcUaDevice)(nil)
var _ runtime.VariableValue = (*Variable)(nil)

type Variable struct {
	DataType     constant.DataType   `json:"dataType"`               // bool、int16、float32、float64、int32、int64、uint16
	Name         string              `json:"name"`                   // 变量名称
	Address      interface{}         `json:"address"`                // 变量地址
	Namespace    uint16              `json:"Namespace"`              // namespace
	DefaultValue interface{}         `json:"defaultValue,omitempty"` // 默认值
	Value        interface{}         `json:"value,omitempty"`        // 值
	AccessMode   constant.AccessMode `json:"accessMode"`             // 读写属性
}

func (v *Variable) GetVariableAccessMode() constant.AccessMode {
	return v.AccessMode
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
	CollectorCycle   uint                 `json:"collectorCycle"`                    // 采集周期
	VariableInterval uint                 `json:"variableInterval"`                  // 变量间隔
	Address          *Address             `json:"address"`                           // IP地址
	Variables        []*Variable          `json:"variables" binding:"required,dive"` // 自定义变量
	VariablesMap     map[string]*Variable `json:"-"`
}

func (o *OpcUaDevice) IndexDevice() {
	o.VariablesMap = make(map[string]*Variable)
	for _, variable := range o.Variables {
		o.VariablesMap[variable.Name] = variable
	}
}

func (o *OpcUaDevice) GetVariable(key string) (rv runtime.VariableValue, exist bool) {
	if v, isExist := o.VariablesMap[key]; isExist {
		rv = v
		exist = isExist
	}
	return
}

type Address struct {
	Location string  `json:"location"` // 地址路径
	Option   *Option `json:"option"`   // 地址其他参数
}

type Option struct {
	Port     int    `json:"port,omitempty"`     // 端口号
	Username string `json:"username,omitempty"` // 用户名
	Password string `json:"password,omitempty"` // 密码
}

type VariableSlice []*Variable

type ParseVariableResult struct {
	VariableSlice VariableSlice
	Err           []error
}
