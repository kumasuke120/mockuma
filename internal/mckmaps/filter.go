package mckmaps

import (
	"errors"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/kumasuke120/mockuma/internal/myjson"
)

// filters for the preprocessor
var (
	ppRemoveComment  = &commentFilter{}
	ppRenderTemplate = makeTemplateFilter()
	ppLoadFile       = makeLoadFileFilter()
	ppParseRegexp    = makeParseRegexpFilter()
	ppToJSONMatcher  = &jsonMatcherFilter{}
)

func makeTemplateFilter() *templateFilter {
	f := templateFilter{}
	f.reset()
	return &f
}

func makeLoadFileFilter() *loadFileFilter {
	f := new(loadFileFilter)
	f.reset()
	return f
}

func makeParseRegexpFilter() *parseRegexpFilter {
	f := new(parseRegexpFilter)
	f.reset()
	return f
}

func doFiltersOnV(v interface{}, filters ...filter) (interface{}, error) {
	chain := &filterChain{filters: filters}
	err := chain.doFilter(v)
	if err != nil {
		return nil, err
	}
	return chain.v, nil
}

type filterChain struct {
	filters []filter
	idx     int
	v       interface{}
}

func (c *filterChain) doFilter(v interface{}) error {
	c.v = v

	if c.idx < len(c.filters) {
		next := c.filters[c.idx]
		c.idx += 1
		return next.doFilter(v, c)
	}

	return nil
}

type filter interface {
	doFilter(v interface{}, chain *filterChain) error
}

type commentFilter struct { // removes all @comment directives in mockuMappings
}

func (f *commentFilter) doFilter(v interface{}, chain *filterChain) error {
	f.removeComment(v)
	return chain.doFilter(v)
}

func (f *commentFilter) removeComment(v interface{}) {
	switch v.(type) {
	case myjson.Object:
		delete(v.(myjson.Object), dComment)
		for _, value := range v.(myjson.Object) {
			f.removeComment(value)
		}
	case myjson.Array:
		for _, _v := range v.(myjson.Array) {
			f.removeComment(_v)
		}
	}
}

type loadFileFilter struct { // loads file contents for @file
	fileCache map[string][]byte
}

func (f *loadFileFilter) doFilter(v interface{}, chain *filterChain) error {
	rV, err := f.load(v)
	if err != nil {
		return err
	}
	return chain.doFilter(rV)
}

func (f *loadFileFilter) load(v interface{}) (interface{}, error) {
	var rV interface{}
	var err error
	switch v.(type) {
	case myjson.Object:
		rV, err = f.loadForObject(v.(myjson.Object))
	case myjson.Array:
		rV, err = f.loadForArray(v.(myjson.Array))
	default:
		rV, err = v, nil
	}
	return rV, err
}

func (f *loadFileFilter) loadForObject(v myjson.Object) (interface{}, error) {
	if v.Has(dFile) {
		filename, err := v.GetString(dFile)
		if err != nil {
			return nil, errors.New("cannot read filename from " + dFile)
		}

		_filename := string(filename)
		var bytes []byte
		var ok bool
		if bytes, ok = f.fileCache[_filename]; !ok {
			if err := checkFilepath(_filename); err != nil {
				return nil, errors.New(err.Error() + ": " + _filename)
			}

			bytes, err = ioutil.ReadFile(_filename)
			if err != nil {
				return nil, err
			}
			recordLoadedFile(_filename)
		}

		return myjson.String(string(bytes)), nil
	} else {
		rV := make(myjson.Object)
		for name, value := range v {
			rValue, err := f.load(value)
			if err != nil {
				return nil, err
			}
			rV[name] = rValue
		}
		return rV, nil
	}
}

func (f *loadFileFilter) loadForArray(v myjson.Array) (interface{}, error) {
	rV := make(myjson.Array, len(v))
	for idx, value := range v {
		rValue, err := f.load(value)
		if err != nil {
			return nil, err
		}
		rV[idx] = rValue
	}
	return rV, nil
}

func (f *loadFileFilter) reset() {
	f.fileCache = make(map[string][]byte)
}

type parseRegexpFilter struct {
	regexpCache map[string]myjson.ExtRegexp
}

func (f *parseRegexpFilter) doFilter(v interface{}, chain *filterChain) error {
	rV, err := f.parse(v)
	if err != nil {
		return err
	}
	return chain.doFilter(rV)
}

func (f *parseRegexpFilter) parse(v interface{}) (rV interface{}, err error) {
	switch v.(type) {
	case myjson.Object:
		rV, err = f.parseForObject(v.(myjson.Object))
	case myjson.Array:
		rV, err = f.parseForArray(v.(myjson.Array))
	case myjson.ExtJSONMatcher:
		rV, err = f.parseForExtJSONMatcher(v.(myjson.ExtJSONMatcher))
	default:
		rV, err = v, nil
	}
	return
}

func (f *parseRegexpFilter) parseForObject(v myjson.Object) (interface{}, error) {
	if v.Has(dRegexp) {
		pattern, err := v.GetString(dRegexp)
		if err != nil {
			return nil, errors.New("cannot read regexp pattern from " + dRegexp)
		}

		_pattern := string(pattern)
		var r myjson.ExtRegexp
		var ok bool
		if r, ok = f.regexpCache[_pattern]; !ok {
			r, err = regexp.Compile(_pattern)
			if err != nil {
				return nil, err
			}
		}

		return r, nil
	} else {
		rV := make(myjson.Object)
		for name, value := range v {
			rValue, err := f.parse(value)
			if err != nil {
				return nil, err
			}
			rV[name] = rValue
		}
		return rV, nil
	}
}

func (f *parseRegexpFilter) parseForArray(v myjson.Array) (interface{}, error) {
	rV := make(myjson.Array, len(v))
	for idx, value := range v {
		rValue, err := f.parse(value)
		if err != nil {
			return nil, err
		}
		rV[idx] = rValue
	}
	return rV, nil
}

func (f *parseRegexpFilter) parseForExtJSONMatcher(v myjson.ExtJSONMatcher) (interface{}, error) {
	_v := v.Unwrap()
	rV, err := f.parse(_v)
	if err != nil {
		return nil, err
	}
	return myjson.MakeExtJSONMatcher(rV), nil
}

func (f *parseRegexpFilter) reset() {
	f.regexpCache = make(map[string]myjson.ExtRegexp)
}

type jsonMatcherFilter struct {
}

func (f *jsonMatcherFilter) doFilter(v interface{}, chain *filterChain) error {
	gV, err := f.generate(v)
	if err != nil {
		return err
	}
	return chain.doFilter(gV)
}

func (f *jsonMatcherFilter) generate(v interface{}) (gV interface{}, err error) {
	switch v.(type) {
	case myjson.Object:
		gV, err = f.generateObject(v.(myjson.Object))
	case myjson.Array:
		gV, err = f.generateForArray(v.(myjson.Array))
	default:
		gV, err = v, nil
	}
	return
}

func (f *jsonMatcherFilter) generateObject(v myjson.Object) (interface{}, error) {
	if v.Has(dJSON) {
		json := v.Get(dJSON)
		raw, err := f.toRawJSONMatcher(json)
		if err != nil {
			return nil, err
		}
		return myjson.MakeExtJSONMatcher(raw), nil
	} else {
		gV := make(myjson.Object)
		for name, value := range v {
			gValue, err := f.generate(value)
			if err != nil {
				return nil, err
			}
			gV[name] = gValue
		}
		return gV, nil
	}
}

func (f *jsonMatcherFilter) generateForArray(v myjson.Array) (myjson.Array, error) {
	result := make(myjson.Array, len(v))
	for idx, value := range v {
		gValue, err := f.generate(value)
		if err != nil {
			return nil, err
		}
		result[idx] = gValue
	}
	return result, nil
}

func (f *jsonMatcherFilter) toRawJSONMatcher(v interface{}) (interface{}, error) {
	switch v.(type) {
	case myjson.Object:
		return f.objectToRawJSONMatcher(v.(myjson.Object))
	case myjson.Array:
		return f.arrayToRawJSONMatcher(v.(myjson.Array))
	default:
		return v, nil
	}
}

func (f *jsonMatcherFilter) objectToRawJSONMatcher(v myjson.Object) (myjson.Object, error) {
	result := make(myjson.Object)
	jPaths := make(map[string]interface{})
	for name, value := range v {
		gValue, err := f.generate(value)
		if err != nil {
			return nil, err
		}

		if strings.HasPrefix(name, "$$") {
			name = name[1:]
			result[name] = gValue
		} else if strings.HasPrefix(name, "$") { // treats as json-path
			jPaths[name] = gValue
		} else {
			result[name] = gValue
		}
	}

	for pStr, rValue := range jPaths {
		path, err := myjson.ParsePath(pStr)
		if err != nil {
			return nil, err
		}
		newRV, err := result.SetByPath(path, rValue)
		if err != nil {
			return nil, err
		}
		result = newRV
	}

	return result, nil
}

func (f *jsonMatcherFilter) arrayToRawJSONMatcher(v myjson.Array) (myjson.Array, error) {
	result := make(myjson.Array, len(v))
	for idx, value := range v {
		gValue, err := f.generate(value)
		if err != nil {
			return nil, err
		}
		result[idx] = gValue
	}
	return result, nil
}
