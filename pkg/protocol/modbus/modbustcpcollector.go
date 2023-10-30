package modbus

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	modbusruntime "harnsgateway/pkg/protocol/modbus/runtime"
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/utils/binutil"
	"k8s.io/klog/v2"
	"net"
	"sort"
	"sync"
	"time"
)

/**
modbus 协议 ADU = 地址(1) + pdu(253) + 16位校验(2) = 256
modbus tcp报文
tcp报文头(6)  +  地址(1)   +   pdu(253) = 260
*/

// ModBusDataFrame 报文对应的数据点位
type ModBusDataFrame struct {
	MemoryLayout      runtime.MemoryLayout
	StartAddress      uint
	FunctionCode      uint8
	MaxDataSize       uint // 最大数量01 代表线圈  03代表word
	TransactionId     uint16
	DataFrame         []byte
	ResponseDataFrame []byte
	Variables         []*VariableParse
}

func (df *ModBusDataFrame) GenerateReadMessage(slave uint, functionCode uint8, startAddress uint, maxDataSize uint) {
	df.TransactionId = 0
	df.FunctionCode = functionCode
	df.StartAddress = startAddress
	df.MaxDataSize = maxDataSize
	// 00 01 00 00 00 06 18 03 00 02 00 02
	// 00 01  此次通信事务处理标识符，一般每次通信之后将被要求加1以区别不同的通信数据报文
	// 00 00  表示协议标识符，00 00为modbus协议
	// 00 06  数据长度，用来指示接下来数据的长度，单位字节
	// 18  设备地址，用以标识连接在串行线或者网络上的远程服务端的地址。以上七个字节也被称为modbus报文头
	// 03  功能码，此时代码03为读取保持寄存器数据
	// 00 02  起始地址
	// 00 02  寄存器数量(word数量)/线圈数量
	message := make([]byte, 12)

	binutil.WriteUint16(message[2:], 0) // 协议版本
	binutil.WriteUint16(message[4:], 6) // 剩余长度
	message[6] = byte(slave)
	message[7] = byte(functionCode)
	binutil.WriteUint16(message[8:], uint16(startAddress))
	binutil.WriteUint16(message[10:], uint16(df.MaxDataSize))
	df.DataFrame = message
	df.ResponseDataFrame = make([]byte, 260)
}

func (df *ModBusDataFrame) WriteTransactionId() {
	df.TransactionId++
	id := df.TransactionId
	binutil.WriteUint16(df.DataFrame, id)
}

func (df *ModBusDataFrame) ValidateMessage(least int) ([]byte, error) {
	buf := df.ResponseDataFrame[:least]

	transactionId := binutil.ParseUint16(buf[:])
	if transactionId != df.TransactionId {
		klog.V(2).InfoS("Failed to match message transaction id", "request transactionId", df.TransactionId, "response transactionId", transactionId)
		return nil, modbusruntime.ErrMessageTransaction
	}

	length := binutil.ParseUint16(buf[4:])
	if int(length)+6 > least {
		klog.V(2).InfoS("Failed to get message enough length")
		return nil, modbusruntime.ErrMessageDataLengthNotEnough
	}

	functionCode := buf[7]
	if functionCode&0x80 > 0 {
		klog.V(2).InfoS("Failed to parse modbus tcp message", "error code", functionCode-128)
		return nil, modbusruntime.ErrMessageFunctionCodeError
	}
	return buf, nil
}

func (df *ModBusDataFrame) ParseVariableValue(data []byte) []*modbusruntime.Variable {
	vvs := make([]*modbusruntime.Variable, 0, len(df.Variables))
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

		vvs = append(vvs, &modbusruntime.Variable{
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
	Variable *modbusruntime.Variable
	Start    uint // 报文中数据[]byte开始位置
}

type ModbusTcpCollector struct {
	exitCh                   chan struct{}
	Device                   *modbusruntime.ModBusDevice
	Tunnels                  *Tunnels
	FunctionCodeDataFrameMap map[uint8][]*ModBusDataFrame
	VariableCount            int
	VariableCh               chan *runtime.ParseVariableResult
	CanCollect               bool
}

func NewCollector(d runtime.Device) (runtime.Collector, chan *runtime.ParseVariableResult, error) {
	device, ok := d.(*modbusruntime.ModBusDevice)
	if !ok {
		klog.V(2).InfoS("Failed to new modbus tcp collector,device type not supported")
		return nil, nil, modbusruntime.ErrDeviceType
	}
	VariableCount := 0
	CanCollect := false
	functionCodeDataFrameMap := make(map[uint8][]*ModBusDataFrame, 0)
	variables := device.Variables
	functionCodeVariableMap := make(map[uint8][]*modbusruntime.Variable, 0)
	for _, variable := range variables {
		functionCodeVariableMap[variable.FunctionCode] = append(functionCodeVariableMap[variable.FunctionCode], variable)
	}
	// functionCode03 modbus 一次最多读取123个寄存器，246个字节
	// functionCode01 modbus 一次最多读取246个字节 总共246 * 8 = 1968个线圈
	for code, variables := range functionCodeVariableMap {
		VariableCount = VariableCount + len(variables)
		sort.Sort(modbusruntime.VariableSlice(variables))
		dfs := make([]*ModBusDataFrame, 0)
		firstVariable := variables[0]
		startOffset := firstVariable.Address - device.PositionAddress
		startAddress := startOffset
		var maxDataSize uint = 0
		df := &ModBusDataFrame{MemoryLayout: device.MemoryLayout}
		switch code {
		case 1, 2:
			dataFrameDataLength := startAddress + 1967
			for i := 0; i < len(variables); i++ {
				variable := variables[i]
				if variable.Address <= dataFrameDataLength {
					vp := &VariableParse{
						Variable: variable,
						Start:    variable.Address - startAddress,
					}
					df.Variables = append(df.Variables, vp)
					maxDataSize = variable.Address - startAddress + 1
				} else {
					df.GenerateReadMessage(device.Slave, code, startAddress, maxDataSize)
					dfs = append(dfs, df)
					df = &ModBusDataFrame{MemoryLayout: device.MemoryLayout}
					maxDataSize = 0
					startAddress = variable.Address
					dataFrameDataLength = startAddress + 1967
					i--
				}
			}
		case 3, 4:
			dataFrameDataLength := startAddress + 122
			for i := 0; i < len(variables); i++ {
				variable := variables[i]
				if variable.Address+runtime.DataTypeWord[variable.DataType] <= dataFrameDataLength {
					vp := &VariableParse{
						Variable: variable,
						Start:    (variable.Address - startAddress) * 2,
					}
					df.Variables = append(df.Variables, vp)
					maxDataSize = variable.Address - startAddress + runtime.DataTypeWord[variable.DataType]
				} else {
					df.GenerateReadMessage(device.Slave, code, startAddress, maxDataSize)
					dfs = append(dfs, df)
					df = &ModBusDataFrame{MemoryLayout: device.MemoryLayout}
					maxDataSize = 0
					startAddress = variable.Address
					dataFrameDataLength = startAddress + 122
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

	tcpChannel := 0
	for _, values := range functionCodeDataFrameMap {
		tcpChannel += len(values)
	}
	if tcpChannel > 0 {
		tcpChannel = tcpChannel/5 + 1
		CanCollect = true
	}

	addr := fmt.Sprintf("%s:%d", device.Address, device.Port)
	ms := list.New()
	for i := 0; i < tcpChannel; i++ {
		tunnel, err := net.Dial("tcp", addr)
		if err != nil {
			klog.V(2).InfoS("Failed to connect modbus server", "error", err)
			return nil, nil, err
		}
		m := &Messenger{
			Tunnel:  tunnel,
			Timeout: 1,
		}
		ms.PushBack(m)
	}

	ts := &Tunnels{
		Messengers:   ms,
		Max:          tcpChannel,
		Idle:         tcpChannel,
		Mux:          &sync.Mutex{},
		NextRequest:  1,
		ConnRequests: make(map[uint64]chan *Messenger, 0),
		newMessenger: func() (*Messenger, error) {
			tunnel, err := net.Dial("tcp", addr)
			if err != nil {
				klog.V(2).InfoS("Failed to connect modbus server", "error", err)
				return nil, err
			}
			return &Messenger{
				Tunnel:  tunnel,
				Timeout: 1,
			}, nil
		},
	}

	mtc := &ModbusTcpCollector{
		Device:                   device,
		exitCh:                   make(chan struct{}, 0),
		FunctionCodeDataFrameMap: functionCodeDataFrameMap,
		Tunnels:                  ts,
		VariableCh:               make(chan *runtime.ParseVariableResult, 1),
		VariableCount:            VariableCount,
		CanCollect:               CanCollect,
	}
	return mtc, mtc.VariableCh, nil
}

func (collector *ModbusTcpCollector) Destroy(ctx context.Context) {
	collector.exitCh <- struct{}{}
	collector.Tunnels.Destroy(ctx)
	close(collector.VariableCh)
}

func (collector *ModbusTcpCollector) Collect(ctx context.Context) {
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

func (collector *ModbusTcpCollector) poll(ctx context.Context) bool {
	select {
	case <-collector.exitCh:
		return false
	default:
		sw := &sync.WaitGroup{}
		dfvCh := make(chan *modbusruntime.ParseVariableResult, 0)
		for _, DataFrames := range collector.FunctionCodeDataFrameMap {
			for _, frame := range DataFrames {
				sw.Add(1)
				go collector.message(ctx, frame, dfvCh, sw, collector.Tunnels)
			}
		}
		go collector.rollVariable(ctx, dfvCh)
		sw.Wait()
		close(dfvCh)
		return true
	}
}

func (collector *ModbusTcpCollector) message(ctx context.Context, dataFrame *ModBusDataFrame, pvrCh chan<- *modbusruntime.ParseVariableResult, sw *sync.WaitGroup, tunnels *Tunnels) {
	defer sw.Done()
	defer func() {
		if err := recover(); err != nil {
			klog.V(2).InfoS("Failed to ask modbus tcp message", "error", err)
		}
	}()
	tunnel, err := tunnels.getTunnel(ctx)
	defer collector.Tunnels.releaseTunnel(tunnel)
	if err != nil {
		klog.V(2).InfoS("Failed to get tunnel", "error", err)
		if tunnel, err = collector.Tunnels.newMessenger(); err != nil {
			return
		}
	}

	var buf []byte

	if err := collector.retry(func(tunnel *Messenger, dataFrame *ModBusDataFrame) error {
		dataFrame.WriteTransactionId()
		least, err := tunnel.AskAtLeast(dataFrame.DataFrame, dataFrame.ResponseDataFrame, 9)
		if err != nil {
			return modbusruntime.ErrBadConn
		}
		buf, err = dataFrame.ValidateMessage(least)
		if err != nil {
			return modbusruntime.ErrServerBadResp
		}
		return nil
	}, tunnel, dataFrame); err != nil {
		klog.V(2).InfoS("Failed to connect modbus server", "error", err)
		pvrCh <- &modbusruntime.ParseVariableResult{Err: []error{err}}
		return
	}
	// 解析数据
	// count代表数据字节长度
	count := int(buf[8])
	var bb []byte
	switch buf[7] {
	case 1, 2:
		// 数组解压
		bb = binutil.ExpandBool(buf[9:], count)
	case 3, 4, 23:
		bb = binutil.Dup(buf[9:])
	case 5, 15, 6, 16:
	default:
		klog.V(2).InfoS("Unsupported function code", "functionCode", buf[7])
	}

	pvrCh <- &modbusruntime.ParseVariableResult{Err: nil, VariableSlice: dataFrame.ParseVariableValue(bb)}
}

func (collector *ModbusTcpCollector) retry(fun func(tunnel *Messenger, dataFrame *ModBusDataFrame) error, tunnel *Messenger, dataFrame *ModBusDataFrame) error {
	for i := 0; i < 3; i++ {
		err := fun(tunnel, dataFrame)
		if err == nil {
			return nil
		} else if errors.Is(err, modbusruntime.ErrBadConn) {
			tunnel.Tunnel.Close()
			newTunnel, err := collector.Tunnels.newMessenger()
			if err != nil {
				return err
			}
			tunnel.Tunnel = newTunnel.Tunnel
			i = i - 1
		} else {
			klog.V(2).InfoS("Failed to connect modbus tcp server", "error", err)
		}
	}
	return modbusruntime.ErrManyRetry
}

func (collector *ModbusTcpCollector) rollVariable(ctx context.Context, ch chan *modbusruntime.ParseVariableResult) {
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
