package runtime

import (
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/utils/binutil"
)

var _ runtime.Device = (*ModBusDevice)(nil)

type Variable struct {
	DataType     runtime.DataType `json:"dataType"`               // bool、int16、float32、float64、int32、int64、uint16
	Name         string           `json:"name"`                   // 变量名称
	Address      uint             `json:"address"`                // 变量地址
	Bits         uint8            `json:"bits"`                   // 位 1、2、3、4、5、6、7、8、9、10、11、12、13、14、15、16
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
	Address          *Address             `json:"address"`                           // IP地址\串口地址
	Slave            uint                 `json:"slave"`                             // 下位机号
	MemoryLayout     runtime.MemoryLayout `json:"memoryLayout"`                      // 内存布局 DCBA CDAB BADC ABCD
	PositionAddress  uint                 `json:"positionAddress"`                   // 起始地址
	Variables        []*Variable          `json:"variables" binding:"required,dive"` // 自定义变量
	VariablesMap     map[string]*Variable `json:"-"`                                 // 自定义变量Map
}

func (m *ModBusDevice) IndexDevice() {
	m.VariablesMap = make(map[string]*Variable)
	for _, variable := range m.Variables {
		m.VariablesMap[variable.Name] = variable
	}
}

func (m *ModBusDevice) GetVariablesMap() map[string]runtime.VariableValue {
	vm := make(map[string]runtime.VariableValue)
	for k, variable := range m.VariablesMap {
		vm[k] = variable
	}
	return vm
}

type Address struct {
	Location string  `json:"location"` // 地址路径
	Option   *Option `json:"option"`   // 地址其他参数
}

type Option struct {
	Port     int              `json:"port,omitempty"`     // 端口号
	BaudRate int              `json:"baudRate,omitempty"` // 波特率
	DataBits int              `json:"dataBits,omitempty"` // 数据位
	Parity   runtime.Parity   `json:"parity,omitempty"`   // 校验位
	StopBits runtime.StopBits `json:"stopBits,omitempty"` // 停止位
}

type VariableSlice []*Variable

func (vs VariableSlice) Len() int {
	return len(vs)
}

func (vs VariableSlice) Less(i, j int) bool {
	return vs[i].Address < vs[j].Address
}

func (vs VariableSlice) Swap(i, j int) {
	vs[i], vs[j] = vs[j], vs[i]
}

type ParseVariableResult struct {
	VariableSlice VariableSlice
	Err           []error
}

type VariableParse struct {
	Variable *Variable
	Start    uint // 报文中数据[]byte开始位置
}

type ModBusDataFrame struct {
	Slave             uint
	MemoryLayout      runtime.MemoryLayout
	StartAddress      uint
	FunctionCode      uint8
	MaxDataSize       uint // 最大数量01 代表线圈  03代表word
	TransactionId     uint16
	DataFrame         []byte
	ResponseDataFrame []byte
	Variables         []*VariableParse
}

func (df *ModBusDataFrame) WriteTransactionId() {
	df.TransactionId++
	id := df.TransactionId
	binutil.WriteUint16BigEndian(df.DataFrame, id)
}

func (df *ModBusDataFrame) ParseVariableValue(data []byte) []*Variable {
	vvs := make([]*Variable, 0, len(df.Variables))
	for _, vp := range df.Variables {
		var value interface{}
		switch FunctionCode(df.FunctionCode) {
		case ReadInputStatus, ReadCoilStatus:
			switch vp.Variable.DataType {
			case runtime.BOOL:
				v := int(data[vp.Start])
				value = v == 1
			case runtime.INT16:
				value = int16(data[vp.Start])
			case runtime.UINT16:
				value = uint16(data[vp.Start])
			case runtime.INT32:
				value = int32(data[vp.Start])
			case runtime.INT64:
				value = int64(data[vp.Start])
			case runtime.FLOAT32:
				value = float32(data[vp.Start])
			case runtime.FLOAT64:
				value = float64(data[vp.Start])
			}
		case ReadInputRegister, ReadHoldRegister:
			vpData := data[vp.Start:]
			switch vp.Variable.DataType {
			case runtime.BOOL:
				var v int16
				switch df.MemoryLayout {
				case runtime.ABCD, runtime.CDAB:
					v = int16(binutil.ParseUint16BigEndian(vpData))
				case runtime.BADC, runtime.DCBA:
					v = int16(binutil.ParseUint16LittleEndian(vpData))
				}
				value = 1<<(vp.Variable.Bits-1)&v != 0
			case runtime.INT16:
				var v interface{}
				switch df.MemoryLayout {
				case runtime.ABCD, runtime.CDAB:
					v = int16(binutil.ParseUint16BigEndian(vpData))
				case runtime.BADC, runtime.DCBA:
					v = int16(binutil.ParseUint16LittleEndian(vpData))
				}
				if vp.Variable.Rate != 0 && vp.Variable.Rate != 1 {
					value = int16((v.(float64)) * vp.Variable.Rate)
				} else {
					value = v
				}
			case runtime.UINT16:
				var v interface{}
				switch df.MemoryLayout {
				case runtime.ABCD, runtime.CDAB:
					v = binutil.ParseUint16BigEndian(vpData)
				case runtime.BADC, runtime.DCBA:
					v = binutil.ParseUint16LittleEndian(vpData)
				}
				if vp.Variable.Rate != 0 && vp.Variable.Rate != 1 {
					value = uint16((v.(float64)) * vp.Variable.Rate)
				} else {
					value = v
				}
			case runtime.INT32:
				var v interface{}
				switch df.MemoryLayout {
				case runtime.ABCD:
					v = int32(binutil.ParseUint32BigEndian(vpData))
				case runtime.BADC:
					// 大端交换
					v = int32(binutil.ParseUint32BigEndianByteSwap(vpData))
				case runtime.CDAB:
					v = int32(binutil.ParseUint32LittleEndianByteSwap(vpData))
				case runtime.DCBA:
					v = int32(binutil.ParseUint32LittleEndian(vpData))
				}
				if vp.Variable.Rate != 0 && vp.Variable.Rate != 1 {
					value = int32((v.(float64)) * vp.Variable.Rate)
				} else {
					value = v
				}
			case runtime.INT64:
				var v interface{}
				switch df.MemoryLayout {
				case runtime.ABCD:
					v = int64(binutil.ParseUint64BigEndian(vpData))
				case runtime.BADC:
					v = int64(binutil.ParseUint64BigEndianByteSwap(vpData))
				case runtime.CDAB:
					v = int64(binutil.ParseUint64LittleEndianByteSwap(vpData))
				case runtime.DCBA:
					v = int64(binutil.ParseUint64LittleEndian(vpData))
				}
				if vp.Variable.Rate != 0 && vp.Variable.Rate != 1 {
					value = int64((v.(float64)) * vp.Variable.Rate)
				} else {
					value = v
				}
			case runtime.FLOAT32:
				var v interface{}
				switch df.MemoryLayout {
				case runtime.ABCD:
					v = binutil.ParseFloat32BigEndian(vpData)
				case runtime.BADC:
					v = binutil.ParseFloat32BigEndianByteSwap(vpData)
				case runtime.CDAB:
					v = binutil.ParseFloat32LittleEndianByteSwap(vpData)
				case runtime.DCBA:
					v = binutil.ParseFloat32LittleEndian(vpData)
				}
				if vp.Variable.Rate != 0 && vp.Variable.Rate != 1 {
					value = float32((v.(float64)) * vp.Variable.Rate)
				} else {
					value = v
				}
			case runtime.FLOAT64:
				var v interface{}
				switch df.MemoryLayout {
				case runtime.ABCD:
					v = binutil.ParseFloat64BigEndian(vpData)
				case runtime.BADC:
					v = binutil.ParseFloat64BigEndianByteSwap(vpData)
				case runtime.CDAB:
					v = binutil.ParseFloat64LittleEndianByteSwap(vpData)
				case runtime.DCBA:
					v = binutil.ParseFloat64LittleEndian(vpData)
				}
				if vp.Variable.Rate != 0 && vp.Variable.Rate != 1 {
					value = (v.(float64)) * vp.Variable.Rate
				} else {
					value = v
				}
			}
		}

		vp.Variable.SetValue(value)
		vvs = append(vvs, &Variable{
			DataType:     vp.Variable.DataType,
			Name:         vp.Variable.Name,
			Address:      vp.Variable.Address,
			Bits:         vp.Variable.Bits,
			FunctionCode: vp.Variable.FunctionCode,
			Rate:         vp.Variable.Rate,
			Amount:       vp.Variable.Amount,
			DefaultValue: vp.Variable.DefaultValue,
			Value:        vp.Variable.Value,
		})
	}
	return vvs
}
