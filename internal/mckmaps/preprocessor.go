package mckmaps

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kumasuke120/mockuma/internal/myjson"
	"github.com/kumasuke120/mockuma/internal/types"
)

// removes all @comment directives in mockuMappings
type dCommentProcessor struct{}

func (p *dCommentProcessor) DoFilter(v interface{}, chain *types.FilterChain) error {
	p.removeComment(v)
	return chain.DoFilter(v)
}

func (p *dCommentProcessor) removeComment(v interface{}) {
	switch v.(type) {
	case myjson.Object:
		delete(v.(myjson.Object), dComment)
		for _, value := range v.(myjson.Object) {
			p.removeComment(value)
		}
	case myjson.Array:
		for _, _v := range v.(myjson.Array) {
			p.removeComment(_v)
		}
	}
}

func makeDFileProcessor() *dFileProcessor {
	p := new(dFileProcessor)
	p.reset()
	return p
}

// loads file contents for @file
type dFileProcessor struct {
	fileCache map[string][]byte
}

func (p *dFileProcessor) DoFilter(v interface{}, chain *types.FilterChain) error {
	rV, err := p.load(v)
	if err != nil {
		return err
	}
	return chain.DoFilter(rV)
}

func (p *dFileProcessor) load(v interface{}) (interface{}, error) {
	var rV interface{}
	var err error
	switch v.(type) {
	case myjson.Object:
		rV, err = p.loadForObject(v.(myjson.Object))
	case myjson.Array:
		rV, err = p.loadForArray(v.(myjson.Array))
	default:
		rV, err = v, nil
	}
	return rV, err
}

func (p *dFileProcessor) loadForObject(v myjson.Object) (interface{}, error) {
	if v.Has(dFile) {
		filename, err := v.GetString(dFile)
		if err != nil {
			return nil, errors.New("cannot read filename from " + dFile)
		}

		_filename := string(filename)
		var bytes []byte
		var ok bool
		if bytes, ok = p.fileCache[_filename]; !ok {
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
			rValue, err := p.load(value)
			if err != nil {
				return nil, err
			}
			rV[name] = rValue
		}
		return rV, nil
	}
}

func (p *dFileProcessor) loadForArray(v myjson.Array) (interface{}, error) {
	rV := make(myjson.Array, len(v))
	for idx, value := range v {
		rValue, err := p.load(value)
		if err != nil {
			return nil, err
		}
		rV[idx] = rValue
	}
	return rV, nil
}

func (p *dFileProcessor) reset() {
	p.fileCache = make(map[string][]byte)
}

func makeDRegexpProcessor() *dRegexpProcessor {
	p := new(dRegexpProcessor)
	p.reset()
	return p
}

// converts all @regexp directives to ExtRegexp
type dRegexpProcessor struct {
	regexpCache map[string]myjson.ExtRegexp
}

func (p *dRegexpProcessor) DoFilter(v interface{}, chain *types.FilterChain) error {
	rV, err := p.parse(v)
	if err != nil {
		return err
	}
	return chain.DoFilter(rV)
}

func (p *dRegexpProcessor) parse(v interface{}) (rV interface{}, err error) {
	switch v.(type) {
	case myjson.Object:
		rV, err = p.parseForObject(v.(myjson.Object))
	case myjson.Array:
		rV, err = p.parseForArray(v.(myjson.Array))
	case myjson.ExtJSONMatcher:
		rV, err = p.parseForExtJSONMatcher(v.(myjson.ExtJSONMatcher))
	default:
		rV, err = v, nil
	}
	return
}

func (p *dRegexpProcessor) parseForObject(v myjson.Object) (interface{}, error) {
	if v.Has(dRegexp) {
		pattern, err := v.GetString(dRegexp)
		if err != nil {
			return nil, errors.New("cannot read regexp pattern from " + dRegexp)
		}

		_pattern := string(pattern)
		var r myjson.ExtRegexp
		var ok bool
		if r, ok = p.regexpCache[_pattern]; !ok {
			r, err = regexp.Compile(_pattern)
			if err != nil {
				return nil, err
			}
		}

		return r, nil
	} else {
		rV := make(myjson.Object)
		for name, value := range v {
			rValue, err := p.parse(value)
			if err != nil {
				return nil, err
			}
			rV[name] = rValue
		}
		return rV, nil
	}
}

func (p *dRegexpProcessor) parseForArray(v myjson.Array) (interface{}, error) {
	rV := make(myjson.Array, len(v))
	for idx, value := range v {
		rValue, err := p.parse(value)
		if err != nil {
			return nil, err
		}
		rV[idx] = rValue
	}
	return rV, nil
}

func (p *dRegexpProcessor) parseForExtJSONMatcher(v myjson.ExtJSONMatcher) (interface{}, error) {
	_v := v.Unwrap()
	rV, err := p.parse(_v)
	if err != nil {
		return nil, err
	}
	return myjson.MakeExtJSONMatcher(rV), nil
}

func (p *dRegexpProcessor) reset() {
	p.regexpCache = make(map[string]myjson.ExtRegexp)
}

// converts all @json directives to ExtJSONMatcher
type dJSONProcessor struct{}

func (p *dJSONProcessor) DoFilter(v interface{}, chain *types.FilterChain) error {
	gV, err := p.generate(v)
	if err != nil {
		return err
	}
	return chain.DoFilter(gV)
}

func (p *dJSONProcessor) generate(v interface{}) (gV interface{}, err error) {
	switch v.(type) {
	case myjson.Object:
		gV, err = p.generateObject(v.(myjson.Object))
	case myjson.Array:
		gV, err = p.generateForArray(v.(myjson.Array))
	default:
		gV, err = v, nil
	}
	return
}

func (p *dJSONProcessor) generateObject(v myjson.Object) (interface{}, error) {
	if v.Has(dJSON) {
		json := v.Get(dJSON)
		raw, err := p.toRawJSONMatcher(json)
		if err != nil {
			return nil, err
		}
		return myjson.MakeExtJSONMatcher(raw), nil
	} else {
		gV := make(myjson.Object)
		for name, value := range v {
			gValue, err := p.generate(value)
			if err != nil {
				return nil, err
			}
			gV[name] = gValue
		}
		return gV, nil
	}
}

func (p *dJSONProcessor) generateForArray(v myjson.Array) (myjson.Array, error) {
	result := make(myjson.Array, len(v))
	for idx, value := range v {
		gValue, err := p.generate(value)
		if err != nil {
			return nil, err
		}
		result[idx] = gValue
	}
	return result, nil
}

func (p *dJSONProcessor) toRawJSONMatcher(v interface{}) (interface{}, error) {
	switch v.(type) {
	case myjson.Object:
		return p.objectToRawJSONMatcher(v.(myjson.Object))
	case myjson.Array:
		return p.arrayToRawJSONMatcher(v.(myjson.Array))
	default:
		return v, nil
	}
}

func (p *dJSONProcessor) objectToRawJSONMatcher(v myjson.Object) (myjson.Object, error) {
	result := make(myjson.Object)
	jPaths := make(map[string]interface{})
	for name, value := range v {
		gValue, err := p.generate(value)
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

func (p *dJSONProcessor) arrayToRawJSONMatcher(v myjson.Array) (myjson.Array, error) {
	result := make(myjson.Array, len(v))
	for idx, value := range v {
		gValue, err := p.generate(value)
		if err != nil {
			return nil, err
		}
		result[idx] = gValue
	}
	return result, nil
}

func makeDTemplateProcessor() *dTemplateProcessor {
	p := dTemplateProcessor{}
	p.reset()
	return &p
}

// renders @template directives with given vars
type dTemplateProcessor struct {
	templateCache  map[string]*template
	varsSliceCache map[string][]*vars
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

func (p *dTemplateProcessor) DoFilter(v interface{}, chain *types.FilterChain) error {
	rV, err := p.render(v)
	if err != nil {
		return err
	}
	return chain.DoFilter(rV)
}

func (p *dTemplateProcessor) render(v interface{}) (interface{}, error) {
	var rV interface{}
	var err error
	switch v.(type) {
	case myjson.Object:
		rV, err = p.renderObject(v.(myjson.Object))
	case myjson.Array:
		rV, err = p.renderArray(v.(myjson.Array))
	default:
		rV, err = v, nil
	}
	return rV, err
}

func (p *dTemplateProcessor) renderObject(v myjson.Object) (interface{}, error) {
	if v.Has(dTemplate) { // if v is a @template directive
		template, err := p.getTemplateFromDTemplate(v)
		if err != nil {
			return nil, err
		}
		varsSlice, err := p.getVarsFromDTemplate(v)
		if err != nil {
			return nil, err
		}

		rV, err := template.renderAll(varsSlice)
		if err != nil {
			return nil, err
		}
		return fromTemplate{rV: rV}, nil
	} else {
		result := make(myjson.Object, len(v))
		for name, value := range v {
			rValue, err := p.render(value)
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

func (p *dTemplateProcessor) renderArray(v myjson.Array) (myjson.Array, error) {
	var result myjson.Array
	for _, value := range v {
		rValue, err := p.render(value)
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

func (p *dTemplateProcessor) getTemplateFromDTemplate(v myjson.Object) (*template, error) {
	filename, err := v.GetString(dTemplate)
	if err != nil {
		return nil, errors.New("cannot read the name of template file")
	}

	var template *template
	var ok bool
	_filename := string(filename)
	if template, ok = p.templateCache[_filename]; !ok {
		tParser := &templateParser{Parser: Parser{filename: _filename}}
		template, err = tParser.parse()
		if err != nil {
			return nil, err
		}
		p.templateCache[_filename] = template
	}
	return template, nil
}

func (p *dTemplateProcessor) getVarsFromDTemplate(v myjson.Object) (varsSlice []*vars, err error) {
	if v.Has(tVars) { // if @template directive has a 'vars' attribute
		varsSlice, err = new(varsJSONParser).parseVars(v)
	} else {
		var filename myjson.String
		filename, err = v.GetString(dVars)
		if err != nil {
			err = errors.New("cannot read filename from " + dVars)
			return
		}

		var ok bool
		_filename := string(filename)
		if varsSlice, ok = p.varsSliceCache[_filename]; !ok {
			ext := filepath.Ext(_filename)
			if ext == ".csv" {
				vParser := &varsCSVParser{Parser: Parser{filename: _filename}}
				varsSlice, err = vParser.parse()
			} else {
				vParser := &varsJSONParser{Parser: Parser{filename: _filename}}
				varsSlice, err = vParser.parse()
			}
			p.varsSliceCache[_filename] = varsSlice
		}
	}
	return
}

func (p *dTemplateProcessor) reset() { // clears caches
	p.templateCache = make(map[string]*template)
	p.varsSliceCache = make(map[string][]*vars)
}
