package myjson

import (
	"encoding/json"
)

func Unmarshal(data []byte) (interface{}, error) {
	var v interface{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}

	return toMyJson(v), nil
}
