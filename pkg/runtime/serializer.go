package runtime

import (
	"time"
)

func (t *Time) UnmarshalJSON(bytes []byte) error {
	if ft, err := time.Parse(time.RFC3339Nano, string(bytes)); err != nil {
		return err
	} else {
		*t = (Time)(ft)
	}
	return nil
}

func (t *Time) MarshalJSON() ([]byte, error) {
	ft := (*time.Time)(t).Format(time.RFC3339Nano)
	return []byte(ft), nil
}
