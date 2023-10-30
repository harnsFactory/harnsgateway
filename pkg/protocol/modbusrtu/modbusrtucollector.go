package modbusrtu

import (
	"container/list"
	"context"
	"errors"
	"go.bug.st/serial"
	modbusruntime "harnsgateway/pkg/protocol/modbus/runtime"
	modbusrturuntime "harnsgateway/pkg/protocol/modbusrtu/runtime"
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/utils/binutil"
	"harnsgateway/pkg/utils/crcutil"
	"k8s.io/klog/v2"
	"sort"
	"sync"
	"time"
)

/**
modbus 协议 ADU = 地址(1) + pdu(253) + 16位校验(2) = 256
modbus Rtu报文
地址(1)   +   pdu(253) = 256
*/

// ModBusRtuDataFrame 报文对应的数据点位
type ModBusRtuDataFrame struct {
	MemoryLayout      runtime.MemoryLayout
	StartAddress      uint
	FunctionCode      uint8
	MaxDataSize       uint // 最大数量01 代表线圈  03代表word
	DataFrame         []byte
	ResponseDataFrame []byte
	Variables         []*VariableParse
}

func (df *ModBusRtuDataFrame) GenerateReadMessage(slave uint, functionCode uint8, startAddress uint, maxDataSize uint) {
	df.FunctionCode = functionCode
	df.StartAddress = startAddress
	df.MaxDataSize = maxDataSize
	// 01 03 00 00 00 0A C5 CD
	// 01  设备地址
	// 03  功能码
	// 00 00  起始地址
	// 00 0A  寄存器数量(word数量)/线圈数量
	// C5 CD  crc16检验码
	message := make([]byte, 6)
	message[0] = byte(slave)
	message[1] = byte(functionCode)
	binutil.WriteUint16(message[2:], uint16(startAddress))
	binutil.WriteUint16(message[4:], uint16(df.MaxDataSize))
	crc16 := make([]byte, 2)
	binutil.WriteUint16(crc16, crcutil.CheckCrc16sum(message))
	message = append(message, crc16...)
	df.DataFrame = message
	bytesLength := 0
	switch functionCode {
	case 1, 2:
		if maxDataSize%8 == 0 {
			bytesLength = int(maxDataSize/8 + 5)
		} else {
			bytesLength = int(maxDataSize/8 + 1 + 5)
		}
	case 3, 4:
		bytesLength = int(maxDataSize*2 + 5)
	}
	df.ResponseDataFrame = make([]byte, bytesLength)
}

func (df *ModBusRtuDataFrame) ValidateMessage(least int) ([]byte, error) {
	buf := df.ResponseDataFrame[:least]

	functionCode := buf[1]
	if functionCode&0x80 > 0 {
		klog.V(2).InfoS("Failed to parse modbus rtu message", "error code", functionCode-128)
		return nil, modbusrturuntime.ErrMessageFunctionCodeError
	}

	byteDataLength := buf[2]
	checkBufData := buf[:byteDataLength+3]
	sum := crcutil.CheckCrc16sum(checkBufData)
	crc := binutil.ParseUint16BigEndian(buf[byteDataLength+3 : byteDataLength+5])
	if sum != crc {
		klog.V(2).InfoS("Failed to check CRC16")
		return nil, modbusrturuntime.ErrCRC16Error
	}
	return buf, nil
}

func (df *ModBusRtuDataFrame) ParseVariableValue(data []byte) []*modbusrturuntime.Variable {
	vvs := make([]*modbusrturuntime.Variable, 0, len(df.Variables))
	for _, vp := range df.Variables {
		var value interface{}
		switch df.FunctionCode {
		case 1, 2:
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
		case 3, 4:
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
					v = int64(binutil.ParseUint64BigEndianByteData(vpData))
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
					v = binutil.ParseUint32LittleEndianByteSwap(vpData)
				case runtime.DCBA:
					v = binutil.ParseUint32LittleEndian(vpData)
				}
				if vp.Variable.Rate != 0 && vp.Variable.Rate != 1 {
					value = (v.(float64)) * vp.Variable.Rate
				} else {
					value = v
				}
			}
		}

		vvs = append(vvs, &modbusrturuntime.Variable{
			DataType:     vp.Variable.DataType,
			Name:         vp.Variable.Name,
			Address:      vp.Variable.Address,
			Bits:         vp.Variable.Bits,
			FunctionCode: vp.Variable.FunctionCode,
			Rate:         vp.Variable.Rate,
			Amount:       vp.Variable.Amount,
			DefaultValue: vp.Variable.DefaultValue,
			Value:        value,
		})
	}
	return vvs
}

type VariableParse struct {
	Variable *modbusrturuntime.Variable
	Start    uint // 报文中数据[]byte开始位置
}

type ModbusRtuCollector struct {
	exitCh                   chan struct{}
	Device                   *modbusrturuntime.ModBusRtuDevice
	Clients                  *SerialClients
	FunctionCodeDataFrameMap map[uint8][]*ModBusRtuDataFrame
	VariableCount            int
	VariableCh               chan *runtime.ParseVariableResult
	CanCollect               bool
}

func NewCollector(d runtime.Device) (runtime.Collector, chan *runtime.ParseVariableResult, error) {
	device, ok := d.(*modbusrturuntime.ModBusRtuDevice)
	if !ok {
		klog.V(2).InfoS("Failed to new modbus rtu collector,device type not supported")
		return nil, nil, modbusrturuntime.ErrDeviceType
	}
	VariableCount := 0
	CanCollect := false
	functionCodeDataFrameMap := make(map[uint8][]*ModBusRtuDataFrame, 0)
	variables := device.Variables
	functionCodeVariableMap := make(map[uint8][]*modbusrturuntime.Variable, 0)
	for _, variable := range variables {
		functionCodeVariableMap[variable.FunctionCode] = append(functionCodeVariableMap[variable.FunctionCode], variable)
	}
	// functionCode03 modbus rtu 每次请求最大字节数256 从站地址1字节 功能码1字节 CRC2字节 数据位252字节 word最多读取125个寄存器
	// functionCode01 modbus 一次最多读取251个字节 总共251 * 8 = 2008个线圈
	for code, variables := range functionCodeVariableMap {
		VariableCount = VariableCount + len(variables)
		sort.Sort(modbusrturuntime.VariableSlice(variables))
		dfs := make([]*ModBusRtuDataFrame, 0)
		firstVariable := variables[0]
		startOffset := firstVariable.Address - device.PositionAddress
		startAddress := startOffset
		var maxDataSize uint = 0
		df := &ModBusRtuDataFrame{MemoryLayout: device.MemoryLayout}
		switch code {
		case 1, 2:
			limitDataSize := startAddress + 2007
			for i := 0; i < len(variables); i++ {
				variable := variables[i]
				if variable.Address < limitDataSize {
					vp := &VariableParse{
						Variable: variable,
						Start:    variable.Address - startAddress,
					}
					df.Variables = append(df.Variables, vp)
					maxDataSize = variable.Address - startAddress + 1
				} else {
					df.GenerateReadMessage(device.Slave, code, startAddress, maxDataSize)
					dfs = append(dfs, df)
					df = &ModBusRtuDataFrame{MemoryLayout: device.MemoryLayout}
					maxDataSize = 0
					startAddress = variable.Address
					limitDataSize = startAddress + 2007
					i--
				}
			}
		case 3, 4:
			limitDataSize := startAddress + 124
			for i := 0; i < len(variables); i++ {
				variable := variables[i]
				if variable.Address+runtime.DataTypeWord[variable.DataType] <= limitDataSize {
					vp := &VariableParse{
						Variable: variable,
						Start:    (variable.Address - startAddress) * 2,
					}
					df.Variables = append(df.Variables, vp)
					maxDataSize = variable.Address - startAddress + runtime.DataTypeWord[variable.DataType]
				} else {
					df.GenerateReadMessage(device.Slave, code, startAddress, maxDataSize)
					dfs = append(dfs, df)
					df = &ModBusRtuDataFrame{MemoryLayout: device.MemoryLayout}
					maxDataSize = 0
					startAddress = variable.Address
					limitDataSize = startAddress + 124
					i--
				}
			}
		}
		if len(df.Variables) > 0 {
			df.GenerateReadMessage(device.Slave, code, startAddress, maxDataSize)
			dfs = append(dfs, df)
		}
		functionCodeDataFrameMap[code] = append(functionCodeDataFrameMap[code], dfs...)
	}

	if len(functionCodeDataFrameMap) == 0 {
		klog.V(2).InfoS("Failed to collect modbus rtu.Because of the variables is empty", "deviceId", device.ID)
		return nil, nil, nil
	}

	CanCollect = true
	mode := &serial.Mode{
		BaudRate: device.BaudRate,
		Parity:   modbusrturuntime.ParityToParity[device.Parity],
		DataBits: device.DataBits,
		StopBits: modbusrturuntime.StopBitsToStopBits[device.StopBits],
	}
	port, err := serial.Open(device.Address, mode)
	if err != nil {
		klog.V(2).InfoS("Failed to connect serial port", "address", device.Address)
		return nil, nil, err
	}

	cs := list.New()
	cs.PushBack(&SerialClient{
		Timeout: 1,
		Port:    port,
	})

	scs := &SerialClients{
		Clients:      cs,
		Max:          1,
		Idle:         1,
		Mux:          &sync.Mutex{},
		NextRequest:  1,
		ConnRequests: make(map[uint64]chan *SerialClient, 0),
		newSerialClient: func() (*SerialClient, error) {
			newPort, err := serial.Open(device.Address, mode)
			if err != nil {
				klog.V(2).InfoS("Failed to connect serial port", "address", device.Address)
				return nil, err
			}
			return &SerialClient{
				Timeout: 1,
				Port:    newPort,
			}, nil
		},
	}

	mtc := &ModbusRtuCollector{
		Device:                   device,
		exitCh:                   make(chan struct{}, 0),
		FunctionCodeDataFrameMap: functionCodeDataFrameMap,
		Clients:                  scs,
		VariableCh:               make(chan *runtime.ParseVariableResult, 1),
		VariableCount:            VariableCount,
		CanCollect:               CanCollect,
	}
	return mtc, mtc.VariableCh, nil
}

func (collector *ModbusRtuCollector) Destroy(ctx context.Context) {
	collector.exitCh <- struct{}{}
	collector.Clients.Destroy(ctx)
	close(collector.VariableCh)
}

func (collector *ModbusRtuCollector) Collect(ctx context.Context) {
	if collector.CanCollect {
		go func() {
			for {
				start := time.Now().Unix()
				if !collector.poll(ctx) {
					return
				}
				select {
				case <-collector.exitCh:
					return
				default:
					end := time.Now().Unix()
					elapsed := end - start
					if elapsed < int64(collector.Device.CollectorCycle) {
						time.Sleep(time.Duration(int64(collector.Device.CollectorCycle)) * time.Second)
					}
				}
			}
		}()
	}
}

func (collector *ModbusRtuCollector) poll(ctx context.Context) bool {
	select {
	case <-collector.exitCh:
		return false
	default:
		sw := &sync.WaitGroup{}
		dfvCh := make(chan *modbusrturuntime.ParseVariableResult, 0)
		for _, DataFrames := range collector.FunctionCodeDataFrameMap {
			for _, frame := range DataFrames {
				sw.Add(1)
				go collector.message(ctx, frame, dfvCh, sw, collector.Clients)
			}
		}
		go collector.rollVariable(ctx, dfvCh)
		sw.Wait()
		close(dfvCh)
		return true
	}
}

func (collector *ModbusRtuCollector) message(ctx context.Context, dataFrame *ModBusRtuDataFrame, pvrCh chan<- *modbusrturuntime.ParseVariableResult, sw *sync.WaitGroup, clients *SerialClients) {
	defer sw.Done()
	defer func() {
		if err := recover(); err != nil {
			klog.V(2).InfoS("Failed to ask modbus tcp message", "error", err)
		}
	}()
	client, err := clients.getClient(ctx)
	if err != nil {
		klog.V(2).InfoS("Failed to get serial client", "error", err)
		pvrCh <- &modbusrturuntime.ParseVariableResult{Err: []error{err}}
		return
	}

	defer collector.Clients.releaseClient(client)
	var buf []byte
	if err := collector.retry(func(sc *SerialClient, dataFrame *ModBusRtuDataFrame) error {
		least, err := client.AskAtLeast(dataFrame.DataFrame, dataFrame.ResponseDataFrame)
		if err != nil {
			return modbusrturuntime.ErrBadConn
		}
		buf, err = dataFrame.ValidateMessage(least)
		if err != nil {
			return modbusrturuntime.ErrServerBadResp
		}
		return nil
	}, client, dataFrame); err != nil {
		klog.V(2).InfoS("Failed to connect modbus rtu server by retry three times")
		pvrCh <- &modbusrturuntime.ParseVariableResult{Err: []error{err}}
		return
	}
	// 解析数据
	// count代表数据字节长度
	count := int(buf[2])
	var bb []byte
	switch buf[1] {
	case 1, 2:
		// 数组解压
		bb = binutil.ExpandBool(buf[3:], count)
	case 3, 4, 23:
		bb = binutil.Dup(buf[3:])
	case 5, 15, 6, 16:
	default:
		klog.V(2).InfoS("Unsupported function code", "functionCode", buf[1])
	}

	pvrCh <- &modbusrturuntime.ParseVariableResult{Err: nil, VariableSlice: dataFrame.ParseVariableValue(bb)}
}

func (collector *ModbusRtuCollector) retry(fun func(sc *SerialClient, dataFrame *ModBusRtuDataFrame) error, sc *SerialClient, dataFrame *ModBusRtuDataFrame) error {
	for i := 0; i < 3; i++ {
		err := fun(sc, dataFrame)
		if err == nil {
			return nil
		} else if errors.Is(err, modbusruntime.ErrBadConn) {
			sc.Port.Close()
			newPort, err := collector.Clients.newSerialClient()
			if err != nil {
				return err
			}
			sc.Port = newPort.Port
			i = i - 1
		} else {
			klog.V(2).InfoS("Failed to connect modbus tcp server", "error", err)
		}
	}
	return modbusrturuntime.ErrManyRetry
}

func (collector *ModbusRtuCollector) rollVariable(ctx context.Context, ch chan *modbusrturuntime.ParseVariableResult) {
	rvs := make([]runtime.VariableValue, 0, collector.VariableCount)
	errs := make([]error, 0)
	for {
		select {
		case pvr, ok := <-ch:
			if !ok {
				collector.VariableCh <- &runtime.ParseVariableResult{Err: errs, VariableSlice: rvs}
				return
			} else if pvr.Err != nil {
				errs = append(errs, pvr.Err...)
			} else {
				for _, variable := range pvr.VariableSlice {
					rvs = append(rvs, variable)
				}
			}
		}
	}
}
