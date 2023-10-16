package options

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/pflag"
	"harnsgateway/cmd/gateway/config"
	"harnsgateway/pkg/collector"
	"harnsgateway/pkg/generic"
	baseoptions "harnsgateway/pkg/generic/options"
	"harnsgateway/pkg/storage"
	"k8s.io/klog/v2"
	"time"
)

type Options struct {
	Port           string        `json:"port"`
	Wait           time.Duration `json:"graceful-timeout"`
	MqttBrokerUrls []string      `json:"mqtt-broker-urls"`
	MqttUsername   string        `json:"mqtt-username"`
	MqttPassword   string        `json:"mqtt-password"`
	CertFile       string        `json:"cert-file"`
	KeyFile        string        `json:"key-file"`
	baseoptions.BaseOptions
	// logs.BaseOptions
}

const (
	_defaultPort         = "32200"
	_defaultWait         = 15 * time.Second
	_defaultMqttUsername = ""
	_defaultMqttPassword = ""
)

var (
	_defaultMqttBrokerUrls = []string{"tcp://127.0.0.1:1883"}
)

func NewDefaultOptions() *Options {
	return &Options{
		Port:           _defaultPort,
		Wait:           _defaultWait,
		MqttBrokerUrls: _defaultMqttBrokerUrls,
		MqttUsername:   _defaultMqttUsername,
		MqttPassword:   _defaultMqttPassword,
		BaseOptions:    baseoptions.NewDefaultBaseOptions(),
		CertFile:       "",
		KeyFile:        "",
		// BaseOptions: logs.NewOptions(),
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	// refer to node port assignment https://rancher.com/docs/rancher/v2.x/en/installation/requirements/ports/#commonly-used-ports
	fs.StringVarP(&o.Port, "port", "P", o.Port, "Port exposed")
	fs.DurationVar(&o.Wait, "graceful-timeout", o.Wait, "The duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	fs.StringSliceVarP(&o.MqttBrokerUrls, "mqtt-broker-urls", "", o.MqttBrokerUrls, "The MQTT broker urls. The format should be scheme://host:port Where \"scheme\" is one of \"tcp\", \"ssl\", or \"ws\"")
	fs.StringVarP(&o.MqttUsername, "mqtt-username", "u", o.MqttUsername, "The MQTT username")
	fs.StringVarP(&o.MqttPassword, "mqtt-password", "p", o.MqttPassword, "The MQTT password")
	fs.StringVarP(&o.CertFile, "cert-file", "", o.CertFile, "The Cert file")
	fs.StringVarP(&o.KeyFile, "key-file", "", o.KeyFile, "The Key file")
}

func (o *Options) Config(stopCh <-chan struct{}) (*config.Config, error) {
	c := &config.Config{}
	store, _ := generic.NewStore(storage.StoreGroupToString[storage.StoreGroupDevice], storage.Devices, generic.DeviceTypeObjectMap)
	// mqtt
	mqttOption := mqtt.NewClientOptions()
	for _, s := range o.MqttBrokerUrls {
		mqttOption = mqttOption.AddBroker(s)
	}
	mqttOption.SetUsername(o.MqttUsername)
	mqttOption.SetPassword(o.MqttPassword)
	mqttOption.SetOrderMatters(false)
	mqttOption.SetClientID("harns-gateway-" + o.Port)
	// mqttOption.SetConnectTimeout()
	mqttClient := mqtt.NewClient(mqttOption)
	klog.V(1).InfoS("Connected to MQTT", "servers", o.MqttBrokerUrls)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		klog.ErrorS(token.Error(), "Failed to connect MQTT", "servers", o.MqttBrokerUrls)
		return nil, token.Error()
	}

	collectorMgr := collector.NewCollectorManager(store, mqttClient, stopCh)

	collectorMgr.Init()
	c.CollectorMgr = collectorMgr
	c.KeyFile = o.KeyFile
	c.CertFile = o.CertFile
	return c, nil
}
