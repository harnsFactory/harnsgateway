package gateway

import (
	"bytes"
	"encoding/json"
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/storage"
	"harnsgateway/pkg/utils/randutil"
	"harnsgateway/pkg/utils/uuidutil"
	"k8s.io/klog/v2"
	"os"
	"strconv"
	"time"
)

type Option func(*Manager)

type Manager struct {
	gatewayMeta *GatewayMeta
	stopCh      <-chan struct{}
}

func NewGatewayManager(stop <-chan struct{}, opts ...Option) *Manager {
	m := &Manager{
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
}

func (m *Manager) GetGatewayMeta() (*GatewayMeta, error) {
	return m.gatewayMeta, nil
}
