package myjson

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var validJavascriptIdentifier = regexp.MustCompile("(?i)^[_$a-z][_$a-z\\d]*$")

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
			panic("Shouldn't happen")
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
	var result strings.Builder
	result.WriteRune('$')

	for _, v := range p.paths {
		switch v.(type) {
		case string:
			_v := v.(string)
			if validJavascriptIdentifier.MatchString(_v) {
				result.WriteRune('.')
				result.WriteString(_v)
			} else {
				_v = strconv.Quote(_v)
				_v = _v[1 : len(_v)-1]
				_v = strings.ReplaceAll(_v, "\\\"", "\"")
				_v = strings.ReplaceAll(_v, "'", "\\'")
				result.WriteString("['")
				result.WriteString(_v)
				result.WriteString("']")
			}
		case int:
			result.WriteString(fmt.Sprintf("[%d]", v.(int)))
		}
	}
	return result.String()
}

type pathParseError struct {
	pathStr string
}

func (e *pathParseError) Error() string {
	return "cannot parse the string as json-path: " + strconv.Quote(e.pathStr)
}

const (
	psReady = iota
	psInKey
	psMaybeQuotedKeyOrIndex
	psInQuotedKey
	psOutQuotedKey
	psInIndex
)

func ParsePath(pathStr string) (*Path, error) {
	parseError := &pathParseError{pathStr: pathStr}

	if pathStr == "" || pathStr[0] != '$' {
		return nil, parseError
	}

	runes := []rune(pathStr)
	var paths []interface{}

	s := psReady
	var temp strings.Builder
	for i := 1; i < len(runes); i++ {
		r := runes[i]
		doWrite := false

		switch s {
		case psReady:
			if r == '.' {
				s = psInKey
			} else if r == '[' {
				s = psMaybeQuotedKeyOrIndex
			} else {
				return nil, parseError
			}
		case psInKey:
			if r == '.' || r == '[' || i == len(runes)-1 {
				s = psReady
				if i == len(runes)-1 {
					temp.WriteRune(r)
				} else {
					i -= 1
				}

				key, err := toKey(&temp, false)
				if err != nil {
					return nil, parseError
				}
				paths = append(paths, key)
			} else {
				doWrite = true
			}
		case psMaybeQuotedKeyOrIndex:
			if unicode.IsDigit(r) {
				s = psInIndex
				i -= 1
			} else if r == '\'' {
				s = psInQuotedKey
			} else {
				return nil, parseError
			}
		case psInQuotedKey:
			if r == '\'' && runes[i-1] != '\\' {
				s = psOutQuotedKey
			} else {
				doWrite = true
			}
		case psOutQuotedKey:
			if r == ']' {
				s = psReady

				key, err := toKey(&temp, true)
				if err != nil {
					return nil, parseError
				}
				paths = append(paths, key)
			} else {
				return nil, parseError
			}
		case psInIndex:
			if r == ']' {
				s = psReady

				index, err := toIndex(&temp)
				if err != nil {
					return nil, parseError
				}
				paths = append(paths, index)
			} else {
				doWrite = true
			}
		}

		if doWrite {
			temp.WriteRune(r)
		}
	}

	if s != psReady {
		return nil, parseError
	}

	return &Path{paths: paths}, nil
}

func toIndex(builder *strings.Builder) (int, error) {
	str := builder.String()
	builder.Reset()
	return strconv.Atoi(str)
}

func toKey(builder *strings.Builder, quoted bool) (string, error) {
	str := builder.String()
	builder.Reset()

	if quoted {
		str = strings.ReplaceAll(str, "\"", "\\\"")
		str = strings.ReplaceAll(str, "\\'", "'")
		_str, err := strconv.Unquote("\"" + str + "\"")
		if err != nil {
			return "", err
		}
		str = _str
	} else {
		if !validJavascriptIdentifier.MatchString(str) {
			return "", errors.New("invalid identifier for dot reference")
		}
	}

	return str, nil
}

func (o Object) SetEx(path *Path, v interface{}) (Object, error) {
	var t interface{}
	t = o

	for idx, p := range path.paths {
		isLast := idx == len(path.paths)-1

		switch p.(type) {
		case int:
			switch t.(type) {
			case Array:

			}
		}
	}

}
