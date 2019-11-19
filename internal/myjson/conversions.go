package myjson

import (
	"strconv"

	"github.com/kumasuke120/mockuma/internal/typeutil"
)

func toNumber(v interface{}, name string) (Number, error) {
	switch v.(type) {
	case Number:
		return v.(Number), nil
	case String:
		f, err := strconv.ParseFloat(string(v.(String)), 64)
		if err != nil {
			return Number(0), &ValueError{Name: name}
		}
		return Number(f), nil
	default:
		return Number(0), &ValueError{Name: name}
	}
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

func NewArray(v ...interface{}) Array {
	return toMyJsonArray(v)
}
