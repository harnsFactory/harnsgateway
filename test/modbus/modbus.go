package main

func main() {
	// demo()
	// cestc()
}

// func cestc() {
//
// 	v1 := &runtime.Variable{
// 		DataType:     modbus.BOOL,
// 		Name:         "故障",
// 		Address:      8,
// 		Bits:         0,
// 		FunctionCode: 1,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v2 := &runtime.Variable{
// 		DataType:     modbus.BOOL,
// 		Name:         "切换",
// 		Address:      6,
// 		Bits:         0,
// 		FunctionCode: 1,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v3 := &runtime.Variable{
// 		DataType:     modbus.BOOL,
// 		Name:         "就地远程",
// 		Address:      5,
// 		Bits:         0,
// 		FunctionCode: 1,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v4 := &runtime.Variable{
// 		DataType:     modbus.BOOL,
// 		Name:         "急停",
// 		Address:      4,
// 		Bits:         0,
// 		FunctionCode: 1,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v5 := &runtime.Variable{
// 		DataType:     modbus.BOOL,
// 		Name:         "手动停止",
// 		Address:      3,
// 		Bits:         0,
// 		FunctionCode: 1,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v6 := &runtime.Variable{
// 		DataType:     modbus.BOOL,
// 		Name:         "电机启动",
// 		Address:      2,
// 		Bits:         0,
// 		FunctionCode: 1,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v7 := &runtime.Variable{
// 		DataType:     modbus.FLOAT32,
// 		Name:         "电位计反馈",
// 		Address:      305,
// 		Bits:         0,
// 		FunctionCode: 3,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v8 := &runtime.Variable{
// 		DataType:     modbus.FLOAT32,
// 		Name:         "电压表设置",
// 		Address:      353,
// 		Bits:         0,
// 		FunctionCode: 3,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v9 := &runtime.Variable{
// 		DataType:     modbus.FLOAT32,
// 		Name:         "电机电流反馈",
// 		Address:      301,
// 		Bits:         0,
// 		FunctionCode: 3,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v10 := &runtime.Variable{
// 		DataType:     modbus.FLOAT32,
// 		Name:         "进线电压反馈",
// 		Address:      303,
// 		Bits:         0,
// 		FunctionCode: 3,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v11 := &runtime.Variable{
// 		DataType:     modbus.FLOAT32,
// 		Name:         "频率给点",
// 		Address:      351,
// 		Bits:         0,
// 		FunctionCode: 3,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	d := &runtime.Device{
// 		ID:               "abcd",
// 		Model:            "modbus-tcp",
// 		CollectorCycle:   1,
// 		VariableInterval: 0,
// 		TcpChannel:       2,
// 		Address:          "10.56.50.223",
// 		Port:             502,
// 		Slave:            1,
// 		MemoryLayout:     modbus.CDAB,
// 		PositionAddress:  0,
// 		Variables:        []*runtime.Variable{v1, v2, v3, v4, v5, v6, v7, v8, v9, v10, v11},
// 	}
//
// 	collector, _ := modbus.NewCollector(d)
// 	collector.Init()
// 	collector.Collect()
// 	<-time.After(1 * time.Hour)
// }
//
// func demo() {
// 	v1 := &runtime.Variable{
// 		DataType:     modbus.INT16,
// 		Name:         "A",
// 		Address:      1,
// 		Bits:         0,
// 		FunctionCode: 3,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v2 := &runtime.Variable{
// 		DataType:     modbus.INT16,
// 		Name:         "B",
// 		Address:      2,
// 		Bits:         0,
// 		FunctionCode: 3,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v3 := &runtime.Variable{
// 		DataType:     modbus.INT32,
// 		Name:         "C",
// 		Address:      17,
// 		Bits:         0,
// 		FunctionCode: 3,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v4 := &runtime.Variable{
// 		DataType:     modbus.INT32,
// 		Name:         "D",
// 		Address:      123,
// 		Bits:         0,
// 		FunctionCode: 3,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v5 := &runtime.Variable{
// 		DataType:     modbus.INT64,
// 		Name:         "E",
// 		Address:      134,
// 		Bits:         0,
// 		FunctionCode: 3,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v6 := &runtime.Variable{
// 		DataType:     modbus.FLOAT32,
// 		Name:         "F",
// 		Address:      155,
// 		Bits:         0,
// 		FunctionCode: 3,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v7 := &runtime.Variable{
// 		DataType:     modbus.FLOAT64,
// 		Name:         "H",
// 		Address:      246,
// 		Bits:         0,
// 		FunctionCode: 3,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v8 := &runtime.Variable{
// 		DataType:     modbus.BOOL,
// 		Name:         "I",
// 		Address:      277,
// 		Bits:         3,
// 		FunctionCode: 3,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v9 := &runtime.Variable{
// 		DataType:     modbus.BOOL,
// 		Name:         "J",
// 		Address:      277,
// 		Bits:         5,
// 		FunctionCode: 3,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
// 	v10 := &runtime.Variable{
// 		DataType:     modbus.INT16,
// 		Name:         "K",
// 		Address:      477,
// 		Bits:         0,
// 		FunctionCode: 3,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v11 := &runtime.Variable{
// 		DataType:     modbus.INT16,
// 		Name:         "L",
// 		Address:      677,
// 		Bits:         0,
// 		FunctionCode: 3,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v12 := &runtime.Variable{
// 		DataType:     modbus.INT16,
// 		Name:         "M",
// 		Address:      877,
// 		Bits:         0,
// 		FunctionCode: 3,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v13 := &runtime.Variable{
// 		DataType:     modbus.INT16,
// 		Name:         "N",
// 		Address:      1000,
// 		Bits:         0,
// 		FunctionCode: 3,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v14 := &runtime.Variable{
// 		DataType:     modbus.INT16,
// 		Name:         "O",
// 		Address:      900,
// 		Bits:         0,
// 		FunctionCode: 3,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v15 := &runtime.Variable{
// 		DataType:     modbus.INT16,
// 		Name:         "P",
// 		Address:      800,
// 		Bits:         0,
// 		FunctionCode: 3,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v16 := &runtime.Variable{
// 		DataType:     modbus.INT16,
// 		Name:         "Q",
// 		Address:      700,
// 		Bits:         0,
// 		FunctionCode: 3,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v17 := &runtime.Variable{
// 		DataType:     modbus.BOOL,
// 		Name:         "A1True",
// 		Address:      1,
// 		Bits:         0,
// 		FunctionCode: 1,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v18 := &runtime.Variable{
// 		DataType:     modbus.BOOL,
// 		Name:         "A3True",
// 		Address:      3,
// 		Bits:         0,
// 		FunctionCode: 1,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v19 := &runtime.Variable{
// 		DataType:     modbus.BOOL,
// 		Name:         "A5True",
// 		Address:      5,
// 		Bits:         0,
// 		FunctionCode: 1,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v20 := &runtime.Variable{
// 		DataType:     modbus.BOOL,
// 		Name:         "A7True",
// 		Address:      7,
// 		Bits:         0,
// 		FunctionCode: 1,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v21 := &runtime.Variable{
// 		DataType:     modbus.BOOL,
// 		Name:         "A10True",
// 		Address:      10,
// 		Bits:         0,
// 		FunctionCode: 1,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v22 := &runtime.Variable{
// 		DataType:     modbus.BOOL,
// 		Name:         "A12True",
// 		Address:      12,
// 		Bits:         0,
// 		FunctionCode: 1,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v23 := &runtime.Variable{
// 		DataType:     modbus.BOOL,
// 		Name:         "A2",
// 		Address:      2,
// 		Bits:         0,
// 		FunctionCode: 1,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v24 := &runtime.Variable{
// 		DataType:     modbus.BOOL,
// 		Name:         "A9",
// 		Address:      9,
// 		Bits:         0,
// 		FunctionCode: 1,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v25 := &runtime.Variable{
// 		DataType:     modbus.BOOL,
// 		Name:         "A4",
// 		Address:      4,
// 		Bits:         0,
// 		FunctionCode: 1,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v26 := &runtime.Variable{
// 		DataType:     modbus.BOOL,
// 		Name:         "A6",
// 		Address:      6,
// 		Bits:         0,
// 		FunctionCode: 1,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	v27 := &runtime.Variable{
// 		DataType:     modbus.BOOL,
// 		Name:         "A8",
// 		Address:      8,
// 		Bits:         0,
// 		FunctionCode: 1,
// 		Rate:         0,
// 		Amount:       0,
// 		DefaultValue: nil,
// 	}
//
// 	d := &runtime.Device{
// 		ID:               "abcd",
// 		Model:            "modbus-tcp",
// 		CollectorCycle:   1,
// 		VariableInterval: 0,
// 		TcpChannel:       2,
// 		Address:          "127.0.0.1",
// 		Port:             505,
// 		Slave:            3,
// 		MemoryLayout:     modbus.ABCD,
// 		PositionAddress:  0,
// 		Variables:        []*runtime.Variable{v1, v2, v3, v4, v5, v6, v7, v8, v9, v10, v11, v12, v13, v14, v15, v16, v17, v18, v19, v20, v21, v22, v23, v24, v25, v26, v27},
// 	}
//
// 	collector, _ := modbus.NewCollector(d)
// 	collector.Init()
// 	collector.Collect()
// 	<-time.After(1 * time.Hour)
// }
