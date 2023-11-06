package runtime

import "harnsgateway/pkg/runtime"

func (o *OpcUaDevice) DeepCopyObject() runtime.RunObject {
	out := *o
	return &out
}
