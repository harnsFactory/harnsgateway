package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"harnsgateway/pkg/host"
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/storage"
	"harnsgateway/pkg/utils/randutil"
	"harnsgateway/pkg/utils/uuidutil"
	"k8s.io/klog/v2"
	"os"
	"strconv"
	"sync"
	"time"
)

type Option func(*Manager)

type Manager struct {
	cpus        []float64
	mux         *sync.RWMutex
	gatewayMeta *GatewayMeta
	stopCh      <-chan struct{}
}

func NewGatewayManager(stop <-chan struct{}, opts ...Option) *Manager {
	m := &Manager{
		cpus:        make([]float64, 0, 15),
		mux:         &sync.RWMutex{},
		gatewayMeta: &GatewayMeta{},
		stopCh:      stop,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *Manager) Init() {
	client := &storage.FsClient{}
	client.Init(storage.StoreGroupGateway)

	gd, err := client.Get(gateway)
	if err != nil && os.IsNotExist(err) {
		m.gatewayMeta = &GatewayMeta{
			Secret: "",
			ObjectMeta: runtime.ObjectMeta{
				Name:    "harnsgateway",
				ID:      uuidutil.UUID(),
				Version: strconv.FormatUint(randutil.Uint64n(), 10),
				ModTime: time.Now(),
			},
		}
		klog.V(3).InfoS("Gateway information not exist,been created automatically", "gatewayId", m.gatewayMeta.ID)
		if _, err := client.Create(gateway, m.gatewayMeta); err != nil {
			klog.V(2).InfoS("Failed to create gateway information", "err", err)
		}
	} else if err = json.NewDecoder(bytes.NewReader(gd.([]byte))).Decode(m.gatewayMeta); err != nil {
		klog.V(2).InfoS("Failed to unmarshal gateway information", "err", err)
		return
	}

	go m.collectCpu()
}

func (m *Manager) GetGatewayMeta() (*GatewayMeta, error) {
	return m.gatewayMeta, nil
}

func (m *Manager) getGatewayCpu() (map[string]string, error) {
	m.mux.RLock()
	data := make(map[string]string, 0)
	if len(m.cpus) == 0 {
		data["1min"] = fmt.Sprintf("%0.0f%%", 0.0)
		data["5min"] = fmt.Sprintf("%0.0f%%", 0.0)
		data["15min"] = fmt.Sprintf("%0.0f%%", 0.0)
		return data, nil
	}
	data["1min"] = fmt.Sprintf("%0.0f%%", m.cpus[len(m.cpus)-1])
	if len(m.cpus) < 5 {
		count := 0.0
		for _, cpu := range m.cpus {
			count += cpu
		}
		data["5min"] = fmt.Sprintf("%0.0f%%", count/float64(len(m.cpus)))
		data["15min"] = fmt.Sprintf("%0.0f%%", count/float64(len(m.cpus)))
	} else {
		count := 0.0
		for i := len(m.cpus) - 1; i > len(m.cpus)-6; i-- {
			count += m.cpus[i]
		}
		data["5min"] = fmt.Sprintf("%0.0f%%", count/float64(5))
		for i := 0; i < len(m.cpus)-5; i++ {
			count += m.cpus[i]
		}
		data["15min"] = fmt.Sprintf("%0.0f%%", count/float64(len(m.cpus)))
	}
	m.mux.RUnlock()
	return data, nil
}

func (m *Manager) getGatewayMem() (*MemUsageInfo, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		klog.V(2).InfoS("Failed to get gateway mem info", "err", err)
		return nil, err
	}
	total := v.Total / 1024 / 1024
	used := v.Used / 1024 / 1024
	mui := &MemUsageInfo{
		Total:       fmt.Sprintf("%vMB", total),
		Used:        fmt.Sprintf("%vMB", used),
		UsedPercent: fmt.Sprintf("%0.0f%%", v.UsedPercent),
	}
	return mui, nil
}

func (m *Manager) getGatewayDisk() (map[string]*DiskUsageInfo, error) {
	data := map[string]*DiskUsageInfo{}
	osDiskUsage, err := disk.Usage(host.GetOSDisk())
	if err != nil {
		klog.V(2).InfoS("Failed to get gateway os disk info", "err", err)
		return nil, err
	}
	total := osDiskUsage.Total / 1024 / 1024
	used := osDiskUsage.Used / 1024 / 1024
	odu := &DiskUsageInfo{
		Total:       fmt.Sprintf("%vMB", total),
		Used:        fmt.Sprintf("%vMB", used),
		UsedPercent: fmt.Sprintf("%0.0f%%", osDiskUsage.UsedPercent),
	}
	data[host.GetOSDisk()] = odu

	dataDiskUsage, err := disk.Usage(host.GetDataDisk())
	if err != nil {
		klog.V(2).InfoS("Failed to get gateway data disk info", "err", err)
		return nil, err
	}
	total = dataDiskUsage.Total / 1024 / 1024
	used = dataDiskUsage.Used / 1024 / 1024
	ddu := &DiskUsageInfo{
		Total:       fmt.Sprintf("%vMB", total),
		Used:        fmt.Sprintf("%vMB", used),
		UsedPercent: fmt.Sprintf("%0.0f%%", dataDiskUsage.UsedPercent),
	}
	data[host.GetDataDisk()] = ddu

	return data, nil
}

func (m *Manager) collectCpu() {
	for {
		select {
		case _, ok := <-m.stopCh:
			if !ok {
				return
			}
		default:
			percent, err := cpu.Percent(time.Minute, false)
			if err != nil {
				klog.V(2).InfoS("Failed to collect gateway cpu info")
			}
			m.mux.Lock()
			if len(m.cpus) < 15 {
				m.cpus = append(m.cpus, percent[0])
			} else {
				m.cpus = append(m.cpus[1:], percent[0])
			}
			m.mux.Unlock()
		}
	}
}
