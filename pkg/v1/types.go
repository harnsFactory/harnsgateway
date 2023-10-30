package v1

type DeviceType interface {
	GetDeviceType() string
}

type DeviceMeta struct {
	Name        string `json:"name" binding:"required,min=1,max=64,excludesall=\u002F\u005C"`
	DeviceCode  string `json:"deviceCode" binding:"required,min=1,max=32,excludesall=\u002F\u005C"`
	DeviceType  string `json:"deviceType" binding:"required,min=1,max=32,excludesall=\u002F\u005C"`
	DeviceModel string `json:"deviceModel" binding:"required,min=1,max=32,excludesall=\u002F\u005C"`
}

func (d *DeviceMeta) GetDeviceType() string {
	return d.DeviceType
}
