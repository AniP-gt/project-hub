package parse

import (
	"encoding/json"
	"fmt"

	"project-hub/internal/state"
)

func ParseItemList(out []byte) ([]state.Item, error) {
	var items []any
	if err := json.Unmarshal(out, &items); err != nil {
		var obj map[string]any
		if err2 := json.Unmarshal(out, &obj); err2 != nil {
			return nil, fmt.Errorf("parse item-list json: %w", err)
		}
		if arr, ok := obj["items"].([]any); ok {
			items = arr
		} else {
			return nil, fmt.Errorf("parse item-list json: items not found")
		}
	}
	var result []state.Item
	for _, r := range items {
		if it, ok := ParseItemMap(r); ok {
			result = append(result, it)
		}
	}
	return result, nil
}
