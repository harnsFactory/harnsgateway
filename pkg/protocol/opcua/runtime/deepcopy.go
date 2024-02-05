package runtime

import "harnsgateway/pkg/runtime"

func (in *OpcUaDevice) DeepCopyObject() runtime.RunObject {
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

func (in *Address) DeepCopy() *Address {
	if in == nil {
		return nil
	}

	out := *in
	out.Option = in.Option.DeepCopy()

	return &out
}

func (in *Option) DeepCopy() *Option {
	if in == nil {
		return nil
	}

	out := *in

	return &out
}
