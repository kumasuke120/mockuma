package mckmaps

import (
	"errors"
	"path/filepath"

	"github.com/kumasuke120/mockuma/internal/myjson"
)

type template struct {
	content  interface{}
	filename string
}

type templateParser struct {
	json     myjson.Object
	jsonPath *myjson.Path
	Parser
}

func (p *templateParser) parse() (*template, error) {
	if p.json == nil {
		json, err := p.load(true, ppRemoveComment)
		if err != nil {
			return nil, err
		}

		switch json.(type) {
		case myjson.Object:
			p.json = json.(myjson.Object)
		default:
			return nil, newParserError(p.filename, p.jsonPath)
		}
	}

	p.jsonPath = myjson.NewPath("")
	p.jsonPath.SetLast(dType)
	_type, err := p.json.GetString(dType)
	if err != nil || string(_type) != tTemplate {
		return nil, newParserError(p.filename, p.jsonPath)
	}

	template := new(template)

	p.jsonPath.SetLast(tTemplate)
	v := p.json.Get(tTemplate)
	switch v.(type) {
	case myjson.Object:
		template.content = v
	case myjson.Array:
		template.content = v
	case myjson.String:
		template.content = v
	default:
		return nil, newParserError(p.filename, p.jsonPath)
	}
	template.filename = p.filename

	return template, nil
}

// renders @template directives with given vars
type templateFilter struct {
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

func (f *templateFilter) doFilter(v interface{}, chain *filterChain) error {
	rV, err := f.render(v)
	if err != nil {
		return err
	}
	return chain.doFilter(rV)
}

func (f *templateFilter) render(v interface{}) (interface{}, error) {
	var rV interface{}
	var err error
	switch v.(type) {
	case myjson.Object:
		rV, err = f.renderObject(v.(myjson.Object))
	case myjson.Array:
		rV, err = f.renderArray(v.(myjson.Array))
	default:
		rV, err = v, nil
	}
	return rV, err
}

func (f *templateFilter) renderObject(v myjson.Object) (interface{}, error) {
	if v.Has(dTemplate) { // if v is a @template directive
		template, err := f.getTemplateFromDTemplate(v)
		if err != nil {
			return nil, err
		}
		varsSlice, err := f.getVarsFromDTemplate(v)
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
			rValue, err := f.render(value)
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

func (f *templateFilter) renderArray(v myjson.Array) (myjson.Array, error) {
	var result myjson.Array
	for _, value := range v {
		rValue, err := f.render(value)
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

func (f *templateFilter) getTemplateFromDTemplate(v myjson.Object) (*template, error) {
	filename, err := v.GetString(dTemplate)
	if err != nil {
		return nil, errors.New("cannot read the name of template file")
	}

	var template *template
	var ok bool
	_filename := string(filename)
	if template, ok = f.templateCache[_filename]; !ok {
		tParser := &templateParser{Parser: Parser{filename: _filename}}
		template, err = tParser.parse()
		if err != nil {
			return nil, err
		}
		f.templateCache[_filename] = template
	}
	return template, nil
}

func (f *templateFilter) getVarsFromDTemplate(v myjson.Object) (varsSlice []*vars, err error) {
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
		if varsSlice, ok = f.varsSliceCache[_filename]; !ok {
			ext := filepath.Ext(_filename)
			if ext == ".csv" {
				vParser := &varsCSVParser{Parser: Parser{filename: _filename}}
				varsSlice, err = vParser.parse()
			} else {
				vParser := &varsJSONParser{Parser: Parser{filename: _filename}}
				varsSlice, err = vParser.parse()
			}
			f.varsSliceCache[_filename] = varsSlice
		}
	}
	return
}

func (f *templateFilter) reset() { // clears caches
	f.templateCache = make(map[string]*template)
	f.varsSliceCache = make(map[string][]*vars)
}
