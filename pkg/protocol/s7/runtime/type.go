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

func (v *Variable) DataLength() uint16 {
	switch v.DataType {
	case runtime.BOOL:
		return uint16(1)
	case runtime.STRING:
		// todo 字符串长度
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
}

func (v *Variable) Zone() S7StoreAddress {
	if strings.HasPrefix(v.Address, "I") {
		return I
	} else if strings.HasPrefix(v.Address, "M") {
		return M
	} else if strings.HasPrefix(v.Address, "DB") {
		return DB
	} else if strings.HasPrefix(v.Address, "Q") {
		return Q
	}
	// todo
	return I
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
	// todo
	return 0
}

func (v *Variable) ParseVariableAddress() (zone S7StoreAddress, blockSize uint, addressType AddressType, startAddress uint32, bitAddress uint8) {
	zone = v.Zone()
	switch zone {
	case I:
		// I10.0
		// I10.2
		blockSize = 0
		addressType = Bool
		index := strings.LastIndex(v.Address, ".")
		startAddressString := v.Address[1:index]
		i, err := strconv.Atoi(startAddressString)
		if err != nil {
			klog.V(2).InfoS("Failed to read s7 variable address", "variableName", v.Name)
		}
		startAddress = uint32(i)
		bitAddressString := v.Address[index+1:]
		j, err := strconv.Atoi(bitAddressString)
		if err != nil {
			klog.V(2).InfoS("Failed to read s7 variable address", "variableName", v.Name)
		}
		bitAddress = uint8(j)
	case Q:
		// todo
	case M:
		// todo
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
		blockSize = uint(bs)
		address := v.Address[index+1:]
		if strings.HasPrefix(address, "DBX") {
			addressType = Bool
			startAddressString := address[3:strings.LastIndex(address, ".")]
			i, err := strconv.Atoi(startAddressString)
			if err != nil {
				klog.V(2).InfoS("Failed to read s7 variable address", "variableName", v.Name)
			}
			startAddress = uint32(i)
			bitAddressString := address[strings.LastIndex(address, ".")+1:]
			j, err := strconv.Atoi(bitAddressString)
			if err != nil {
				klog.V(2).InfoS("Failed to read s7 variable address", "variableName", v.Name)
			}
			bitAddress = uint8(j)
		} else if strings.HasPrefix(address, "DBD") {
			addressType = DWord
			startAddressString := address[3:]
			i, err := strconv.Atoi(startAddressString)
			if err != nil {
				klog.V(2).InfoS("Failed to read s7 variable address", "variableName", v.Name)
			}
			startAddress = uint32(i)
			bitAddress = uint8(0)
		} else if strings.HasPrefix(address, "DBW") {
			addressType = Word
			startAddressString := address[3:]
			i, err := strconv.Atoi(startAddressString)
			if err != nil {
				klog.V(2).InfoS("Failed to read s7 variable address", "variableName", v.Name)
			}
			startAddress = uint32(i)
			bitAddress = uint8(0)
		} else if strings.HasPrefix(address, "B") {
			addressType = Byte
			startAddressString := address[1:]
			i, err := strconv.Atoi(startAddressString)
			if err != nil {
				klog.V(2).InfoS("Failed to read s7 variable address", "variableName", v.Name)
			}
			startAddress = uint32(i)
			bitAddress = uint8(0)
		} else if strings.HasPrefix(address, "STRING") {
			addressType = String
			startAddressString := address[6:strings.LastIndex(address, ".")]
			i, err := strconv.Atoi(startAddressString)
			if err != nil {
				klog.V(2).InfoS("Failed to read s7 variable address", "variableName", v.Name)
			}
			startAddress = uint32(i)
			bitAddressString := address[index+1:]
			j, err := strconv.Atoi(bitAddressString)
			if err != nil {
				klog.V(2).InfoS("Failed to read s7 variable address", "variableName", v.Name)
			}
			// length
			bitAddress = uint8(j)
		}
	}
	return
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
	Address          string      `json:"address"`                           // IP地址\串口地址
	Port             uint        `json:"port"`                              // 端口号
	Rack             uint8       `json:"rack,omitempty"`                    // 机架号
	Slot             uint8       `json:"slot,omitempty"`                    // 槽位号
	Variables        []*Variable `json:"variables" binding:"required,dive"` // 自定义变量
}

type VariableSlice []*Variable

func (vs VariableSlice) Len() int {
	return len(vs)
}

func (vs VariableSlice) Less(i, j int) bool {
	switch vs[i].Zone() {
	case I:
		_, _, _, startAddressI, bitAddressI := vs[i].ParseVariableAddress()
		_, _, _, startAddressJ, bitAddressJ := vs[j].ParseVariableAddress()
		if startAddressI != startAddressJ {
			return startAddressI < startAddressJ
		} else {
			return bitAddressI < bitAddressJ
		}
	case DB:
		_, bsI, _, startAddressI, _ := vs[i].ParseVariableAddress()
		_, bsJ, _, startAddressJ, _ := vs[j].ParseVariableAddress()
		if bsI != bsJ {
			return bsI < bsJ
		} else {
			return startAddressI < startAddressJ
		}
	// todo
	case M:
	case Q:
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
