package runtime

import "harnsgateway/pkg/runtime"

func (d *S7Device) DeepCopyObject() runtime.RunObject {
	out := *d
	return &out
}
