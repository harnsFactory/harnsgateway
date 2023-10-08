package runtime

import "harnsgateway/pkg/runtime"

func (d *ModBusRtuDevice) DeepCopyObject() runtime.RunObject {
	out := *d
	return &out
}
