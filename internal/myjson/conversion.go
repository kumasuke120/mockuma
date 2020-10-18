package myjson

import (
	"fmt"
	"strconv"

	"github.com/kumasuke120/mockuma/internal/types"
)

type valueError struct {
	name string
}

func (e *valueError) Error() string {
	if e.name == "" {
		return "cannot interpret value as json value"
	}

	return fmt.Sprintf("cannot read value of name '%s'", e.name)
}

func ToObject(v interface{}) (Object, error) {
	return toObject(v, "")
}

func toObject(v interface{}, name string) (Object, error) {
	switch v.(type) {
	case Object:
		return v.(Object), nil
	default:
		return Object{}, &valueError{name: name}
	}
}

func ToArray(v interface{}) (Array, error) {
	return toArray(v, "")
}

func toArray(v interface{}, name string) (Array, error) {
	switch v.(type) {
	case Array:
		return v.(Array), nil
	default:
		return Array{}, &valueError{name: name}
	}
}

func ToNumber(v interface{}) (Number, error) {
	return toNumber(v, "")
}

func toNumber(v interface{}, name string) (Number, error) {
	switch v.(type) {
	case Number:
		return v.(Number), nil
	case String:
		f, err := strconv.ParseFloat(string(v.(String)), 64)
		if err != nil {
			return Number(0), &valueError{name: name}
		}
		return Number(f), nil
	default:
		return Number(0), &valueError{name: name}
	}
}

func ToString(v interface{}) (String, error) {
	return toString(v, "")
}

func toString(v interface{}, name string) (String, error) {
	switch v.(type) {
	case nil:
		return "", &valueError{name: name}
	case String:
		return v.(String), nil
	default:
		return String(types.ToString(v)), nil
	}
}

func ToBoolean(v interface{}) (Boolean, error) {
	return toBoolean(v, "")
}

func toBoolean(v interface{}, name string) (Boolean, error) {
	switch v.(type) {
	case Boolean:
		return v.(Boolean), nil
	default:
		return false, &valueError{name: name}
	}
}

func NewArray(v ...interface{}) Array {
	return toMyJSONArray(v)
}
