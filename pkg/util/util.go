package util

import (
	"encoding/json"
	"fmt"
)

func ConvertToMap(obj interface{}) (map[string]interface{}, error) {
	d := map[string]interface{}{}
	if b, err := json.Marshal(obj); err != nil {
		return nil, fmt.Errorf("marshal with error: %w", err)
	} else if err := json.Unmarshal(b, &d); err != nil {
		return nil, fmt.Errorf("unmarshal with error: %w", err)
	}
	return d, nil
}
