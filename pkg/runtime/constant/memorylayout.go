package constant

import (
	"encoding/json"
	"fmt"
)

type MemoryLayout byte

const (
	DCBA MemoryLayout = iota // little-endian
	CDAB                     // little-endian byte swap
	BADC                     // big-endian byte swap
	ABCD                     // big-endian
)

var MemoryLayoutToString = map[MemoryLayout]string{
	DCBA: "DCBA",
	CDAB: "CDAB",
	BADC: "BADC",
	ABCD: "ABCD",
}

var StringToMemoryLayout = map[string]MemoryLayout{
	"DCBA": DCBA,
	"CDAB": CDAB,
	"BADC": BADC,
	"ABCD": ABCD,
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
