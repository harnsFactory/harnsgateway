package runtime

import (
	"encoding/json"
	"fmt"
)

func (dt S7StoreAddress) MarshalJSON() ([]byte, error) {
	if s, ok := StoreAddressToString[dt]; ok {
		return json.Marshal(s)
	}
	return nil, fmt.Errorf("unknown store address %d", dt)
}

func (dt *S7StoreAddress) UnmarshalJSON(bytes []byte) error {
	var s string
	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}

	v, ok := StringToStoreAddress[s]
	if !ok {
		return fmt.Errorf("unknown store address %s", s)
	}
	*dt = v
	return nil
}
