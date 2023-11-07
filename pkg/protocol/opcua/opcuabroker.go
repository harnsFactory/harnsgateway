package opcua

import (
	"context"
	"errors"
	"github.com/gopcua/opcua/ua"
	genericruntime "harnsgateway/pkg/generic/runtime"
	"harnsgateway/pkg/protocol/opcua/model"
	opcuaruntime "harnsgateway/pkg/protocol/opcua/runtime"
	"harnsgateway/pkg/runtime"
	"io"
	"k8s.io/klog/v2"
	"sync"
	"time"
)

var _ runtime.Broker = (*OpcUaBroker)(nil)

// 一次最多1000个
type OpuUaDataFrame struct {
	Variables        []*opcuaruntime.Variable
	RequestVariables *ua.ReadRequest
}

type OpcUaBroker struct {
	ExitCh                     chan struct{}
	Device                     *opcuaruntime.OpcUaDevice
	Clients                    *opcuaruntime.Clients
	NamespaceVariableDataFrame []*OpuUaDataFrame
	VariableCount              int
	VariableCh                 chan *runtime.ParseVariableResult
	CanCollect                 bool
}

func NewBroker(d runtime.Device) (runtime.Broker, chan *runtime.ParseVariableResult, error) {
	device, ok := d.(*opcuaruntime.OpcUaDevice)
	if !ok {
		klog.V(2).InfoS("Failed to new opc ua device,device type not supported")
		return nil, nil, opcuaruntime.ErrDeviceType
	}

	var CanCollect bool
	groupOf := genericruntime.VariablesInGroupOf[*opcuaruntime.Variable](device.Variables, 1000)
	namespaceVariableDataFrame := make([]*OpuUaDataFrame, 0, 0)

	for _, variables := range groupOf {
		requestVariables := make([]*ua.ReadValueID, 0, 0)
		for _, variable := range variables {
			switch variable.DataType {
			case runtime.NUMBER:
				address := variable.Address.(float64)
				id := ua.NewNumericNodeID(variable.Namespace, uint32(address))
				requestVariables = append(requestVariables, &ua.ReadValueID{NodeID: id})
			case runtime.STRING:
				address := variable.Address.(string)
				id := ua.NewStringNodeID(variable.Namespace, address)
				requestVariables = append(requestVariables, &ua.ReadValueID{NodeID: id})
			}
		}
		namespaceVariableDataFrame = append(namespaceVariableDataFrame, &OpuUaDataFrame{
			Variables: variables,
			RequestVariables: &ua.ReadRequest{
				MaxAge:             2000,
				TimestampsToReturn: ua.TimestampsToReturnBoth,
				NodesToRead:        requestVariables}})
	}
	if len(namespaceVariableDataFrame) == 0 {
		klog.V(2).InfoS("Failed to collect from OPC server.Because of the variables is empty", "deviceId", device.ID)
		return nil, nil, nil
	}
	CanCollect = true

	clients, err := model.OpcUaModelers[device.DeviceModel].NewClients(device.Address, len(namespaceVariableDataFrame))
	if err != nil {
		klog.V(2).InfoS("Failed to collect from OPC server", "error", err, "deviceId", device.ID)
		return nil, nil, err
	}
	mtc := &OpcUaBroker{
		Device:                     device,
		ExitCh:                     make(chan struct{}, 0),
		NamespaceVariableDataFrame: namespaceVariableDataFrame,
		VariableCh:                 make(chan *runtime.ParseVariableResult, 1),
		VariableCount:              len(device.Variables),
		CanCollect:                 CanCollect,
		Clients:                    clients,
	}
	return mtc, mtc.VariableCh, nil
}

func (broker *OpcUaBroker) Destroy(ctx context.Context) {
	broker.ExitCh <- struct{}{}
	broker.Clients.Destroy(ctx)
	close(broker.VariableCh)
}

func (broker *OpcUaBroker) Collect(ctx context.Context) {
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

func (broker *OpcUaBroker) DeliverAction(ctx context.Context, obj map[string]interface{}) error {
	// TODO implement me
	panic("implement me")
}

func (broker *OpcUaBroker) poll(ctx context.Context) bool {
	select {
	case <-broker.ExitCh:
		return false
	default:
		sw := &sync.WaitGroup{}
		dfvCh := make(chan *opcuaruntime.ParseVariableResult, 0)
		for _, dataFrames := range broker.NamespaceVariableDataFrame {
			sw.Add(1)
			go broker.message(ctx, dataFrames, dfvCh, sw)
		}
		go broker.rollVariable(ctx, dfvCh)
		sw.Wait()
		close(dfvCh)
		return true
	}

}

func (broker *OpcUaBroker) message(ctx context.Context, dataFrame *OpuUaDataFrame, pvrCh chan *opcuaruntime.ParseVariableResult, sw *sync.WaitGroup) {
	defer sw.Done()
	messenger, err := broker.Clients.GetMessenger(ctx)
	if err != nil {
		pvrCh <- &opcuaruntime.ParseVariableResult{Err: []error{err}}
	}
	defer broker.Clients.ReleaseMessenger(messenger)

	var response *ua.ReadResponse
	if err := broker.retry(func(messenger opcuaruntime.Messenger, dataFrame *OpuUaDataFrame) error {
		response, err = messenger.Read(ctx, dataFrame.RequestVariables)
		return err
	}, messenger, dataFrame); err != nil {
		klog.V(2).InfoS("Failed to connect opc ua server by retry three times")
		pvrCh <- &opcuaruntime.ParseVariableResult{Err: []error{err}}
		return
	}

	if response == nil {
		klog.V(2).InfoS("Failed to get opc ua server response")
		return
	}

	variables := make([]*opcuaruntime.Variable, 0, len(dataFrame.Variables))
	for i, variable := range dataFrame.Variables {
		if response.Results[i].Status == ua.StatusOK {
			variable.SetValue(response.Results[i].Value.Value())
			variables = append(variables, &opcuaruntime.Variable{
				DataType:     variable.DataType,
				Name:         variable.Name,
				Address:      variable.Address,
				Namespace:    variable.Namespace,
				DefaultValue: variable.DefaultValue,
				Value:        variable.Value,
			})
		}
	}

	pvrCh <- &opcuaruntime.ParseVariableResult{Err: nil, VariableSlice: variables}
}

func (broker *OpcUaBroker) retry(fun func(m opcuaruntime.Messenger, dataFrame *OpuUaDataFrame) error, m opcuaruntime.Messenger, dataFrame *OpuUaDataFrame) error {
	for i := 0; i < 3; i++ {
		err := fun(m, dataFrame)
		if err == nil {
			return nil
		}
		switch {
		case err == io.EOF && m.Available():
			newMessenger, err := broker.Clients.NewMessenger()
			if err != nil {
				return err
			}
			m.Reset(newMessenger)
			i = i - 1
			continue
		case errors.Is(err, ua.StatusBadSessionIDInvalid):
			newMessenger, err := broker.Clients.NewMessenger()
			if err != nil {
				return err
			}
			m.Reset(newMessenger)
			i = i - 1
			continue
		case errors.Is(err, ua.StatusBadSessionNotActivated):
			newMessenger, err := broker.Clients.NewMessenger()
			if err != nil {
				return err
			}
			m.Reset(newMessenger)
			i = i - 1
			continue
		case errors.Is(err, ua.StatusBadSecureChannelIDInvalid):
			continue
		default:
			klog.V(2).InfoS("Failed to read opc ua server data", "err", err)
		}
	}
	return opcuaruntime.ErrManyRetry
}

func (broker *OpcUaBroker) rollVariable(ctx context.Context, ch chan *opcuaruntime.ParseVariableResult) {
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
