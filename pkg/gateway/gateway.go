package gateway

import "harnsgateway/pkg/runtime"

type GatewayMeta struct {
	Secret string `json:"secret"`
	runtime.ObjectMeta
}

const gateway = "meta"
