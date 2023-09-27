package v1

import "harnsgateway/pkg/runtime"

func (d *DeviceMeta) DeepCopyObject() runtime.RunObject {
	out := *d
	return &out
}
