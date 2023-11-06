package runtime

import "harnsgateway/pkg/runtime"

func (m *ModBusDevice) DeepCopyObject() runtime.RunObject {
	out := *m
	return &out
}
