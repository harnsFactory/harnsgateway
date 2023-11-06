package runtime

import "harnsgateway/pkg/runtime"

func (s *S7Device) DeepCopyObject() runtime.RunObject {
	out := *s
	return &out
}
