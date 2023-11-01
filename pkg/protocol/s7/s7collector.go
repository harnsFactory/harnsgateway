package s7

import (
	"context"
	"errors"
	"fmt"
	"harnsgateway/pkg/protocol/s7/model"
	s7runtime "harnsgateway/pkg/protocol/s7/runtime"
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/utils/binutil"
	"k8s.io/klog/v2"
	"sort"
	"sync"
	"time"
)

type S7Item struct {
	RequestData  []byte
	StartAddress uint // 响应报文中的item位置
}

type S7DataFrame struct {
	Zone              s7runtime.S7StoreArea
	DataFrame         []byte
	ItemCount         uint8
	DataLength        int
	ResponseDataFrame []byte
	Variables         []*VariableParse
}

func (df *S7DataFrame) ValidateMessage(least int) ([]byte, error) {
	buf := df.ResponseDataFrame[:least]

	itemLength := binutil.ParseUint16(buf[15:])
	if itemLength != uint16(df.DataLength) {
		klog.V(2).InfoS("Failed to get message enough length")
		return nil, s7runtime.ErrMessageDataLengthNotEnough
	}
	if uint8(buf[17]) != 0 {
		klog.V(2).InfoS("Failed to get s7 message")
		return nil, s7runtime.ErrMessageS7Response
	}
	if uint8(buf[18]) != 0 {
		klog.V(2).InfoS("Failed to get s7 message")
		return nil, s7runtime.ErrMessageS7Response
	}
	return buf, nil
}

func (df *S7DataFrame) ParseVariableValue(data []byte) s7runtime.VariableSlice {
	vvs := make([]*s7runtime.Variable, 0, len(df.Variables))
	for _, vp := range df.Variables {
		var value interface{}
		switch vp.Variable.DataType {
		case runtime.BOOL:
			// startAddress 为item start的索引位置
			v := data[vp.StartAddress+4]
			value = 1<<(vp.BitAddressOrLength)&v != 0
		case runtime.STRING:
			// todo
		case runtime.UINT16:
			var v interface{}
			vpData := data[vp.StartAddress+4:]
			v = binutil.ParseUint16BigEndian(vpData)
			if vp.Variable.Rate != 0 && vp.Variable.Rate != 1 {
				value = uint16((v.(float64)) * vp.Variable.Rate)
			} else {
				value = v
			}
		case runtime.INT16:
			var v interface{}
			vpData := data[vp.StartAddress+4:]
			v = int16(binutil.ParseUint16BigEndian(vpData))
			if vp.Variable.Rate != 0 && vp.Variable.Rate != 1 {
				value = int16((v.(float64)) * vp.Variable.Rate)
			} else {
				value = v
			}
		case runtime.INT32:
			var v interface{}
			vpData := data[vp.StartAddress+4:]
			v = int32(binutil.ParseUint32BigEndian(vpData))
			if vp.Variable.Rate != 0 && vp.Variable.Rate != 1 {
				value = int32((v.(float64)) * vp.Variable.Rate)
			} else {
				value = v
			}
		case runtime.FLOAT32:
			var v interface{}
			vpData := data[vp.StartAddress+4:]
			v = binutil.ParseFloat32BigEndian(vpData)
			if vp.Variable.Rate != 0 && vp.Variable.Rate != 1 {
				value = float32((v.(float64)) * vp.Variable.Rate)
			} else {
				value = v
			}
		case runtime.INT64:
			var v interface{}
			vpData := data[vp.StartAddress+4:]
			v = int64(binutil.ParseUint64BigEndian(vpData))
			if vp.Variable.Rate != 0 && vp.Variable.Rate != 1 {
				value = int64((v.(float64)) * vp.Variable.Rate)
			} else {
				value = v
			}
		case runtime.FLOAT64:
			var v interface{}
			vpData := data[vp.StartAddress+4:]
			v = binutil.ParseFloat64BigEndian(vpData)
			if vp.Variable.Rate != 0 && vp.Variable.Rate != 1 {
				value = (v.(float64)) * vp.Variable.Rate
			} else {
				value = v
			}
		}

		vvs = append(vvs, &s7runtime.Variable{
			DataType:     vp.Variable.DataType,
			Name:         vp.Variable.Name,
			Address:      vp.Variable.Address,
			Rate:         vp.Variable.Rate,
			DefaultValue: vp.Variable.DefaultValue,
			Value:        value,
		})
	}
	return vvs
}

type VariableParse struct {
	Variable           *s7runtime.Variable
	StartAddress       uint
	BitAddressOrLength uint8
	BlockSize          uint
}

type S7Collector struct {
	ExitCh                   chan struct{}
	Device                   *s7runtime.S7Device
	Clients                  *s7runtime.Clients
	StoreAddressDataFrameMap map[s7runtime.S7StoreArea][]*S7DataFrame
	VariableCount            int
	VariableCh               chan *runtime.ParseVariableResult
	CanCollect               bool
	Endpoint                 string
}

func NewCollector(d runtime.Device) (runtime.Collector, chan *runtime.ParseVariableResult, error) {
	device, ok := d.(*s7runtime.S7Device)
	if !ok {
		klog.V(2).InfoS("Failed to new s7 broker,device type not supported")
		return nil, nil, s7runtime.ErrDeviceType
	}
	maxPduLength, err := model.S7Modelers[device.DeviceModel].GetS7DevicePDULength(device.Address)
	if err != nil {
		klog.V(2).InfoS("Failed to connect s7 device for get pdu length")
		return nil, nil, err
	}
	CanCollect := false

	variableCount := 0
	storeAddressVariableMap := make(map[s7runtime.S7StoreArea][]*s7runtime.Variable)
	for _, variable := range device.Variables {
		zone := variable.Zone()
		storeAddressVariableMap[zone] = append(storeAddressVariableMap[zone], variable)
	}

	storeAddressDataFrameMap := make(map[s7runtime.S7StoreArea][]*S7DataFrame)
	for key, variables := range storeAddressVariableMap {
		variableCount = variableCount + len(variables)
		sort.Sort(s7runtime.VariableSlice(variables))
		maxPdu := maxPduLength
		dataFrames := make([]*S7DataFrame, 0)
		// 请求报文19个字节 + item 每个item12个字节 返回的报文 cotp占19个字节 header2个字节
		maxItemPerDataFrame := (maxPdu - 19) / 12
		itemMap := make(map[string]*S7Item, 0)
		items := make([]*S7Item, 0)
		variableParses := make([]*VariableParse, 0)
		startAddressOffset := 0
		for _, variable := range variables {
			zone, blockSize, startAddress, bitAddress := variable.ParseVariableAddress()
			offset := int(4 + variable.DataResponseLength(key))
			addressKey := fmt.Sprintf("%s.%d.%d", s7runtime.StoreAddressToString[zone], blockSize, startAddress)
			if item, exist := itemMap[addressKey]; !exist {
				it := &S7Item{
					RequestData:  newS7COMMReadParameterItem(s7runtime.StoreAreaTransportSize[key], variable.DataRequestLength(key), uint16(blockSize), s7runtime.StoreAreaCode[key], startAddress, bitAddress),
					StartAddress: uint(startAddressOffset),
				}
				startAddressOffset = startAddressOffset + offset
				itemMap[addressKey] = it
				items = append(items, it)
				vp := &VariableParse{
					Variable:           variable,
					StartAddress:       it.StartAddress,
					BitAddressOrLength: bitAddress,
					BlockSize:          blockSize,
				}
				variableParses = append(variableParses, vp)
			} else {
				vp := &VariableParse{
					Variable:           variable,
					StartAddress:       item.StartAddress,
					BitAddressOrLength: bitAddress,
					BlockSize:          blockSize,
				}
				variableParses = append(variableParses, vp)
			}

			if uint16(len(items)) == maxItemPerDataFrame {
				frame := newS7DataFrame(key, variableParses, items, startAddressOffset, maxPdu)
				dataFrames = append(dataFrames, frame)

				itemMap = make(map[string]*S7Item, 0)
				items = make([]*S7Item, 0)
				variableParses = make([]*VariableParse, 0)
				startAddressOffset = 0
			}
		}
		if len(items) > 0 {
			frame := newS7DataFrame(key, variableParses, items, startAddressOffset, maxPdu)
			dataFrames = append(dataFrames, frame)
		}
		storeAddressDataFrameMap[key] = dataFrames
	}

	dataFrameCount := 0
	for _, values := range storeAddressDataFrameMap {
		dataFrameCount += len(values)
	}
	if dataFrameCount == 0 {
		klog.V(2).InfoS("Failed to collect from s7 server.Because of the variables is empty", "deviceId", device.ID)
		return nil, nil, nil
	}

	CanCollect = true
	clients, err := model.S7Modelers[device.DeviceModel].NewClients(device.Address, dataFrameCount)
	if err != nil {
		klog.V(2).InfoS("Failed to collect from S7 server", "error", err, "deviceId", device.ID)
		return nil, nil, nil
	}

	s7c := &S7Collector{
		Device:                   device,
		ExitCh:                   make(chan struct{}, 0),
		StoreAddressDataFrameMap: storeAddressDataFrameMap,
		VariableCh:               make(chan *runtime.ParseVariableResult, 1),
		VariableCount:            len(device.Variables),
		CanCollect:               CanCollect,
		Clients:                  clients,
	}
	return s7c, s7c.VariableCh, nil
}

func (collector *S7Collector) Destroy(ctx context.Context) {
	collector.ExitCh <- struct{}{}
	collector.Clients.Destroy(ctx)
	close(collector.VariableCh)
}

func (collector *S7Collector) Collect(ctx context.Context) {
	if collector.CanCollect {
		go func() {
			for {
				start := time.Now().Unix()
				if !collector.poll(ctx) {
					return
				}
				select {
				case <-collector.ExitCh:
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

func (collector *S7Collector) poll(ctx context.Context) bool {
	select {
	case <-collector.ExitCh:
		return false
	default:
		sw := &sync.WaitGroup{}
		dfvCh := make(chan *s7runtime.ParseVariableResult, 0)
		for _, dataFrames := range collector.StoreAddressDataFrameMap {
			for _, frame := range dataFrames {
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
func (collector *S7Collector) message(ctx context.Context, dataFrame *S7DataFrame, pvrCh chan<- *s7runtime.ParseVariableResult, sw *sync.WaitGroup, clients *s7runtime.Clients) {
	defer sw.Done()
	defer func() {
		if err := recover(); err != nil {
			klog.V(2).InfoS("Failed to ask s7 message", "error", err)
		}
	}()
	messenger, err := clients.GetMessenger(ctx)
	defer clients.ReleaseMessenger(messenger)
	if err != nil {
		klog.V(2).InfoS("Failed to get tunnel", "error", err)
		if messenger, err = collector.Clients.NewMessenger(); err != nil {
			return
		}
	}

	var buf []byte
	if err := collector.retry(func(messenger s7runtime.Messenger, dataFrame *S7DataFrame) error {
		least, err := messenger.AskAtLeast(dataFrame.DataFrame, dataFrame.ResponseDataFrame, 19)
		if err != nil {
			return s7runtime.ErrBadConn
		}
		buf, err = dataFrame.ValidateMessage(least)
		if err != nil {
			return s7runtime.ErrServerBadResp
		}
		return nil
	}, messenger, dataFrame); err != nil {
		klog.V(2).InfoS("Failed to connect s7 server by retry three times")
		pvrCh <- &s7runtime.ParseVariableResult{Err: []error{err}}
		return
	}

	pvrCh <- &s7runtime.ParseVariableResult{Err: nil, VariableSlice: dataFrame.ParseVariableValue(buf[21:])}
}

func (collector *S7Collector) retry(fun func(messenger s7runtime.Messenger, dataFrame *S7DataFrame) error, messenger s7runtime.Messenger, dataFrame *S7DataFrame) error {
	for i := 0; i < 3; i++ {
		err := fun(messenger, dataFrame)
		if err == nil {
			return nil
		} else if errors.Is(err, s7runtime.ErrBadConn) {
			messenger.Close()
			newMessenger, err := collector.Clients.NewMessenger()
			if err != nil {
				return err
			}
			messenger.Reset(newMessenger)
			i = i - 1
		} else {
			klog.V(2).InfoS("Failed to connect s7 server", "error", err)
		}
	}
	return s7runtime.ErrManyRetry
}

func (collector *S7Collector) rollVariable(ctx context.Context, ch chan *s7runtime.ParseVariableResult) {
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

func newS7DataFrame(key s7runtime.S7StoreArea, variableParse []*VariableParse, items []*S7Item, dataLength int, pdu uint16) *S7DataFrame {
	data := []byte{0x03, 0x00, 0x00}
	maxBytes := 7 + 10 + 2 + len(items)*12
	data = append(data, uint8(maxBytes))
	cotpBytes := []byte{0x02, 0xf0, 0x80}
	data = append(data, cotpBytes...)
	s7HeaderBytesSuffix := []byte{0x32, 0x01, 0x00, 0x00, 0x00, 0x01}
	data = append(data, s7HeaderBytesSuffix...)
	data = append(data, binutil.Uint16ToBytes(uint16(2+len(items)*12))...)
	data = append(data, binutil.Uint16ToBytes(uint16(0))...)
	data = append(data, uint8(4))
	data = append(data, uint8(len(items)))
	for _, item := range items {
		data = append(data, item.RequestData...)
	}
	// 0xff, // 总字节数
	//		// cotp
	//		0x02, // parameter length
	//		0xf0, // 设置通信
	//		0x80, // TPDU number
	//		// s7 header                                10
	//		0x32, // 协议
	//		0x01, // 主站发送请求
	//		0x00, 0x00,
	//		0x00, 0x01,
	//		0x00, 0x0e, // parameter length
	//		0x00, 0x00, // Data length
	//		// s7 parameter                              2
	//		0x04, // read value
	//		0x01, // item count
	df := &S7DataFrame{
		Zone:              key,
		DataFrame:         data,
		ResponseDataFrame: make([]byte, pdu),
		Variables:         variableParse,
		ItemCount:         uint8(len(items)),
		DataLength:        dataLength,
	}
	return df
}

func newS7COMMReadParameterItem(transportSize uint8, length uint16, dbNumber uint16, zone uint8, address uint32, bitAddress uint8) []byte {
	itemBytes := []byte{
		0x12, // 结构标识
		0x0a, // 此字节往后的字节长度
		0x10, // Syntax Id Address data s7-any pointer
		// 0xff,       // Transport size 0x01 BIT 0x02 Byte 0x03 CHAR 0x04 WORD 0x05 INT 0x06 DWORD 0x07 DINT 0x08 REAL 0x09 DATE
		// 0xff, 0xff, // 数据长度
		// 0xff, 0xff, // 数据块编号 DB2  DB2.DBD24
		// 0xff,             // Area 0x81 I   0x82 Q  0x83 M  0x84 (DB) V  0x85 DI  0x86 L 0x87 V  0x1c C   0x1d T   0x1e IEC计数器   0x1f IEC定时器
		// 0xff, 0xff, 0xff, // Byte Address(18-3) BitAdress(2-0)
	}
	// Transport size 0x01 BIT 0x02 Byte 0x03 CHAR 0x04 WORD 0x05 INT 0x06 DWORD 0x07 DINT 0x08 REAL 0x09 DATE
	itemBytes = append(itemBytes, transportSize)
	itemBytes = append(itemBytes, binutil.Uint16ToBytes(length)...)
	itemBytes = append(itemBytes, binutil.Uint16ToBytes(dbNumber)...)
	// Area 0x81 I   0x82 Q  0x83 M  0x84 (DB) V  0x85 DI  0x86 L 0x87 V  0x1c C   0x1d T   0x1e IEC计数器   0x1f IEC定时器
	itemBytes = append(itemBytes, zone)
	itemBytes = append(itemBytes, uint8((address<<3)/256/256))
	itemBytes = append(itemBytes, uint8((address<<3)/256%256))
	itemBytes = append(itemBytes, uint8((address<<3)%256)+bitAddress)
	return itemBytes
}
