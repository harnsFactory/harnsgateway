package runtime

func (d *DeviceMeta) DeepCopyObject() RunObject {
	out := *d
	return &out
}
