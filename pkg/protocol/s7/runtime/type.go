package runtime

import (
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/runtime/constant"
	"k8s.io/klog/v2"
	"strconv"
	"strings"
)

var _ runtime.Device = (*S7Device)(nil)
var _ runtime.VariableValue = (*Variable)(nil)

type Variable struct {
	DataType     constant.DataType   `json:"dataType"`               // bool、int16、float32、float64、int32、int64、uint16
	Name         string              `json:"name"`                   // 变量名称
	Address      string              `json:"address"`                // 变量地址
	Rate         float64             `json:"rate,omitempty"`         // 比率
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

func (v *Variable) DataRequestLength(area S7StoreArea) uint16 {
	switch area {
	// 以byte形式读取 发送的一个item占12个字节    返回的一个item至少占5个字节
	case I, Q, M:
		switch v.DataType {
		case constant.BOOL:
			return uint16(1)
		case constant.STRING:
			return uint16(1)
		case constant.UINT16:
			return uint16(2)
		case constant.INT16:
			return uint16(2)
		case constant.INT32:
			return uint16(4)
		case constant.FLOAT32:
			return uint16(4)
		case constant.INT64:
			return uint16(8)
		case constant.FLOAT64:
			return uint16(8)
		default:
			return uint16(1)
		}
	// 以word形式读取 发送的一个item占12个字节    返回的一个item至少占8个字节
	case DB:
		switch v.DataType {
		case constant.BOOL:
			return uint16(1)
		case constant.STRING:
			return uint16(1)
		case constant.UINT16:
			return uint16(1)
		case constant.INT16:
			return uint16(1)
		case constant.INT32:
			return uint16(2)
		case constant.FLOAT32:
			return uint16(2)
		case constant.INT64:
			return uint16(4)
		case constant.FLOAT64:
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
		case constant.BOOL:
			return uint16(1)
		case constant.STRING:
			return uint16(1)
		case constant.UINT16:
			return uint16(2)
		case constant.INT16:
			return uint16(2)
		case constant.INT32:
			return uint16(4)
		case constant.FLOAT32:
			return uint16(4)
		case constant.INT64:
			return uint16(8)
		case constant.FLOAT64:
			return uint16(8)
		default:
			return uint16(1)
		}
	case DB:
		switch v.DataType {
		case constant.BOOL:
			return uint16(2)
		case constant.STRING:
			return uint16(2)
		case constant.UINT16:
			return uint16(2)
		case constant.INT16:
			return uint16(2)
		case constant.INT32:
			return uint16(4)
		case constant.FLOAT32:
			return uint16(4)
		case constant.INT64:
			return uint16(8)
		case constant.FLOAT64:
			return uint16(8)
		default:
			return uint16(2)
		}
	default:
		return uint16(1)
	}
}

// 数据类型对应的位个数 s7 write 变量时 item中的length
func (v *Variable) DataTypeBitLength() uint16 {
	switch v.DataType {
	case constant.BOOL:
		return uint16(1)
	case constant.STRING:
		// todo
	case constant.UINT16:
		return uint16(16)
	case constant.INT16:
		return uint16(16)
	case constant.INT32:
		return uint16(32)
	case constant.FLOAT32:
		return uint16(32)
	case constant.INT64:
		return uint16(64)
	case constant.FLOAT64:
		return uint16(64)
	default:
		return uint16(16)
	}
	return uint16(16)
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

type S7Device struct {
	runtime.DeviceMeta
	CollectorCycle   uint                 `json:"collectorCycle"`                    // 采集周期
	VariableInterval uint                 `json:"variableInterval"`                  // 变量间隔
	Address          *S7Address           `json:"address"`                           // IP地址
	Variables        []*Variable          `json:"variables" binding:"required,dive"` // 自定义变量
	VariablesMap     map[string]*Variable `json:"-"`
}

func (s *S7Device) IndexDevice() {
	s.VariablesMap = make(map[string]*Variable)
	for _, variable := range s.Variables {
		s.VariablesMap[variable.Name] = variable
	}
}

func (s *S7Device) GetVariable(key string) (rv runtime.VariableValue, exist bool) {
	if v, isExist := s.VariablesMap[key]; isExist {
		rv = v
		exist = isExist
	}
	return
}

type S7Address struct {
	Location string           `json:"location"` // 地址路径
	Option   *S7AddressOption `json:"option"`   // 地址其他参数
}

type S7AddressOption struct {
	Port uint  `json:"port,omitempty"` // 端口号
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

type ParameterData struct {
	ParameterItem []byte
	DataItem      []byte
}
