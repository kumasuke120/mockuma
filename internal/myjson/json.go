package myjson

import (
	"fmt"
	"regexp"
	"strconv"
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

func (o Object) Has(name string) bool {
	_, ok := map[string]interface{}(o)[name]
	return ok
}

func (o Object) Get(name string) interface{} {
	return map[string]interface{}(o)[name]
}

func (o Object) GetObject(name string) (Object, error) {
	v := o.Get(name)
	return toObject(v, name)
}

func (o Object) GetArray(name string) (Array, error) {
	v := o.Get(name)
	return toArray(v, name)
}

func (o Object) GetNumber(name string) (Number, error) {
	v := o.Get(name)
	return toNumber(v, name)
}

func (o Object) GetString(name string) (String, error) {
	v := o.Get(name)
	return toString(v, name)
}

type ExtRegexp *regexp.Regexp
