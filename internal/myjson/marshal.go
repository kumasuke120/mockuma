package myjson

import "encoding/json"

func Marshal(v interface{}) ([]byte, error) {
	_v := toRawType(v)
	bytes, err := json.Marshal(_v)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func toRawType(v interface{}) interface{} {
	switch v.(type) {
	case Object:
		return toRawTypeMap(v.(Object))
	case Array:
		return toRawTypeSlice(v.(Array))
	case Number:
		return float64(v.(Number))
	case String:
		return string(v.(String))
	case Boolean:
		return bool(v.(Boolean))
	case ExtRegexp:
		return nil
	case ExtJSONMatcher:
		return nil
	default:
		return v
	}
}

func toRawTypeMap(o Object) map[string]interface{} {
	result := make(map[string]interface{}, len(o))
	for name, value := range o {
		result[name] = toRawType(value)
	}
	return result
}

func toRawTypeSlice(a Array) []interface{} {
	result := make([]interface{}, len(a))
	for idx, value := range a {
		result[idx] = toRawType(value)
	}
	return result
}
