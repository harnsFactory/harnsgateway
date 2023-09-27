package runtime

import "harnsgateway/pkg/runtime"

func (d *OpcUaDevice) DeepCopyObject() runtime.RunObject {
	out := *d
	return &out
}
