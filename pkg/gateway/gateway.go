package gateway

import "harnsgateway/pkg/runtime"

type GatewayMeta struct {
	Secret string `json:"secret"`
	runtime.ObjectMeta
}

type ResponseModel struct {
	Cpus  interface{} `json:"cpus,omitempty"`
	Mem   interface{} `json:"mem,omitempty"`
	Disks interface{} `json:"disk,omitempty"`
}

type MemUsageInfo struct {
	Total       string
	Used        string
	UsedPercent string
}

type DiskUsageInfo struct {
	Total       string
	Used        string
	UsedPercent string
}

const gateway = "meta"
