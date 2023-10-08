package runtime

import (
	"context"
	"fmt"
	"net/url"
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

type Collector interface {
	Collect(ctx context.Context)
	Destroy(ctx context.Context)
}

type VariableValue interface {
	SetValue(value interface{})
	GetValue() interface{}
	GetVariableName() string
	SetVariableName(name string)
}

type Object interface {
	GetName() string
	SetName(string)
	GetID() string
	SetID(string)
	GetVersion() string
	SetVersion(string)
	GetModTime() time.Time
	SetModTime(time.Time)
}

type Device interface {
	Object
	GetDeviceCode() string
	SetDeviceCode(string)
	GetDeviceType() string
	SetDeviceType(string)
	GetCollectStatus() bool
	SetCollectStatus(bool)
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
	CollectStatus bool   `json:"collectStatus"`
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

type Time time.Time

type TimeZone time.Location

type Predicate func(value interface{}) bool

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

func (meta *ObjectMeta) GetName() string              { return meta.Name }
func (meta *ObjectMeta) SetName(name string)          { meta.Name = name }
func (meta *ObjectMeta) GetID() string                { return meta.ID }
func (meta *ObjectMeta) SetID(id string)              { meta.ID = id }
func (meta *ObjectMeta) GetVersion() string           { return meta.Version }
func (meta *ObjectMeta) SetVersion(version string)    { meta.Version = version }
func (meta *ObjectMeta) GetModTime() time.Time        { return meta.ModTime }
func (meta *ObjectMeta) SetModTime(modTime time.Time) { meta.ModTime = modTime }

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
