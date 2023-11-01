package config

import (
	"harnsgateway/pkg/broker"
	"harnsgateway/pkg/gateway"
)

type Config struct {
	CollectorMgr *broker.Manager
	GatewayMgr   *gateway.Manager
	CertFile     string
	KeyFile      string
}
