package myjson

import (
	"fmt"
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

type ValueError struct {
	Name string
}

func (e *ValueError) Error() string {
	if e.Name == "" {
		return "cannot interpret value as json value"
	} else {
		return fmt.Sprintf("cannot read value of name '%s'", e.Name)
	}
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

func (o Object) GetNumber(name string) (Number, error) {
	v := o.Get(name)
	return toNumber(v, name)
}

func (o Object) GetString(name string) (String, error) {
	v := o.Get(name)
	return toString(v, name)
}

type Path struct {
	paths []interface{}
}

func NewPath(paths ...interface{}) *Path {
	var _paths []interface{}
	for _, p := range paths {
		switch p.(type) {
		case string:
			_paths = append(_paths, p)
		case int:
			_paths = append(_paths, p)
		default:
			return nil
		}
	}
	return &Path{paths: _paths}
}

func (p *Path) Append(v interface{}) {
	switch v.(type) {
	case string:
		p.paths = append(p.paths, p)
	case int:
		p.paths = append(p.paths, p)
	}
}

func (p *Path) SetLast(v interface{}) {
	if len(p.paths) == 0 {
		return
	}

	lastIdx := len(p.paths) - 1
	switch v.(type) {
	case string:
		p.paths[lastIdx] = v
	case int:
		p.paths[lastIdx] = v
	}
}

func (p *Path) RemoveLast() {
	if len(p.paths) != 0 {
		p.paths = p.paths[:len(p.paths)-1]
	}
}

func (p *Path) String() string {
	result := "$"
	for _, v := range p.paths {
		switch v.(type) {
		case string:
			result += "." + v.(string)
		case int:
			result += fmt.Sprintf("[%d]", v.(int))
		}
	}
	return result
}
