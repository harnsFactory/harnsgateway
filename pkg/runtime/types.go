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

type Time time.Time

type TimeZone time.Location

type Predicate func(value interface{}) bool

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
