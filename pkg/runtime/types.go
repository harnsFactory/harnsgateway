package runtime

import (
	"context"
	"net/url"
	"time"
)

type LabeledCloser struct {
	Label  string
	Closer func(context.Context) error
}

type ResponseModel struct {
	Devices interface{} `json:"devices,omitempty"`
}

type ParseVariableResult struct {
	VariableSlice []VariableValue
	Err           []error
}

type ObjectMeta struct {
	Name    string    `json:"name"`
	ID      string    `json:"id"`
	Version string    `json:"eTag"`
	ModTime time.Time `json:"modTime"`
}

type DeviceMeta struct {
	ObjectMeta
	DeviceCode    string `json:"deviceCode"`
	DeviceType    string `json:"deviceType"`
	CollectStatus bool   `json:"-"`
	DeviceModel   string `json:"deviceModel"`
}

type PublishData struct {
	Payload Payload `json:"payload"`
}

type Payload struct {
	Data []TimeSeriesData `json:"data"`
}

type TimeSeriesData struct {
	Timestamp string      `json:"timestamp"`
	Values    []PointData `json:"values"`
}

type PointData struct {
	DataPointId string      `json:"dataPointId"`
	Value       interface{} `json:"value"`
}

func (d *DeviceMeta) GetDeviceCode() string {
	return d.DeviceCode
}

func (d *DeviceMeta) SetDeviceCode(s string) {
	d.DeviceCode = s
}

func (d *DeviceMeta) GetDeviceType() string {
	return d.DeviceType
}

func (d *DeviceMeta) SetDeviceType(s string) {
	d.DeviceType = s
}

func (d *DeviceMeta) GetCollectStatus() bool {
	return d.CollectStatus
}

func (d *DeviceMeta) SetCollectStatus(collect bool) {
	d.CollectStatus = collect
}

func (d *DeviceMeta) GetDeviceModel() string {
	return d.DeviceModel
}

func (d *DeviceMeta) SetDeviceModel(model string) {
	d.DeviceModel = model
}

type CreateOptions struct {
	Query url.Values
}

type GetOptions struct {
	Version string
	Query   url.Values
}

type ListOptions struct {
	Filter map[string]interface{}
	Query  url.Values
}

type UpdateOptions struct {
	Version string
	Query   url.Values
}

type DeleteOptions struct {
	Version string
	Query   url.Values
}
