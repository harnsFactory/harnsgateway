package config

import (
	"harnsgateway/pkg/device"
	"harnsgateway/pkg/gateway"
)

type Config struct {
	DeviceMgr  *device.Manager
	GatewayMgr *gateway.Manager
	CertFile   string
	KeyFile    string
}
