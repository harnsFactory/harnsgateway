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
		needCheckTransaction = true
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
		case modbus.ReadCoilStatus, modbus.WriteCoilStatus:
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
		case modbus.ReadHoldRegister, modbus.WriteHoldRegister:
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
	action := make(map[string]interface{}, len(obj))
	for name, v := range obj {
		vv, _ := variablesMap[name]
		variableValue := vv.(*modbus.Variable)
		switch variableValue.DataType {
		case runtime.BOOL:
			switch v.(type) {
			case bool:
				action[name] = v
			case string:
				b, err := strconv.ParseBool(v.(string))
				if err == nil {
					action[name] = b
				} else {
					return response.ErrBooleanInvalid(name)
				}
			default:
				return response.ErrBooleanInvalid(name)
			}
		case runtime.INT16:
			switch v.(type) {
			case float64:
				i := uint16(v.(float64))
				action[name] = i
			default:
				return response.ErrInteger16Invalid(name)
			}
		case runtime.UINT16:
			switch v.(type) {
			case float64:
				i := int16(v.(float64))
				action[name] = i
			default:
				return response.ErrInteger16Invalid(name)
			}
		case runtime.INT32:
			switch v.(type) {
			case float64:
				i := int32(v.(float64))
				action[name] = i
			default:
				return response.ErrInteger32Invalid(name)
			}
		case runtime.INT64:
			switch v.(type) {
			case float64:
				i := int64(v.(float64))
				action[name] = i
			default:
				return response.ErrInteger64Invalid(name)
			}
		case runtime.FLOAT32:
			switch v.(type) {
			case float64:
				i := float32(v.(float64))
				action[name] = i
			default:
				return response.ErrFloat32Invalid(name)
			}
		case runtime.FLOAT64:
			switch v.(type) {
			case float64:
				i := v.(float64)
				action[name] = i
			default:
				return response.ErrFloat64Invalid(name)
			}
		default:
			klog.V(3).InfoS("Unsupported dataType", "dataType", variableValue.DataType)
		}
	}
	// TODO implement me
	panic("implement me")
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
		_, err := messenger.AskAtLeast(dataFrame.DataFrame, dataFrame.ResponseDataFrame, 9)
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
			i = i - 1
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
	switch buf[1] {
	case 1, 2:
		// 数组解压
		bb = binutil.ExpandBool(buf[3:], int(byteDataLength))
	case 3, 4, 23:
		bb = binutil.Dup(buf[3:])
	case 5, 15, 6, 16:
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
