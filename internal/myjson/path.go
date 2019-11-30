package myjson

import "fmt"

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
