package myjson

import (
	"fmt"
	"strconv"

	"github.com/kumasuke120/mockuma/internal/typeutil"
)

type Object map[string]interface{}
type Array []interface{}

func (o Object) String() string {
	return fmt.Sprintf("%v", map[string]interface{}(o))
}

func (a Array) String() string {
	return fmt.Sprintf("%v", []interface{}(a))
}

type Number float64
type String string
type Boolean bool

func (n Number) String() string {
	return fmt.Sprintf("%v", float64(n))
}

func (s String) String() string {
	return strconv.Quote(string(s))
}

func (b Boolean) String() string {
	return strconv.FormatBool(bool(b))
}

type ValueError struct {
	Name string
}

func (e *ValueError) Error() string {
	if e.Name == "" {
		return "Cannot interpret value as json value"
	} else {
		return fmt.Sprintf("Cannot read value of name '%s'", e.Name)
	}
}

func (o Object) Get(name string) interface{} {
	return map[string]interface{}(o)[name]
}

func (o Object) GetObject(name string) (Object, error) {
	v := o.Get(name)
	switch v.(type) {
	case Object:
		return v.(Object), nil
	default:
		return Object{}, &ValueError{Name: name}
	}
}

func (o Object) GetArray(name string) (Array, error) {
	v := o.Get(name)
	switch v.(type) {
	case Array:
		return v.(Array), nil
	default:
		return Array{}, &ValueError{Name: name}
	}
}

func (o Object) GetString(name string) (String, error) {
	v := o.Get(name)
	return toString(v, name)
}

func ToString(v interface{}) (String, error) {
	return toString(v, "")
}

func toString(v interface{}, name string) (String, error) {
	if v == nil {
		return "", &ValueError{Name: name}
	} else {
		return String(typeutil.ToString(v)), nil
	}
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
	var result []interface{}
	for _, _v := range v {
		result = append(result, toMyJson(_v))
	}
	return result
}
