package myjson

import (
	"regexp"

	"github.com/kumasuke120/mockuma/internal/typeutil"
)

var Undefined = ExtUndefined{}

type ExtRegexp *regexp.Regexp

type ExtUndefined struct {
}

type ExtJsonMatcher struct {
	v interface{}
}

func MakeExtJsonMatcher(v interface{}) ExtJsonMatcher {
	return ExtJsonMatcher{v: v}
}

func (m ExtJsonMatcher) Matches(v interface{}) bool {
	return m.matches(m.v, v)
}

func (m ExtJsonMatcher) matches(mv interface{}, v interface{}) bool {
	switch mv.(type) {
	case nil:
		return v == nil
	case Object:
		return m.objectMatches(mv.(Object), v)
	case Array:
		return m.arrayMatches(mv.(Array), v)
	case Number:
		return m.numberMatches(mv.(Number), v)
	case String:
		return m.stringMatches(mv.(String), v)
	case Boolean:
		return m.booleanMatches(mv.(Boolean), v)
	case ExtRegexp:
		return m.regexpMatches(mv.(ExtRegexp), v)
	case ExtJsonMatcher:
		return m.matches(mv.(ExtJsonMatcher).v, v)
	case ExtUndefined:
		return true
	}

	panic("Shouldn't happen")
}

func (m ExtJsonMatcher) objectMatches(mv Object, v interface{}) bool {
	switch v.(type) {
	case Object:
		_v := v.(Object)
		for key, val := range mv {
			_val := _v[key]
			if !m.matches(val, _val) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (m ExtJsonMatcher) arrayMatches(mv Array, v interface{}) bool {
	switch v.(type) {
	case Array:
		_v := v.(Array)
		for idx, val := range mv {
			_val := _v[idx]
			if !m.matches(val, _val) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (m ExtJsonMatcher) numberMatches(mv Number, v interface{}) bool {
	_mv := float64(mv)
	switch v.(type) {
	case Number:
		return typeutil.Float64AlmostEquals(_mv, float64(v.(Number)))
	case String:
		return typeutil.ToString(_mv) == string(v.(String))
	default:
		return false
	}
}

func (m ExtJsonMatcher) stringMatches(mv String, v interface{}) bool {
	_mv := string(mv)
	switch v.(type) {
	case String:
		return _mv == string(v.(String))
	case Number:
		return _mv == typeutil.ToString(float64(v.(Number)))
	case Boolean:
		return _mv == typeutil.ToString(bool(v.(Boolean)))
	default:
		return false
	}
}

func (m ExtJsonMatcher) booleanMatches(mv Boolean, v interface{}) bool {
	_mv := bool(mv)
	switch v.(type) {
	case Boolean:
		return _mv == bool(v.(Boolean))
	case String:
		return typeutil.ToString(_mv) == string(v.(String))
	default:
		return false
	}
}

func (m ExtJsonMatcher) regexpMatches(mv ExtRegexp, v interface{}) bool {
	var _mv *regexp.Regexp = mv
	switch v.(type) {
	case String:
		return _mv.MatchString(string(v.(String)))
	case Number:
		return _mv.MatchString(typeutil.ToString(v))
	case Boolean:
		return _mv.MatchString(typeutil.ToString(v))
	case ExtRegexp:
		var _v *regexp.Regexp = v.(ExtRegexp)
		return _mv.String() == _v.String()
	default:
		return false
	}
}
