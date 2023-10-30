package runtime

import "harnsgateway/pkg/runtime"

func (d *ModBusDevice) DeepCopyObject() runtime.RunObject {
	out := *d
	return &out
}
