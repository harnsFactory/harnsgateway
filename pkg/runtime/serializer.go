package runtime

import (
	"encoding/json"
	"fmt"
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

func (dt DataType) MarshalJSON() ([]byte, error) {
	if s, ok := DataTypeToString[dt]; ok {
		return json.Marshal(s)
	}
	return nil, fmt.Errorf("unknown data type %d", dt)
}

func (dt *DataType) UnmarshalJSON(bytes []byte) error {
	var s string
	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}

	v, ok := StringToDataType[s]
	if !ok {
		return fmt.Errorf("unknown data type %s", s)
	}
	*dt = v
	return nil
}

func (ml MemoryLayout) MarshalJSON() ([]byte, error) {
	if s, ok := MemoryLayoutToString[ml]; ok {
		return json.Marshal(s)
	}
	return nil, fmt.Errorf("unknown memory layout type %d", ml)
}

func (ml *MemoryLayout) UnmarshalJSON(bytes []byte) error {
	var s string
	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}

	v, ok := StringToMemoryLayout[s]
	if !ok {
		return fmt.Errorf("unknown memory layout type %s", s)
	}
	*ml = v
	return nil
}
