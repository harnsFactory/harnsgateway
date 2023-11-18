package constant

import (
	"encoding/json"
	"fmt"
)

type DataType int8

const (
	BOOL DataType = iota
	INT16
	FLOAT32
	FLOAT64
	INT32
	INT64
	UINT16
	NUMBER
	STRING
)

var DataTypeToString = map[DataType]string{
	BOOL:    "bool",
	INT16:   "int16",
	FLOAT32: "float32",
	FLOAT64: "float64",
	INT32:   "int32",
	INT64:   "int64",
	UINT16:  "uint16",
	NUMBER:  "number",
	STRING:  "string",
}

var StringToDataType = map[string]DataType{
	"bool":    BOOL,
	"int16":   INT16,
	"float32": FLOAT32,
	"float64": FLOAT64,
	"int32":   INT32,
	"int64":   INT64,
	"uint16":  UINT16,
	"number":  NUMBER,
	"string":  STRING,
}

var DataTypeWord = map[DataType]uint{
	BOOL:    1,
	INT16:   1,
	FLOAT32: 2,
	FLOAT64: 4,
	INT32:   2,
	INT64:   4,
	UINT16:  1,
	NUMBER:  1,
	STRING:  1,
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
