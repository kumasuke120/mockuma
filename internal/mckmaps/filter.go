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
	ppToJsonMatcher  = &jsonMatcherFilter{}
)

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

type templateFilter struct { // rewrites @template directives with given vars
	templateCache  map[string]*template
	varsSliceCache map[string][]*vars
}

func makeTemplateFilter() *templateFilter {
	f := templateFilter{}
	f.reset()
	return &f
}

func (f *templateFilter) doFilter(v interface{}, chain *filterChain) error {
	rV, err := f.rewrite(v)
	if err != nil {
		return err
	}
	return chain.doFilter(rV)
}

func (f *templateFilter) rewrite(v interface{}) (interface{}, error) {
	var rV interface{}
	var err error
	switch v.(type) {
	case myjson.Object:
		rV, err = f.rewriteObject(v.(myjson.Object))
	case myjson.Array:
		rV, err = f.rewriteArray(v.(myjson.Array))
	default:
		rV, err = v, nil
	}
	return rV, err
}

func (f *templateFilter) rewriteObject(v myjson.Object) (interface{}, error) {
	if v.Has(dTemplate) { // if v is a @template directive
		template, ctx, err := f.getTemplateFromDTemplate(v)
		if err != nil {
			return nil, err
		}
		varsSlice, err := f.getVarsFromDTemplate(v)
		if err != nil {
			return nil, err
		}

		rV, err := template.render(ctx, varsSlice)
		if err != nil {
			return nil, err
		}
		return fromTemplate{rV: rV}, nil
	} else {
		result := make(myjson.Object, len(v))
		for name, value := range v {
			rValue, err := f.rewrite(value)
			if err != nil {
				return nil, err
			}

			switch rValue.(type) {
			case fromTemplate:
				rValue = rValue.(fromTemplate).forObject()
			}

			result[name] = rValue
		}
		return result, nil
	}
}

func (f *templateFilter) rewriteArray(v myjson.Array) (myjson.Array, error) {
	var result myjson.Array
	for _, value := range v {
		rValue, err := f.rewrite(value)
		if err != nil {
			return nil, err
		}

		switch rValue.(type) {
		case fromTemplate:
			for _, _rValue := range rValue.(fromTemplate).forArray() {
				result = append(result, _rValue)
			}
		default:
			result = append(result, rValue)
		}
	}
	return result, nil
}

func (f *templateFilter) getTemplateFromDTemplate(v myjson.Object) (*template, *renderContext, error) {
	filename, err := v.GetString(dTemplate)
	if err != nil {
		return nil, nil, errors.New("cannot read the name of template file")
	}

	var template *template
	var ok bool
	_filename := string(filename)
	if template, ok = f.templateCache[_filename]; !ok {
		tParser := &templateParser{parser: parser{filename: _filename}}
		template, err = tParser.parse()
		if err != nil {
			return nil, nil, err
		}
	}
	return template, &renderContext{filename: _filename}, nil
}

func (f *templateFilter) getVarsFromDTemplate(v myjson.Object) ([]*vars, error) {
	var varsSlice []*vars
	var err error
	if v.Has(tVars) { // if @template directive has a 'vars' attribute
		varsSlice, err = new(varsParser).parseVars(v)
	} else {
		filename, err := v.GetString(dVars)
		if err != nil {
			return nil, errors.New("cannot read filename from " + dVars)
		}

		var ok bool
		_filename := string(filename)
		if varsSlice, ok = f.varsSliceCache[_filename]; !ok {
			vParser := &varsParser{parser: parser{filename: _filename}}
			varsSlice, err = vParser.parse()
			f.varsSliceCache[_filename] = varsSlice
		}
	}
	return varsSlice, err
}

func (f *templateFilter) reset() { // clears caches
	f.templateCache = make(map[string]*template)
	f.varsSliceCache = make(map[string][]*vars)
}

type fromTemplate struct {
	rV myjson.Array
}

func (ft fromTemplate) forObject() interface{} {
	if len(ft.rV) == 0 {
		return nil
	} else if len(ft.rV) == 1 {
		return ft.rV[0]
	} else {
		return ft.rV
	}
}

func (ft fromTemplate) forArray() []interface{} {
	return ft.rV
}

type loadFileFilter struct { // loads file contents for @file
	fileCache map[string][]byte
}

func makeLoadFileFilter() *loadFileFilter {
	f := new(loadFileFilter)
	f.reset()
	return f
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
			bytes, err = ioutil.ReadFile(_filename)
			if err != nil {
				return nil, err
			}
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

func (f *loadFileFilter) loadForExtJsonMatcher(v myjson.ExtJsonMatcher) (interface{}, error) {
	_v := v.Unwrap()
	rV, err := f.load(_v)
	if err != nil {
		return nil, err
	}
	return myjson.MakeExtJsonMatcher(rV), nil
}

func (f *loadFileFilter) reset() {
	f.fileCache = make(map[string][]byte)
}

type parseRegexpFilter struct {
	regexpCache map[string]myjson.ExtRegexp
}

func makeParseRegexpFilter() *parseRegexpFilter {
	f := new(parseRegexpFilter)
	f.reset()
	return f
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
	case myjson.ExtJsonMatcher:
		rV, err = f.parseForExtJsonMatcher(v.(myjson.ExtJsonMatcher))
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

func (f *parseRegexpFilter) parseForExtJsonMatcher(v myjson.ExtJsonMatcher) (interface{}, error) {
	_v := v.Unwrap()
	rV, err := f.parse(_v)
	if err != nil {
		return nil, err
	}
	return myjson.MakeExtJsonMatcher(rV), nil
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
	if v.Has(dJson) {
		json := v.Get(dJson)
		raw, err := f.toRawJsonMatcher(json)
		if err != nil {
			return nil, err
		}
		return myjson.MakeExtJsonMatcher(raw), nil
	} else {
		rV := make(myjson.Object)
		for name, value := range v {
			rValue, err := f.generate(value)
			if err != nil {
				return nil, err
			}
			rV[name] = rValue
		}
		return rV, nil
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

func (f *jsonMatcherFilter) toRawJsonMatcher(v interface{}) (interface{}, error) {
	switch v.(type) {
	case myjson.Object:
		return f.objectToRawJsonMatcher(v.(myjson.Object))
	case myjson.Array:
		return f.arrayToRawJsonMatcher(v.(myjson.Array))
	default:
		return v, nil
	}
}

func (f *jsonMatcherFilter) objectToRawJsonMatcher(v myjson.Object) (myjson.Object, error) {
	result := make(myjson.Object)
	jPaths := make(map[string]interface{})
	for name, value := range v {
		rValue, err := f.toRawJsonMatcher(value)
		if err != nil {
			return nil, err
		}

		if strings.HasPrefix(name, "$$") {
			name = name[1:]
			result[name] = rValue
		} else if strings.HasPrefix(name, "$") { // treats as json-path
			jPaths[name] = rValue
		} else {
			result[name] = rValue
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

func (f *jsonMatcherFilter) arrayToRawJsonMatcher(v myjson.Array) (myjson.Array, error) {
	result := make(myjson.Array, len(v))
	for idx, value := range v {
		rValue, err := f.toRawJsonMatcher(value)
		if err != nil {
			return nil, err
		}
		result[idx] = rValue
	}
	return result, nil
}
