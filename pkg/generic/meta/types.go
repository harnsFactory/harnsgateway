package meta

// import (
// 	"net/url"
// 	"time"
// )
//
// type Time time.Time
//
// func (t *Time) UnmarshalJSON(bytes []byte) error {
// 	if ft, err := time.Parse(time.RFC3339Nano, string(bytes)); err != nil {
// 		return err
// 	} else {
// 		*t = (Time)(ft)
// 	}
// 	return nil
// }
//
// func (t *Time) MarshalJSON() ([]byte, error) {
// 	ft := (*time.Time)(t).Format(time.RFC3339Nano)
// 	return []byte(ft), nil
// }
//
// type ObjectMeta struct {
// 	Name    string    `json:"name"`
// 	ID      string    `json:"id"`
// 	Version string    `json:"eTag"`
// 	ModTime time.Time `json:"modTime"`
// }
//
// type DeviceMeta struct {
// 	DeviceCode string `json:"deviceCode"`
// 	DeviceType string `json:"deviceType"`
// }
//
// type CreateOptions struct {
// 	Query url.Values
// }
//
// type GetOptions struct {
// 	Version string
// 	Query   url.Values
// }
//
// type ListOptions struct {
// 	Filter map[string]interface{}
// 	Query  url.Values
// }
//
// type UpdateOptions struct {
// 	Version string
// 	Query   url.Values
// }
//
// type DeleteOptions struct {
// 	Version string
// 	Query   url.Values
// }
