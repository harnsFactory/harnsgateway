package runtime

import (
	"context"
	"fmt"
	"time"
)

var (
	ErrNotObject = fmt.Errorf("object does not implement the Object interfaces")
)

type RunObject interface {
	DeepCopyObject() RunObject
}

type ObjectMetaAccessor interface {
	GetObjectMeta() Object
}

type Broker interface {
	Collect(ctx context.Context)
	Destroy(ctx context.Context)
	DeliverAction(ctx context.Context, obj map[string]interface{}) error
}

type VariableValue interface {
	SetValue(value interface{})
	GetValue() interface{}
	GetVariableName() string
	SetVariableName(name string)
}

type Object interface {
	RunObject
	GetName() string
	SetName(string)
	GetID() string
	SetID(string)
	GetVersion() string
	SetVersion(string)
	GetModTime() time.Time
	SetModTime(time.Time)
}

type Publisher interface {
	SetTopic(string)
	GetTopic() string
}

type VariablesMap interface {
	GetVariablesMap() map[string]VariableValue
}
type IndexDevice interface {
	IndexDevice()
}

type Device interface {
	Object
	Publisher
	VariablesMap
	IndexDevice
	GetDeviceCode() string
	SetDeviceCode(string)
	GetDeviceType() string
	SetDeviceType(string)
	GetDeviceModel() string
	SetDeviceModel(string)
	GetCollectStatus() bool
	SetCollectStatus(bool)
}

var _ Device = (*DeviceMeta)(nil)

type DeviceMeta struct {
	ObjectMeta
	PublishMeta
	DeviceCode    string                   `json:"deviceCode"`
	DeviceType    string                   `json:"deviceType"`
	DeviceModel   string                   `json:"deviceModel"`
	CollectStatus bool                     `json:"-"`
	VariablesMap  map[string]VariableValue `json:"-"`
}

func (d *DeviceMeta) IndexDevice() {
}

func (d *DeviceMeta) GetVariablesMap() map[string]VariableValue {
	return d.VariablesMap
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

type ObjectMeta struct {
	Name    string    `json:"name"`
	ID      string    `json:"id"`
	Version string    `json:"eTag"`
	ModTime time.Time `json:"modTime"`
}

func (meta *ObjectMeta) GetName() string              { return meta.Name }
func (meta *ObjectMeta) SetName(name string)          { meta.Name = name }
func (meta *ObjectMeta) GetID() string                { return meta.ID }
func (meta *ObjectMeta) SetID(id string)              { meta.ID = id }
func (meta *ObjectMeta) GetVersion() string           { return meta.Version }
func (meta *ObjectMeta) SetVersion(version string)    { meta.Version = version }
func (meta *ObjectMeta) GetModTime() time.Time        { return meta.ModTime }
func (meta *ObjectMeta) SetModTime(modTime time.Time) { meta.ModTime = modTime }

type PublishMeta struct {
	Topic string `json:"topic,omitempty"`
}

func (pm *PublishMeta) GetTopic() string      { return pm.Topic }
func (pm *PublishMeta) SetTopic(topic string) { pm.Topic = topic }

func Accessor(obj interface{}) (Object, error) {
	switch t := obj.(type) {
	case Object:
		return t, nil
	case ObjectMetaAccessor:
		if m := t.GetObjectMeta(); m != nil {
			return m, nil
		}
		return nil, ErrNotObject
	default:
		return nil, ErrNotObject
	}
}

func AccessorDevice(obj interface{}) (Device, error) {
	switch t := obj.(type) {
	case Device:
		return t, nil
	default:
		return nil, ErrNotObject
	}
}
