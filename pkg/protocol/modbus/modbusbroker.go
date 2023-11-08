package modbus

import (
	"context"
	"errors"
	"harnsgateway/pkg/apis/response"
	"harnsgateway/pkg/protocol/modbus/model"
	modbus "harnsgateway/pkg/protocol/modbus/runtime"
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/utils/binutil"
	"harnsgateway/pkg/utils/crcutil"
	"k8s.io/klog/v2"
	"sort"
	"strconv"
	"sync"
	"time"
)

/**
modbus 协议 ADU = 地址(1) + pdu(253) + 16位校验(2) = 256
modbus tcp报文
tcp报文头(6)  +  地址(1)   +   pdu(253) = 260
modbus rtu报文
地址(1) + pdu(253) + 16位校验(2) = 256
modbus rtu over tcp
tcp报文头(6)  +  地址(1)   +   pdu(253)   +  16位校验(2)  = 262
*/

// ModBusDataFrame 报文对应的数据点位
var _ runtime.Broker = (*ModbusBroker)(nil)

type ModbusBroker struct {
	NeedCheckTransaction     bool
	NeedCheckCrc16Sum        bool
	ExitCh                   chan struct{}
	Device                   *modbus.ModBusDevice
	Clients                  *modbus.Clients
	FunctionCodeDataFrameMap map[uint8][]*modbus.ModBusDataFrame
	VariableCount            int
	VariableCh               chan *runtime.ParseVariableResult
	CanCollect               bool
}

func NewBroker(d runtime.Device) (runtime.Broker, chan *runtime.ParseVariableResult, error) {
	device, ok := d.(*modbus.ModBusDevice)
	if !ok {
		klog.V(2).InfoS("Failed to new modbus tcp device,device type not supported")
		return nil, nil, modbus.ErrDeviceType
	}

	needCheckTransaction := false
	needCheckCrc16Sum := false
	switch modbus.StringToModbusModel[device.DeviceModel] {
	case modbus.Tcp:
		needCheckTransaction = true
	case modbus.Rtu:
		needCheckCrc16Sum = true
	case modbus.RtuOverTcp:
		needCheckCrc16Sum = true
	}

	VariableCount := 0
	CanCollect := false
	functionCodeDataFrameMap := make(map[uint8][]*modbus.ModBusDataFrame, 0)
	functionCodeVariableMap := make(map[uint8][]*modbus.Variable, 0)
	for _, variable := range device.Variables {
		functionCodeVariableMap[variable.FunctionCode] = append(functionCodeVariableMap[variable.FunctionCode], variable)
	}
	for code, variables := range functionCodeVariableMap {
		VariableCount = VariableCount + len(variables)
		sort.Sort(modbus.VariableSlice(variables))
		dfs := make([]*modbus.ModBusDataFrame, 0)
		firstVariable := variables[0]
		startOffset := firstVariable.Address - device.PositionAddress
		startAddress := startOffset
		var maxDataSize uint = 0
		vps := make([]*modbus.VariableParse, 0)
		switch modbus.FunctionCode(code) {
		case modbus.ReadCoilStatus, modbus.ReadInputStatus:
			dataFrameDataLength := startAddress + modbus.PerRequestMaxCoil
			for i := 0; i < len(variables); i++ {
				variable := variables[i]
				if variable.Address <= dataFrameDataLength {
					vp := &modbus.VariableParse{
						Variable: variable,
						Start:    variable.Address - startAddress,
					}
					vps = append(vps, vp)
					maxDataSize = variable.Address - startAddress + 1
				} else {
					df := model.ModbusModelers[device.DeviceModel].GenerateReadMessage(device.Slave, code, startAddress, maxDataSize, vps, device.MemoryLayout)
					dfs = append(dfs, df)
					vps = vps[:0:0]
					maxDataSize = 0
					startAddress = variable.Address
					dataFrameDataLength = startAddress + modbus.PerRequestMaxCoil
					i--
				}
			}
		case modbus.ReadHoldRegister, modbus.ReadInputRegister:
			dataFrameDataLength := startAddress + modbus.PerRequestMaxRegister
			for i := 0; i < len(variables); i++ {
				variable := variables[i]
				if variable.Address+runtime.DataTypeWord[variable.DataType] <= dataFrameDataLength {
					vp := &modbus.VariableParse{
						Variable: variable,
						Start:    (variable.Address - startAddress) * 2,
					}
					vps = append(vps, vp)
					maxDataSize = variable.Address - startAddress + runtime.DataTypeWord[variable.DataType]
				} else {
					df := model.ModbusModelers[device.DeviceModel].GenerateReadMessage(device.Slave, code, startAddress, maxDataSize, vps, device.MemoryLayout)
					dfs = append(dfs, df)
					vps = vps[:0:0]
					maxDataSize = 0
					startAddress = variable.Address
					dataFrameDataLength = startAddress + modbus.PerRequestMaxRegister
					i--
				}
			}
		}
		if len(vps) > 0 {
			df := model.ModbusModelers[device.DeviceModel].GenerateReadMessage(device.Slave, code, startAddress, maxDataSize, vps, device.MemoryLayout)
			dfs = append(dfs, df)
			vps = vps[:0:0]
		}
		functionCodeDataFrameMap[code] = append(functionCodeDataFrameMap[code], dfs...)
	}

	dataFrameCount := 0
	for _, values := range functionCodeDataFrameMap {
		dataFrameCount += len(values)
	}
	if dataFrameCount == 0 {
		klog.V(2).InfoS("Failed to collect from Modbus server.Because of the variables is empty", "deviceId", device.ID)
		return nil, nil, nil
	}

	CanCollect = true
	clients, err := model.ModbusModelers[device.DeviceModel].NewClients(device.Address, dataFrameCount)
	if err != nil {
		klog.V(2).InfoS("Failed to collect from Modbus server", "error", err, "deviceId", device.ID)
		return nil, nil, nil
	}

	mtc := &ModbusBroker{
		Device:                   device,
		ExitCh:                   make(chan struct{}, 0),
		FunctionCodeDataFrameMap: functionCodeDataFrameMap,
		Clients:                  clients,
		VariableCh:               make(chan *runtime.ParseVariableResult, 1),
		VariableCount:            VariableCount,
		CanCollect:               CanCollect,
		NeedCheckCrc16Sum:        needCheckCrc16Sum,
		NeedCheckTransaction:     needCheckTransaction,
	}
	return mtc, mtc.VariableCh, nil
}

func (broker *ModbusBroker) Destroy(ctx context.Context) {
	broker.ExitCh <- struct{}{}
	broker.Clients.Destroy(ctx)
	close(broker.VariableCh)
}

func (broker *ModbusBroker) Collect(ctx context.Context) {
	if broker.CanCollect {
		go func() {
			for {
				start := time.Now().Unix()
				if !broker.poll(ctx) {
					return
				}
				select {
				case <-broker.ExitCh:
					return
				default:
					end := time.Now().Unix()
					elapsed := end - start
					if elapsed < int64(broker.Device.CollectorCycle) {
						time.Sleep(time.Duration(int64(broker.Device.CollectorCycle)) * time.Second)
					}
				}
			}
		}()
	}
}

func (broker *ModbusBroker) DeliverAction(ctx context.Context, obj map[string]interface{}) error {
	variablesMap := broker.Device.GetVariablesMap()
	action := make([]*modbus.Variable, 0, len(obj))

	for name, value := range obj {
		vv, _ := variablesMap[name]
		variableValue := vv.(*modbus.Variable)

		v := &modbus.Variable{
			DataType:     variableValue.DataType,
			Name:         variableValue.Name,
			Address:      variableValue.Address,
			Bits:         variableValue.Bits,
			FunctionCode: variableValue.FunctionCode,
			Rate:         variableValue.Rate,
			Amount:       variableValue.Amount,
		}
		switch variableValue.DataType {
		case runtime.BOOL:
			switch value.(type) {
			case bool:
				v.Value = value
			case string:
				b, err := strconv.ParseBool(value.(string))
				if err == nil {
					v.Value = b
				} else {
					return response.ErrBooleanInvalid(name)
				}
			default:
				return response.ErrBooleanInvalid(name)
			}
		case runtime.INT16:
			switch value.(type) {
			case float64:
				v.Value = int16(value.(float64))
			default:
				return response.ErrInteger16Invalid(name)
			}
		case runtime.UINT16:
			switch value.(type) {
			case float64:
				v.Value = uint16(value.(float64))
			default:
				return response.ErrInteger16Invalid(name)
			}
		case runtime.INT32:
			switch value.(type) {
			case float64:
				v.Value = int32(value.(float64))
			default:
				return response.ErrInteger32Invalid(name)
			}
		case runtime.INT64:
			switch value.(type) {
			case float64:
				v.Value = int64(value.(float64))
			default:
				return response.ErrInteger64Invalid(name)
			}
		case runtime.FLOAT32:
			switch value.(type) {
			case float64:
				v.Value = float32(value.(float64))
			default:
				return response.ErrFloat32Invalid(name)
			}
		case runtime.FLOAT64:
			switch value.(type) {
			case float64:
				v.Value = value.(float64)
			default:
				return response.ErrFloat64Invalid(name)
			}
		default:
			klog.V(3).InfoS("Unsupported dataType", "dataType", variableValue.DataType)
		}
		action = append(action, v)
	}

	dataBytes := broker.generateActionBytes(broker.Device.MemoryLayout, action)
	dataFrames := make([][]byte, 0, len(dataBytes))
	for i, dbs := range dataBytes {
		var bytes []byte
		if broker.NeedCheckTransaction {
			bytes = append(bytes, make([]byte, 6)...)
			binutil.WriteUint16BigEndian(bytes[0:], uint16(i))
			binutil.WriteUint16BigEndian(bytes[2:], 0)
			binutil.WriteUint16BigEndian(bytes[4:], uint16(1+len(dbs)))
		}
		bytes = append(bytes, byte(broker.Device.Slave))
		bytes = append(bytes, dbs...)
		if broker.NeedCheckCrc16Sum {

		}
		dataFrames = append(dataFrames, bytes)
	}

	messenger, err := broker.Clients.GetMessenger(ctx)
	if err != nil {
		klog.V(2).InfoS("Failed to get messenger", "error", err)
		if messenger, err = broker.Clients.NewMessenger(); err != nil {
			return err
		}
	}
	defer broker.Clients.ReleaseMessenger(messenger)

	errs := &response.MultiError{}
	for _, frame := range dataFrames {
		rp := make([]byte, len(frame))
		_, err = messenger.AskAtLeast(frame, rp, 9)
		if err != nil {
			errs.Add(modbus.ErrModbusBadConn)
			continue
		}
		if broker.NeedCheckTransaction {
			transactionId := binutil.ParseUint16(rp[:])
			requestTransactionId := binutil.ParseUint16(rp[:])
			if transactionId != requestTransactionId {
				klog.V(2).InfoS("Failed to match message transaction id", "request transactionId", requestTransactionId, "response transactionId", transactionId)
				errs.Add(modbus.ErrMessageTransaction)
				continue
			}
			rp = rp[6:]
		}

		slave := rp[0]
		if uint(slave) != broker.Device.Slave {
			klog.V(2).InfoS("Failed to match modbus slave", "request slave", broker.Device.Slave, "response slave", slave)
			errs.Add(modbus.ErrMessageSlave)
			continue
		}
		functionCode := rp[1]
		if functionCode&0x80 > 0 {
			klog.V(2).InfoS("Failed to parse modbus tcp message", "error code", functionCode-128)
			errs.Add(modbus.ErrMessageFunctionCodeError)
			continue
		}
	}

	if errs.Len() > 0 {
		return errs
	}

	return nil
}

func (broker *ModbusBroker) poll(ctx context.Context) bool {
	select {
	case <-broker.ExitCh:
		return false
	default:
		sw := &sync.WaitGroup{}
		dfvCh := make(chan *modbus.ParseVariableResult, 0)
		for _, DataFrames := range broker.FunctionCodeDataFrameMap {
			for _, frame := range DataFrames {
				sw.Add(1)
				go broker.message(ctx, frame, dfvCh, sw, broker.Clients)
			}
		}
		go broker.rollVariable(ctx, dfvCh)
		sw.Wait()
		close(dfvCh)
		return true
	}
}

func (broker *ModbusBroker) message(ctx context.Context, dataFrame *modbus.ModBusDataFrame, pvrCh chan<- *modbus.ParseVariableResult, sw *sync.WaitGroup, clients *modbus.Clients) {
	defer sw.Done()
	defer func() {
		if err := recover(); err != nil {
			klog.V(2).InfoS("Failed to ask Modbus server message", "error", err)
		}
	}()
	messenger, err := clients.GetMessenger(ctx)
	defer broker.Clients.ReleaseMessenger(messenger)
	if err != nil {
		klog.V(2).InfoS("Failed to get messenger", "error", err)
		if messenger, err = broker.Clients.NewMessenger(); err != nil {
			return
		}
	}

	var buf []byte

	if err := broker.retry(func(messenger modbus.Messenger, dataFrame *modbus.ModBusDataFrame) error {
		if broker.NeedCheckTransaction {
			dataFrame.WriteTransactionId()
		}
		_, err := messenger.AskAtLeast(dataFrame.DataFrame, dataFrame.ResponseDataFrame, 4)
		if err != nil {
			return modbus.ErrModbusBadConn
		}
		buf, err = broker.ValidateAndExtractMessage(dataFrame)
		if err != nil {
			return modbus.ErrModbusServerBadResp
		}
		return nil
	}, messenger, dataFrame); err != nil {
		klog.V(2).InfoS("Failed to connect modbus server", "error", err)
		pvrCh <- &modbus.ParseVariableResult{Err: []error{err}}
		return
	}

	pvrCh <- &modbus.ParseVariableResult{Err: nil, VariableSlice: dataFrame.ParseVariableValue(buf)}
}

func (broker *ModbusBroker) retry(fun func(messenger modbus.Messenger, dataFrame *modbus.ModBusDataFrame) error, messenger modbus.Messenger, dataFrame *modbus.ModBusDataFrame) error {
	for i := 0; i < 3; i++ {
		err := fun(messenger, dataFrame)
		if err == nil {
			return nil
		} else if errors.Is(err, modbus.ErrModbusBadConn) {
			messenger.Close()
			newMessenger, err := broker.Clients.NewMessenger()
			if err != nil {
				return err
			}
			messenger.Reset(newMessenger)
		} else {
			klog.V(2).InfoS("Failed to connect Modbus server", "error", err)
		}
	}
	return modbus.ErrManyRetry
}

func (broker *ModbusBroker) ValidateAndExtractMessage(df *modbus.ModBusDataFrame) ([]byte, error) {
	buf := df.ResponseDataFrame[:]

	if broker.NeedCheckTransaction {
		transactionId := binutil.ParseUint16(buf[:])
		if transactionId != df.TransactionId {
			klog.V(2).InfoS("Failed to match message transaction id", "request transactionId", df.TransactionId, "response transactionId", transactionId)
			return nil, modbus.ErrMessageTransaction
		}
		buf = buf[6:]
	}

	slave := buf[0]
	if uint(slave) != df.Slave {
		klog.V(2).InfoS("Failed to match modbus slave", "request slave", df.Slave, "response slave", slave)
		return nil, modbus.ErrMessageSlave
	}
	functionCode := buf[1]
	if functionCode&0x80 > 0 {
		klog.V(2).InfoS("Failed to parse modbus tcp message", "error code", functionCode-128)
		return nil, modbus.ErrMessageFunctionCodeError
	}

	byteDataLength := buf[2]
	if broker.NeedCheckCrc16Sum {
		if int(byteDataLength)+5 != len(buf) {
			klog.V(2).InfoS("Failed to get message enough length")
			return nil, modbus.ErrMessageDataLengthNotEnough
		}
		checkBufData := buf[:byteDataLength+3]
		sum := crcutil.CheckCrc16sum(checkBufData)
		crc := binutil.ParseUint16BigEndian(buf[byteDataLength+3 : byteDataLength+5])
		if sum != crc {
			klog.V(2).InfoS("Failed to check CRC16")
			return nil, modbus.ErrCRC16Error
		}
	} else {
		if int(byteDataLength)+3 != len(buf) {
			klog.V(2).InfoS("Failed to get message enough length")
			return nil, modbus.ErrMessageDataLengthNotEnough
		}
	}

	var bb []byte
	switch modbus.FunctionCode(buf[1]) {
	case modbus.ReadCoilStatus, modbus.ReadInputStatus:
		// 数组解压
		bb = binutil.ExpandBool(buf[3:], int(byteDataLength))
	case modbus.ReadHoldRegister, modbus.ReadInputRegister:
		bb = binutil.Dup(buf[3:])
	case modbus.WriteSingleCoil, modbus.WriteSingleRegister, modbus.WriteMultipleCoil, modbus.WriteMultipleRegister:
	default:
		klog.V(2).InfoS("Unsupported function code", "functionCode", buf[1])
	}

	return bb, nil
}

func (broker *ModbusBroker) rollVariable(ctx context.Context, ch chan *modbus.ParseVariableResult) {
	rvs := make([]runtime.VariableValue, 0, broker.VariableCount)
	errs := make([]error, 0)
	for {
		select {
		case pvr, ok := <-ch:
			if !ok {
				broker.VariableCh <- &runtime.ParseVariableResult{Err: errs, VariableSlice: rvs}
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

func (broker *ModbusBroker) generateActionBytes(memoryLayout runtime.MemoryLayout, action []*modbus.Variable) [][]byte {
	dataBytes := make([][]byte, 0, len(action))

	for _, variable := range action {
		// functioncode + startAddress
		pduByte := make([]byte, 3)
		var fc byte
		dataByte := make([]byte, 0)
		switch modbus.FunctionCode(variable.FunctionCode) {
		case modbus.ReadCoilStatus, modbus.ReadInputStatus:
			switch variable.DataType {
			// 65280
			case runtime.BOOL:
				fc = byte(modbus.WriteSingleCoil)
				if variable.Value.(bool) {
					dataByte = append(dataByte, binutil.Uint16ToBytesBigEndian(uint16(65280))...)
				} else {
					dataByte = append(dataByte, binutil.Uint16ToBytesBigEndian(uint16(0))...)
				}
			case runtime.INT16:
				fc = byte(modbus.WriteSingleCoil)
				if variable.Value.(int16) > 0 {
					dataByte = append(dataByte, binutil.Uint16ToBytesBigEndian(uint16(65280))...)
				} else {
					dataByte = append(dataByte, binutil.Uint16ToBytesBigEndian(uint16(0))...)
				}
			case runtime.UINT16:
				fc = byte(modbus.WriteSingleCoil)
				if variable.Value.(uint16) > 0 {
					dataByte = append(dataByte, binutil.Uint16ToBytesBigEndian(uint16(65280))...)
				} else {
					dataByte = append(dataByte, binutil.Uint16ToBytesBigEndian(uint16(0))...)
				}
			case runtime.INT32:
				fc = byte(modbus.WriteSingleCoil)
				if variable.Value.(int32) > 0 {
					dataByte = append(dataByte, binutil.Uint16ToBytesBigEndian(uint16(65280))...)
				} else {
					dataByte = append(dataByte, binutil.Uint16ToBytesBigEndian(uint16(0))...)
				}
			case runtime.INT64:
				fc = byte(modbus.WriteSingleCoil)
				if variable.Value.(int64) > 0 {
					dataByte = append(dataByte, binutil.Uint16ToBytesBigEndian(uint16(65280))...)
				} else {
					dataByte = append(dataByte, binutil.Uint16ToBytesBigEndian(uint16(0))...)
				}
			case runtime.FLOAT32:
				fc = byte(modbus.WriteSingleCoil)
				if variable.Value.(float32) > 0 {
					dataByte = append(dataByte, binutil.Uint16ToBytesBigEndian(uint16(65280))...)
				} else {
					dataByte = append(dataByte, binutil.Uint16ToBytesBigEndian(uint16(0))...)
				}
			case runtime.FLOAT64:
				fc = byte(modbus.WriteSingleCoil)
				if variable.Value.(float64) > 0 {
					dataByte = append(dataByte, binutil.Uint16ToBytesBigEndian(uint16(65280))...)
				} else {
					dataByte = append(dataByte, binutil.Uint16ToBytesBigEndian(uint16(0))...)
				}
			}
		case modbus.ReadHoldRegister, modbus.ReadInputRegister:
			switch variable.DataType {
			case runtime.BOOL:
				klog.V(2).InfoS("Unsupported bool variable with read hold register", "variableName", variable.Name)
				// todo 需要先读取数据
			case runtime.INT16:
				fc = byte(modbus.WriteSingleRegister)

				var value int16
				if variable.Rate != 0 && variable.Rate != 1 {
					value = int16((variable.Value.(float64)) * variable.Rate)
				} else {
					value = variable.Value.(int16)
				}
				switch memoryLayout {
				case runtime.ABCD, runtime.CDAB:
					dataByte = append(dataByte, binutil.Uint16ToBytesBigEndian(uint16(value))...)
				case runtime.BADC, runtime.DCBA:
					dataByte = append(dataByte, binutil.Uint16ToBytesLittleEndian(uint16(value))...)
				}
			case runtime.UINT16:
				fc = byte(modbus.WriteSingleRegister)

				var value uint16
				if variable.Rate != 0 && variable.Rate != 1 {
					value = uint16((variable.Value.(float64)) * variable.Rate)
				} else {
					value = variable.Value.(uint16)
				}
				switch memoryLayout {
				case runtime.ABCD, runtime.CDAB:
					dataByte = append(dataByte, binutil.Uint16ToBytesBigEndian(value)...)
				case runtime.BADC, runtime.DCBA:
					dataByte = append(dataByte, binutil.Uint16ToBytesLittleEndian(value)...)
				}
			case runtime.INT32:
				fc = byte(modbus.WriteMultipleRegister)
				registerAmount := 2
				dataByte = append(dataByte, binutil.Uint16ToBytesBigEndian(uint16(registerAmount))...)
				dataByte = append(dataByte, byte(2*registerAmount))

				var value int32
				if variable.Rate != 0 && variable.Rate != 1 {
					value = int32((variable.Value.(float64)) * variable.Rate)
				} else {
					value = variable.Value.(int32)
				}
				switch memoryLayout {
				case runtime.ABCD:
					dataByte = append(dataByte, binutil.Uint32ToBytesBigEndian(uint32(value))...)
				case runtime.BADC:
					// 大端交换
					dataByte = append(dataByte, binutil.Uint32ToBytesBigEndianByteSwap(uint32(value))...)
				case runtime.CDAB:
					dataByte = append(dataByte, binutil.Uint32ToBytesLittleEndianByteSwap(uint32(value))...)
				case runtime.DCBA:
					dataByte = append(dataByte, binutil.Uint32ToBytesLittleEndian(uint32(value))...)
				}
			case runtime.INT64:
				fc = byte(modbus.WriteMultipleRegister)
				registerAmount := 4
				dataByte = append(dataByte, binutil.Uint16ToBytesBigEndian(uint16(registerAmount))...)
				dataByte = append(dataByte, byte(2*registerAmount))

				var value int64
				if variable.Rate != 0 && variable.Rate != 1 {
					value = int64((variable.Value.(float64)) * variable.Rate)
				} else {
					value = variable.Value.(int64)
				}
				switch memoryLayout {
				case runtime.ABCD:
					dataByte = append(dataByte, binutil.Uint64ToBytesBigEndian(uint64(value))...)
				case runtime.BADC:
					// 大端交换
					dataByte = append(dataByte, binutil.Uint64ToBytesBigEndianByteSwap(uint64(value))...)
				case runtime.CDAB:
					dataByte = append(dataByte, binutil.Uint64ToBytesLittleEndianByteSwap(uint64(value))...)
				case runtime.DCBA:
					dataByte = append(dataByte, binutil.Uint64ToBytesLittleEndian(uint64(value))...)
				}
			case runtime.FLOAT32:
				fc = byte(modbus.WriteMultipleRegister)
				registerAmount := 2
				dataByte = append(dataByte, binutil.Uint16ToBytesBigEndian(uint16(registerAmount))...)
				dataByte = append(dataByte, byte(2*registerAmount))

				var value float32
				if variable.Rate != 0 && variable.Rate != 1 {
					value = float32((variable.Value.(float64)) * variable.Rate)
				} else {
					value = variable.Value.(float32)
				}
				switch memoryLayout {
				case runtime.ABCD:
					dataByte = append(dataByte, binutil.Float32ToBytesBigEndian(value)...)
				case runtime.BADC:
					// 大端交换
					dataByte = append(dataByte, binutil.Float32ToBytesBigEndianByteSwap(value)...)
				case runtime.CDAB:
					dataByte = append(dataByte, binutil.Float32ToBytesLittleEndianByteSwap(value)...)
				case runtime.DCBA:
					dataByte = append(dataByte, binutil.Float32ToBytesLittleEndian(value)...)
				}
			case runtime.FLOAT64:
				fc = byte(modbus.WriteMultipleRegister)
				registerAmount := 4
				dataByte = append(dataByte, binutil.Uint16ToBytesBigEndian(uint16(registerAmount))...)
				dataByte = append(dataByte, byte(2*registerAmount))

				var value float64
				if variable.Rate != 0 && variable.Rate != 1 {
					value = (variable.Value.(float64)) * variable.Rate
				} else {
					value = variable.Value.(float64)
				}
				switch memoryLayout {
				case runtime.ABCD:
					dataByte = append(dataByte, binutil.Float64ToBytesBigEndian(value)...)
				case runtime.BADC:
					// 大端交换
					dataByte = append(dataByte, binutil.Float64ToBytesBigEndianByteSwap(value)...)
				case runtime.CDAB:
					dataByte = append(dataByte, binutil.Float64ToBytesLittleEndianByteSwap(value)...)
				case runtime.DCBA:
					dataByte = append(dataByte, binutil.Float64ToBytesLittleEndian(value)...)
				}
			}
		}
		pduByte[0] = fc
		binutil.WriteUint16BigEndian(pduByte[1:], uint16(variable.Address))
		pduByte = append(pduByte, dataByte...)

		dataBytes = append(dataBytes, pduByte)
	}
	return dataBytes
}
