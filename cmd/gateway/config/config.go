package config

import (
	"harnsgateway/pkg/collector"
	"harnsgateway/pkg/gateway"
)

type Config struct {
	CollectorMgr *collector.Manager
	GatewayMgr   *gateway.Manager
	CertFile     string
	KeyFile      string
}
