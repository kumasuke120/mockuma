package myjson

import (
	"encoding/json"
	"regexp"
)

func Unmarshal(data []byte) (interface{}, error) {
	var v interface{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}
	return toMyJson(v), nil
}

func toMyJson(v interface{}) interface{} {
	switch v.(type) {
	case map[string]interface{}:
		return toMyJsonObject(v.(map[string]interface{}))
	case []interface{}:
		return toMyJsonArray(v.([]interface{}))
	case float64:
		return Number(v.(float64))
	case string:
		return String(v.(string))
	case bool:
		return Boolean(v.(bool))
	case *regexp.Regexp:
		return ExtRegexp(v.(*regexp.Regexp))
	case Object:
		return v.(Object)
	case Array:
		return v.(Array)
	case Number:
		return v.(Number)
	case String:
		return v.(String)
	case Boolean:
		return v.(Boolean)
	case ExtRegexp:
		return v.(ExtRegexp)
	case ExtJsonMatcher:
		return v.(ExtJsonMatcher)
	case nil:
		return nil
	}

	panic("Shouldn't happen")
}

func toMyJsonObject(v map[string]interface{}) Object {
	result := make(map[string]interface{}, len(v))
	for key, value := range v {
		result[key] = toMyJson(value)
	}
	return result
}

func toMyJsonArray(v []interface{}) Array {
	result := make([]interface{}, len(v))
	for i, _v := range v {
		result[i] = toMyJson(_v)
	}
	return result
}
