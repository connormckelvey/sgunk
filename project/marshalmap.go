package project

import "encoding/json"

func MarshalMap(v any) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var o map[string]any
	if err := json.Unmarshal(b, &o); err != nil {
		return nil, err
	}
	return o, nil
}
