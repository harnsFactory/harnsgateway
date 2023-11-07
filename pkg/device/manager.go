package device

import (
	"context"
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"harnsgateway/pkg/apis"
	"harnsgateway/pkg/apis/response"
	"harnsgateway/pkg/gateway"
	"harnsgateway/pkg/generic"
	"harnsgateway/pkg/runtime"
	v1 "harnsgateway/pkg/v1"
	"k8s.io/klog/v2"
	"os"
	"strings"
	"sync"
	"time"
)

type Option func(*Manager)

type Manager struct {
	gatewayMeta    *gateway.GatewayMeta
	mqttClient     mqtt.Client
	mu             *sync.Mutex
	deviceManager  map[string]DeviceManager
	devices        *sync.Map
	store          *generic.Store
	brokers        map[string]runtime.Broker
	brokerReturnCh map[string]chan *runtime.ParseVariableResult
	stopCh         <-chan struct{}
	restartCh      <-chan string
	closers        []runtime.LabeledCloser
}

func NewManager(store *generic.Store, mqttClient mqtt.Client, gatewayMeta *gateway.GatewayMeta, stop <-chan struct{}, opts ...Option) *Manager {
	m := &Manager{
		gatewayMeta:    gatewayMeta,
		mqttClient:     mqttClient,
		mu:             &sync.Mutex{},
		devices:        &sync.Map{},
		deviceManager:  DeviceManagers,
		brokers:        make(map[string]runtime.Broker, 0),
		brokerReturnCh: make(map[string]chan *runtime.ParseVariableResult, 0),
		store:          store,
		stopCh:         stop,
		restartCh:      make(chan string, 0),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *Manager) Init() {
	devices, _ := m.store.LoadResource()
	for _, objects := range devices {
		// objects.IndexDevice()
		obj, _ := runtime.AccessorDevice(objects)
		m.devices.Store(obj.GetID(), obj)
		if err := m.readyCollect(obj); err != nil {
			klog.V(1).InfoS("Failed to start collect data", "deviceId", obj.GetID())
		}
	}
}

func (m *Manager) GetDeviceById(id string, exploded bool) (runtime.Device, error) {
	d, isExist := m.devices.Load(id)
	if !isExist {
		return nil, os.ErrNotExist
	}
	device, _ := d.(runtime.Device)
	if !exploded {
		return m.foldDevice(device), nil
	}
	return device, nil
}

func (m *Manager) CreateDevice(object v1.DeviceType) (runtime.Device, error) {
	device, err := m.deviceManager[object.GetDeviceType()].CreateDevice(object)
	if err != nil {
		klog.V(2).InfoS("Failed to create device", "error", err)
		return nil, err
	}
	created, err := m.store.Create(device)
	if err != nil {
		klog.V(2).InfoS("Failed to store device", "error", err)
		return nil, err
	}
	rd := created.(runtime.Device)
	m.devices.Store(rd.GetID(), rd)
	obj, _ := runtime.AccessorDevice(created)

	if err := m.readyCollect(obj); err != nil {
		return nil, err
	}
	return rd, nil
}

func (m *Manager) DeleteDevice(id string, version string) (runtime.Device, error) {
	device, err := m.GetDeviceById(id, false)
	if err != nil {
		return nil, err
	}

	if device.GetVersion() != version {
		return nil, apis.ErrMismatch
	}

	d, err := m.deviceManager[device.GetDeviceType()].DeleteDevice(device)
	if err != nil {
		klog.V(2).InfoS("Failed to delete device", "error", err)
		return nil, err
	}

	if _, err := m.store.Delete(d); err != nil {
		klog.V(2).InfoS("Failed to delete device", "deviceId", device.GetID())
	}

	klog.V(2).InfoS("Deleted device", "deviceId", device.GetID())

	go func() {
		if err := m.cancelCollect(device); err != nil {
			klog.V(2).InfoS("Failed to cancel collect data", "deviceId", device.GetID())
		}
	}()

	m.devices.Delete(device.GetID())
	return device, nil
}

func (m *Manager) ListDevices(filter *runtime.DeviceFilter, exploded bool) ([]runtime.Device, error) {
	rds := make([]runtime.Device, 0)
	predicates := runtime.ParseTypeFilter(filter)

	// descend
	byModTime := func(d1, d2 runtime.Device) bool { return d1.GetModTime().Before(d2.GetModTime()) }
	sorter := runtime.ByDevice(byModTime)

	m.devices.Range(func(key, value interface{}) bool {
		isMatch := true
		v := value.(runtime.Device)
		for _, p := range predicates {
			if !p(v) {
				isMatch = false
				break
			}
		}
		if isMatch {
			rds = sorter.Insert(rds, v)
		}
		return true
	})

	if !exploded {
		for i := range rds {
			rds[i] = m.foldDevice(rds[i])
		}
	}

	return rds, nil
}

func (m *Manager) DeliverAction(id string, actions []map[string]interface{}) error {
	device, err := m.GetDeviceById(id, true)
	if err != nil {
		return err
	}

	errs := &response.MultiError{}
	legalActions := make(map[string]interface{}, 0)
	variablesMap := device.GetVariablesMap()
	for _, item := range actions {
		for k, v := range item {
			if _, exist := legalActions[k]; exist {
				errs.Add(response.ErrResourceExists(k))
				continue
			}
			if _, ok := variablesMap[k]; !ok {
				errs.Add(response.ErrResourceNotFound(k))
				continue
			}
			// todo define rw r w 类型

			legalActions[k] = v
		}
	}

	if errs.Len() > 0 {
		return errs
	}

	if len(legalActions) == 0 {
		return response.NewMultiError(response.ErrLegalActionNotFound)
	}

	return m.brokers[id].DeliverAction(context.Background(), legalActions)
}

func (m Manager) cancelCollect(obj runtime.Device) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if v, ok := m.brokers[obj.GetID()]; ok {
		v.Destroy(context.Background())
	}
	delete(m.brokers, obj.GetID())
	delete(m.brokerReturnCh, obj.GetID())
	return nil
}

func (m *Manager) readyCollect(obj runtime.Device) error {
	broker, results, err := generic.DeviceTypeBrokerMap[obj.GetDeviceType()](obj)
	if err != nil {
		klog.V(2).InfoS("Failed to create device", "deviceId", obj.GetID())
		return err
	}
	if broker == nil {
		klog.V(2).InfoS("Failed to collect device data", "deviceId", obj.GetID())
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.brokers[obj.GetID()] = broker
	m.brokerReturnCh[obj.GetID()] = results

	topic := obj.GetTopic()
	if len(topic) == 0 {
		topic = fmt.Sprintf("data/%s/v1/%s", m.gatewayMeta.ID, obj.GetID())
	}

	broker.Collect(context.Background())
	go func(deviceId string, ch chan *runtime.ParseVariableResult) {
		for {
			select {
			case _, ok := <-m.stopCh:
				if !ok {
					return
				}
			case pvr, ok := <-results:
				if ok {
					if v, ok := m.devices.Load(deviceId); ok {
						if len(pvr.Err) == 0 {
							v.(runtime.Device).SetCollectStatus(true)
							pds := make([]runtime.PointData, 0, len(pvr.VariableSlice))
							for _, value := range pvr.VariableSlice {
								pd := runtime.PointData{
									DataPointId: value.GetVariableName(),
									Value:       value.GetValue(),
								}
								pds = append(pds, pd)
							}
							publishData := runtime.PublishData{Payload: runtime.Payload{Data: []runtime.TimeSeriesData{{
								Timestamp: time.Now().UTC().Format("2006-02-01T15:04:05.000Z"),
								Values:    pds,
							}}}}

							marshal, _ := json.Marshal(publishData)
							token := m.mqttClient.Publish(topic, 1, false, marshal)
							if token.WaitTimeout(mqttTimeout) && token.Error() == nil {
								klog.V(5).InfoS("Succeed to publish MQTT", "topic", topic, "data", publishData)
							} else {
								klog.V(1).InfoS("Failed to publish MQTT", "topic", topic, "err", token.Error())
							}
						} else {
							v.(runtime.Device).SetCollectStatus(false)
						}
					} else {
						klog.V(2).InfoS("Failed to load device", "deviceId", deviceId)
					}
				} else {
					klog.V(2).InfoS("Stopped to collect data", "deviceId", deviceId)
					return
				}
			}
		}
	}(obj.GetID(), results)
	return nil
}

func (m *Manager) Shutdown(context context.Context) error {
	for _, c := range m.brokers {
		c.Destroy(context)
	}

	m.mqttClient.Disconnect(2000)
	var errs []string
	for i := len(m.closers); i > 0; i-- {
		lc := m.closers[i-1]
		if err := lc.Closer(context); err != nil {
			klog.V(2).InfoS("Failed to stopped Dependencies service", "service", lc.Label)
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("Failed to shut down server: [%s]\n", strings.Join(errs, ","))
	}
	return nil
}

func (m *Manager) foldDevice(device runtime.Device) runtime.Device {
	return &runtime.DeviceMeta{
		ObjectMeta: runtime.ObjectMeta{
			Name:    device.GetName(),
			ID:      device.GetID(),
			Version: device.GetVersion(),
			ModTime: device.GetModTime(),
		},
		DeviceModel:   device.GetDeviceModel(),
		DeviceCode:    device.GetDeviceCode(),
		DeviceType:    device.GetDeviceType(),
		CollectStatus: device.GetCollectStatus(),
	}
}
