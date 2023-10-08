package options

import (
	"github.com/spf13/pflag"
	"harnsgateway/cmd/gateway/config"
	"harnsgateway/pkg/collector"
	"harnsgateway/pkg/generic"
	baseoptions "harnsgateway/pkg/generic/options"
	"harnsgateway/pkg/protocol/modbus"
	"harnsgateway/pkg/protocol/modbusrtu"
	"harnsgateway/pkg/protocol/opcua"
	"harnsgateway/pkg/protocol/s7"
	"harnsgateway/pkg/storage"
	"time"
)

type Options struct {
	Port string        `json:"port"`
	Wait time.Duration `json:"graceful-timeout"`
	baseoptions.BaseOptions
	// logs.BaseOptions
}

const (
	_defaultPort = "32200"
	_defaultWait = 15 * time.Second
)

func NewDefaultOptions() *Options {
	return &Options{
		Port:        _defaultPort,
		Wait:        _defaultWait,
		BaseOptions: baseoptions.NewDefaultBaseOptions(),
		// BaseOptions: logs.NewOptions(),
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	// refer to node port assignment https://rancher.com/docs/rancher/v2.x/en/installation/requirements/ports/#commonly-used-ports
	fs.StringVarP(&o.Port, "port", "P", o.Port, "Port exposed")
	fs.DurationVar(&o.Wait, "graceful-timeout", o.Wait, "The duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
}

func (o *Options) Config(stopCh <-chan struct{}) (*config.Config, error) {
	c := &config.Config{}
	store, _ := generic.NewStore(storage.StoreGroupToString[storage.StoreGroupDevice], storage.Devices, generic.DeviceTypeObjectMap)
	collectorMgr := collector.NewCollectorManager(store, stopCh,
		collector.WithDeviceManager("modbusTcp", &modbus.ModbusDeviceManager{}),
		collector.WithDeviceManager("opcUa", &opcua.OpcUaDeviceManager{}),
		collector.WithDeviceManager("s71500", &s7.S7DeviceManager{}),
		collector.WithDeviceManager("modbusRtu", &modbusrtu.ModbusRtuDeviceManager{}),
	)

	collectorMgr.Init()
	c.CollectorMgr = collectorMgr

	return c, nil
}
