package collector

import (
	"context"
	"fmt"
	"harnsgateway/pkg/apis"
	"harnsgateway/pkg/generic"
	"harnsgateway/pkg/runtime"
	v1 "harnsgateway/pkg/v1"
	"k8s.io/klog/v2"
	"os"
	"strings"
	"sync"
	"time"
)

type labeledCloser struct {
	label  string
	closer func(context.Context) error
}

type Option func(*Manager)

func WithDeviceManager(protocol string, manager DeviceManager) Option {
	return func(m *Manager) {
		m.deviceManager[protocol] = manager
	}
}

type Manager struct {
	mu                *sync.Mutex
	deviceManager     map[string]DeviceManager
	devices           *sync.Map
	store             *generic.Store
	collectors        map[string]runtime.Collector
	collectorReturnCh map[string]chan *runtime.ParseVariableResult
	stopCh            <-chan struct{}
	restartCh         <-chan string
	closers           []labeledCloser
}

func NewCollectorManager(store *generic.Store, stop <-chan struct{}, opts ...Option) *Manager {
	m := &Manager{
		mu:                &sync.Mutex{},
		devices:           &sync.Map{},
		deviceManager:     make(map[string]DeviceManager, 0),
		collectors:        make(map[string]runtime.Collector, 0),
		collectorReturnCh: make(map[string]chan *runtime.ParseVariableResult, 0),
		store:             store,
		stopCh:            stop,
		restartCh:         make(chan string, 0),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *Manager) Init() {
	devices, _ := m.store.LoadResource()
	for _, objects := range devices {
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

func (m *Manager) deleteDevice(id string, version string) (runtime.Device, error) {
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
	// m.stopCh <- struct{}{}
	if err := m.cancelCollect(device); err != nil {
		klog.V(2).InfoS("Failed to cancel collect data", "deviceId", device.GetID())
	}
	return device, nil
}

func (m Manager) cancelCollect(obj runtime.Device) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if v, ok := m.collectors[obj.GetID()]; ok {
		v.Destroy(context.Background())
	}
	delete(m.collectors, obj.GetID())
	delete(m.collectorReturnCh, obj.GetID())
	return nil
}

func (m *Manager) readyCollect(obj runtime.Device) error {
	collector, results, err := generic.DeviceTypeCollectorMap[obj.GetDeviceType()](obj)
	if err != nil {
		klog.V(2).InfoS("Failed to collector device data", "deviceId", obj.GetID())
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.collectors[obj.GetID()] = collector
	m.collectorReturnCh[obj.GetID()] = results
	collector.Collect(context.Background())
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
							fmt.Println("+++++++++++++++++++++++++++++++++")
							fmt.Printf("_time:%s\n", time.Now())
							fmt.Printf("deviceId:%s\n", deviceId)
							for _, value := range pvr.VariableSlice {
								fmt.Printf("%s->%v\n", value.GetVariableName(), value.GetValue())
							}
							fmt.Println("+++++++++++++++++++++++++++++++++")
						}
					} else {
						// todo
						klog.V(2).InfoS("Stopped to collect data", "deviceId", deviceId)
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
	for _, c := range m.collectors {
		c.Destroy(context)
	}

	var errs []string
	for i := len(m.closers); i > 0; i-- {
		lc := m.closers[i-1]
		if err := lc.closer(context); err != nil {
			klog.V(2).InfoS("Failed to stopped subsystem", "subsystem", lc.label)
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("Failed to shut down server: [%s]\n", strings.Join(errs, ","))
	}
	return nil
}
