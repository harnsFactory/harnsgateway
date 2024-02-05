package runtime

import "harnsgateway/pkg/runtime"

func (in *S7Device) DeepCopyObject() runtime.RunObject {
	if in == nil {
		return nil
	}
	out := *in

	out.Address = in.Address.DeepCopy()

	out.VariablesMap = make(map[string]*Variable, len(in.Variables))
	if in.Variables != nil {
		out.Variables = make([]*Variable, len(in.Variables))
		for i, c := range in.Variables {
			copied := *c
			out.Variables[i] = &copied
			out.VariablesMap[copied.Name] = &copied
		}
	}

	return &out
}

func (in *S7Address) DeepCopy() *S7Address {
	if in == nil {
		return nil
	}

	out := *in
	out.Option = in.Option.DeepCopy()

	return &out
}

func (in *S7AddressOption) DeepCopy() *S7AddressOption {
	if in == nil {
		return nil
	}

	out := *in

	return &out
}
