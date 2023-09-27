package opcua

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	genericruntime "harnsgateway/pkg/generic/runtime"
	opcuaruntime "harnsgateway/pkg/protocol/opcua/runtime"
	"harnsgateway/pkg/runtime"
	"io"
	"k8s.io/klog/v2"
	"sync"
	"time"
)

// 一次最多1000个
type OpuUaDataFrame struct {
	Variables        []*opcuaruntime.Variable
	RequestVariables *ua.ReadRequest
}

type OpcUaCollector struct {
	exitCh                     chan struct{}
	Device                     *opcuaruntime.OpcUaDevice
	Tunnels                    *Tunnels
	NamespaceVariableDataFrame []*OpuUaDataFrame
	VariableCount              int
	VariableCh                 chan *runtime.ParseVariableResult
	CanCollect                 bool
	Endpoint                   string
}

func NewCollector(d runtime.Device) (runtime.Collector, chan *runtime.ParseVariableResult, error) {
	device, ok := d.(*opcuaruntime.OpcUaDevice)
	if !ok {
		klog.V(2).InfoS("Failed to new opc ua collector,device type not supported")
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

	tcpChannel := 0
	tcpChannel += len(namespaceVariableDataFrame)

	if tcpChannel > 0 {
		tcpChannel = tcpChannel/5 + 1
		CanCollect = true
	}

	var endpoint string
	if device.Port <= 0 {
		endpoint = device.Address
	} else {
		endpoint = fmt.Sprintf("%s:%d", device.Address, device.Port)
	}

	cs := list.New()
	for i := 0; i < tcpChannel; i++ {
		c, err := opcua.NewClient(endpoint, opcua.SecurityMode(ua.MessageSecurityModeNone))
		if err != nil {
			klog.V(2).InfoS("Failed to get opc ua client")
		}
		if err := c.Connect(context.Background()); err != nil {
			klog.V(2).InfoS("Failed to connect opc ua server")
		}
		cs.PushBack(c)
	}

	tunnels := &Tunnels{
		clients:      cs,
		Max:          tcpChannel,
		Idle:         tcpChannel,
		Mux:          &sync.Mutex{},
		NextRequest:  1,
		ConnRequests: make(map[uint64]chan *opcua.Client, 0),
		newClient: func() (*opcua.Client, error) {
			c, err := opcua.NewClient(endpoint, opcua.SecurityMode(ua.MessageSecurityModeNone))
			if err != nil {
				klog.V(2).InfoS("Failed to get opc ua client")
			}
			if err := c.Connect(context.Background()); err != nil {
				klog.V(2).InfoS("Failed to connect opc ua server")
			}
			return c, nil
		},
	}

	mtc := &OpcUaCollector{
		Device:                     device,
		exitCh:                     make(chan struct{}, 0),
		NamespaceVariableDataFrame: namespaceVariableDataFrame,
		Endpoint:                   endpoint,
		VariableCh:                 make(chan *runtime.ParseVariableResult, 1),
		VariableCount:              len(device.Variables),
		CanCollect:                 CanCollect,
		Tunnels:                    tunnels,
	}
	return mtc, mtc.VariableCh, nil
}

func (collector *OpcUaCollector) Destroy(ctx context.Context) {
	collector.exitCh <- struct{}{}
	collector.Tunnels.Destroy(ctx)
	close(collector.VariableCh)
}

func (collector *OpcUaCollector) Collect(ctx context.Context) {
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

func (collector *OpcUaCollector) poll(ctx context.Context) bool {
	select {
	case <-collector.exitCh:
		return false
	default:
		sw := &sync.WaitGroup{}
		dfvCh := make(chan *opcuaruntime.ParseVariableResult, 0)
		for _, dataFrames := range collector.NamespaceVariableDataFrame {
			sw.Add(1)
			go collector.message(ctx, dataFrames, dfvCh, sw)
		}
		go collector.rollVariable(ctx, dfvCh)
		sw.Wait()
		close(dfvCh)
		return true
	}

}

func (collector *OpcUaCollector) message(ctx context.Context, dataFrame *OpuUaDataFrame, pvrCh chan *opcuaruntime.ParseVariableResult, sw *sync.WaitGroup) {
	defer sw.Done()
	c, err := collector.Tunnels.getTunnel(ctx)
	if err != nil {
		pvrCh <- &opcuaruntime.ParseVariableResult{Err: []error{err}}
	}
	defer collector.Tunnels.releaseTunnel(c)

	var response *ua.ReadResponse
	if err := collector.retry(func(dataFrame *OpuUaDataFrame) (*opcua.Client, error) {
		response, err = c.Read(ctx, dataFrame.RequestVariables)
		return c, err
	}, dataFrame); err != nil {
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
			variables = append(variables, variable)
		}
	}

	pvrCh <- &opcuaruntime.ParseVariableResult{Err: nil, VariableSlice: variables}
}

func (collector *OpcUaCollector) retry(fun func(dataFrame *OpuUaDataFrame) (*opcua.Client, error), dataFrame *OpuUaDataFrame) error {
	for i := 0; i < 3; i++ {
		c, err := fun(dataFrame)
		if err == nil {
			return nil
		}
		switch {
		case err == io.EOF && c.State() != opcua.Closed:
			continue
		case errors.Is(err, ua.StatusBadSessionIDInvalid):
			continue
		case errors.Is(err, ua.StatusBadSessionNotActivated):
			continue
		case errors.Is(err, ua.StatusBadSecureChannelIDInvalid):
			continue
		default:
			klog.V(2).InfoS("Failed to read opc ua server data", "err", err)
		}
	}
	return opcuaruntime.ErrManyRetry
}

func (collector *OpcUaCollector) rollVariable(ctx context.Context, ch chan *opcuaruntime.ParseVariableResult) {
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
