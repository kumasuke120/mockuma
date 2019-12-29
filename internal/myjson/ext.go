package myjson

import (
	"regexp"

	"github.com/kumasuke120/mockuma/internal/typeutil"
)

type ExtRegexp *regexp.Regexp

type ExtJSONMatcher struct {
	v interface{}
}

func NewExtJSONMatcher(v interface{}) *ExtJSONMatcher {
	return &ExtJSONMatcher{v: v}
}

func MakeExtJSONMatcher(v interface{}) ExtJSONMatcher {
	return ExtJSONMatcher{v: v}
}

func (m ExtJSONMatcher) Unwrap() interface{} {
	return m.v
}

func (m ExtJSONMatcher) Matches(v interface{}) bool {
	return m.matches(m.v, v)
}

func (m ExtJSONMatcher) matches(mv interface{}, v interface{}) bool {
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
	case ExtJSONMatcher:
		return m.matches(mv.(ExtJSONMatcher).v, v)
	}

	return false
}

func (m ExtJSONMatcher) objectMatches(mv Object, v interface{}) bool {
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

func (m ExtJSONMatcher) arrayMatches(mv Array, v interface{}) bool {
	switch v.(type) {
	case Array:
		_v := v.(Array)
		for idx, val := range mv {
			if val == nil { // special treatment for Array
				continue
			}

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

func (m ExtJSONMatcher) numberMatches(mv Number, v interface{}) bool {
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

func (m ExtJSONMatcher) stringMatches(mv String, v interface{}) bool {
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

func (m ExtJSONMatcher) booleanMatches(mv Boolean, v interface{}) bool {
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

func (m ExtJSONMatcher) regexpMatches(mv ExtRegexp, v interface{}) bool {
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
