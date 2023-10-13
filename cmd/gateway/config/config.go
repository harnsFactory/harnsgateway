package config

import (
	"harnsgateway/pkg/collector"
)

type Config struct {
	CollectorMgr *collector.Manager
	CertFile     string
	KeyFile      string
}
