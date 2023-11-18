package constant

import (
	"encoding/json"
	"fmt"
)

type AccessMode int8

const (
	AccessModeReadOnly AccessMode = iota
	AccessModeReadWrite
)

var ReadWritePropertyToString = map[AccessMode]string{
	AccessModeReadOnly:  "r",
	AccessModeReadWrite: "rw",
}

var StringToReadWriteProperty = map[string]AccessMode{
	"r":  AccessModeReadOnly,
	"rw": AccessModeReadWrite,
}

func (dt AccessMode) MarshalJSON() ([]byte, error) {
	if s, ok := ReadWritePropertyToString[dt]; ok {
		return json.Marshal(s)
	}
	return nil, fmt.Errorf("unknown accessMode %d", dt)
}

func (dt *AccessMode) UnmarshalJSON(bytes []byte) error {
	var s string
	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}

	v, ok := StringToReadWriteProperty[s]
	if !ok {
		return fmt.Errorf("unknown accessMode %s", s)
	}
	*dt = v
	return nil
}
