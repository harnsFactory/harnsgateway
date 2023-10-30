package runtime

import (
	"harnsgateway/pkg/runtime"
	"k8s.io/klog/v2"
	"strconv"
	"strings"
)

type Variable struct {
	DataType     runtime.DataType `json:"dataType"`               // bool、int16、float32、float64、int32、int64、uint16
	Name         string           `json:"name"`                   // 变量名称
	Address      string           `json:"address"`                // 变量地址
	Rate         float64          `json:"rate,omitempty"`         // 比率
	DefaultValue interface{}      `json:"defaultValue,omitempty"` // 默认值
	Value        interface{}      `json:"value,omitempty"`        // 值
}

func (v *Variable) DataRequestLength(area S7StoreArea) uint16 {
	switch area {
	// 以byte形式读取 发送的一个item占12个字节    返回的一个item至少占5个字节
	case I, Q, M:
		switch v.DataType {
		case runtime.BOOL:
			return uint16(1)
		case runtime.STRING:
			return uint16(1)
		case runtime.UINT16:
			return uint16(2)
		case runtime.INT16:
			return uint16(2)
		case runtime.INT32:
			return uint16(4)
		case runtime.FLOAT32:
			return uint16(4)
		case runtime.INT64:
			return uint16(8)
		case runtime.FLOAT64:
			return uint16(8)
		default:
			return uint16(1)
		}
	// 以word形式读取 发送的一个item占12个字节    返回的一个item至少占8个字节
	case DB:
		switch v.DataType {
		case runtime.BOOL:
			return uint16(1)
		case runtime.STRING:
			return uint16(1)
		case runtime.UINT16:
			return uint16(1)
		case runtime.INT16:
			return uint16(1)
		case runtime.INT32:
			return uint16(2)
		case runtime.FLOAT32:
			return uint16(2)
		case runtime.INT64:
			return uint16(4)
		case runtime.FLOAT64:
			return uint16(4)
		default:
			return uint16(1)
		}
	default:
		return uint16(1)
	}
}

func (v *Variable) DataResponseLength(area S7StoreArea) uint16 {
	switch area {
	case I, Q, M:
		switch v.DataType {
		case runtime.BOOL:
			return uint16(1)
		case runtime.STRING:
			return uint16(1)
		case runtime.UINT16:
			return uint16(2)
		case runtime.INT16:
			return uint16(2)
		case runtime.INT32:
			return uint16(4)
		case runtime.FLOAT32:
			return uint16(4)
		case runtime.INT64:
			return uint16(8)
		case runtime.FLOAT64:
			return uint16(8)
		default:
			return uint16(1)
		}
	case DB:
		switch v.DataType {
		case runtime.BOOL:
			return uint16(2)
		case runtime.STRING:
			return uint16(2)
		case runtime.UINT16:
			return uint16(2)
		case runtime.INT16:
			return uint16(2)
		case runtime.INT32:
			return uint16(4)
		case runtime.FLOAT32:
			return uint16(4)
		case runtime.INT64:
			return uint16(8)
		case runtime.FLOAT64:
			return uint16(8)
		default:
			return uint16(2)
		}
	default:
		return uint16(1)
	}
}

func (v *Variable) Zone() S7StoreArea {
	if strings.HasPrefix(v.Address, "I") {
		return I
	} else if strings.HasPrefix(v.Address, "M") {
		return M
	} else if strings.HasPrefix(v.Address, "DB") {
		return DB
	} else if strings.HasPrefix(v.Address, "Q") {
		return Q
	}
	return DB
}

func (v *Variable) BlockSize() uint {
	if strings.HasPrefix(v.Address, "I") {
		return 0
	} else if strings.HasPrefix(v.Address, "M") {
		return 0
	} else if strings.HasPrefix(v.Address, "DB") {
		index := strings.Index(v.Address, ".")
		blockSizeString := v.Address[2:index]
		bs, err := strconv.Atoi(blockSizeString)
		if err != nil {
			klog.V(2).InfoS("Failed to read s7 variable address", "variableName", v.Name)
		}
		return uint(bs)
	} else if strings.HasPrefix(v.Address, "Q") {
		return 0
	}
	return 0
}

func (v *Variable) ParseVariableAddress() (zone S7StoreArea, areaSize uint, address uint32, bit uint8) {
	zone = v.Zone()
	switch zone {
	case I, Q, M:
		areaSize = 0
		byteAddress := v.Address[1:]
		if !v.shortening(byteAddress) {
			byteAddress = byteAddress[1:]
		}

		index := strings.LastIndex(byteAddress, ".")
		if index == -1 {
			startAddressString := byteAddress[:]
			i, err := strconv.Atoi(startAddressString)
			if err != nil {
				klog.V(2).InfoS("Failed to read s7 variable address", "variableName", v.Name)
			}
			address = uint32(i)
			bit = uint8(0)
		} else {
			startAddressString := byteAddress[:index]
			i, err := strconv.Atoi(startAddressString)
			if err != nil {
				klog.V(2).InfoS("Failed to read s7 variable address", "variableName", v.Name)
			}
			address = uint32(i)
			bitAddressString := byteAddress[index+1:]
			j, err := strconv.Atoi(bitAddressString)
			if err != nil {
				klog.V(2).InfoS("Failed to read s7 variable address", "variableName", v.Name)
			}
			bit = uint8(j)
		}
	case DB:
		// DB7.DBX0.5
		// DB1.DBD0
		// DB7.DBX22.0
		// DB1.DBW0 //地址为0，类型为整数
		// DB1.STRING2.18 //地址为2，字符长度18 类型为字符串
		// DB1.B22 //地址为22，类型为字节型
		// DB1.DBD24 //地址为24，类型为实数
		// DB1.DBW28 //地址为28，类型为整数
		// DB1.DBW28 //地址为28，类型为整数
		// DB1.DBX29.0 //地址为29.0，类型为布尔
		index := strings.Index(v.Address, ".")
		blockSizeString := v.Address[2:index]
		bs, err := strconv.Atoi(blockSizeString)
		if err != nil {
			klog.V(2).InfoS("Failed to read s7 variable address", "variableName", v.Name)
		}
		areaSize = uint(bs)
		byteAddress := v.Address[index+1:]
		if !v.shortening(byteAddress) {
			byteAddress = byteAddress[3:]
		}

		lastIndex := strings.LastIndex(byteAddress, ".")
		if lastIndex == -1 {
			startAddressString := byteAddress[:]
			i, err := strconv.Atoi(startAddressString)
			if err != nil {
				klog.V(2).InfoS("Failed to read s7 variable address", "variableName", v.Name)
			}
			address = uint32(i)
			bit = uint8(0)
		} else {
			startAddressString := byteAddress[:lastIndex]
			i, err := strconv.Atoi(startAddressString)
			if err != nil {
				klog.V(2).InfoS("Failed to read s7 variable address", "variableName", v.Name)
			}
			address = uint32(i)
			bitAddressString := byteAddress[lastIndex+1:]
			j, err := strconv.Atoi(bitAddressString)
			if err != nil {
				klog.V(2).InfoS("Failed to read s7 variable address", "variableName", v.Name)
			}
			bit = uint8(j)
		}
	}
	return
}

func (v *Variable) shortening(byteAddress string) bool {
	if strings.HasPrefix(byteAddress, "X") || strings.HasPrefix(byteAddress, "B") || strings.HasPrefix(byteAddress, "W") || strings.HasPrefix(byteAddress, "D") || strings.HasPrefix(byteAddress, "S") {
		return false
	} else if strings.HasPrefix(byteAddress, "DBX") || strings.HasPrefix(byteAddress, "DBB") || strings.HasPrefix(byteAddress, "DBW") || strings.HasPrefix(byteAddress, "DBD") || strings.HasPrefix(byteAddress, "DBS") {
		return false
	}
	return true
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

type S7Device struct {
	runtime.DeviceMeta
	CollectorCycle   uint        `json:"collectorCycle"`                    // 采集周期
	VariableInterval uint        `json:"variableInterval"`                  // 变量间隔
	Address          *S7Address  `json:"address"`                           // IP地址
	Variables        []*Variable `json:"variables" binding:"required,dive"` // 自定义变量
}

type S7Address struct {
	Location string           `json:"location"` // 地址路径
	Option   *S7AddressOption `json:"option"`   // 地址其他参数
}

type S7AddressOption struct {
	Port uint  `json:"port"`           // 端口号
	Rack uint8 `json:"rack,omitempty"` // 机架号
	Slot uint8 `json:"slot,omitempty"` // 槽位号
}

type VariableSlice []*Variable

func (vs VariableSlice) Len() int {
	return len(vs)
}

func (vs VariableSlice) Less(i, j int) bool {
	switch vs[i].Zone() {
	case I, Q, M:
		_, _, addressI, bitI := vs[i].ParseVariableAddress()
		_, _, addressJ, bitJ := vs[j].ParseVariableAddress()
		if addressI != addressJ {
			return addressI < addressJ
		} else {
			return bitI < bitJ
		}
	case DB:
		_, bsI, addressI, bitI := vs[i].ParseVariableAddress()
		_, bsJ, addressJ, bitJ := vs[j].ParseVariableAddress()
		if bsI != bsJ {
			return bsI < bsJ
		} else {
			if addressI != addressJ {
				return addressI < addressJ
			} else {
				return bitI < bitJ
			}
		}
	}
	return false
}

func (vs VariableSlice) Swap(i, j int) {
	vs[i], vs[j] = vs[j], vs[i]
}

type ParseVariableResult struct {
	VariableSlice VariableSlice
	Err           []error
}
