package mckmaps

import (
	"errors"

	"github.com/kumasuke120/mockuma/internal/myjson"
)

var (
	fRemoveComment  = commentFilter{}
	fRenderTemplate = templateFilter{}
)

type filterChain struct {
	filters []filter
	idx     int
	v       interface{}
}

func doFiltersOnV(v interface{}, filters ...filter) (interface{}, error) {
	chain := &filterChain{filters: filters}
	err := chain.doFilter(v)
	if err != nil {
		return nil, err
	}
	return chain.v, nil
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

type commentFilter struct {
}

func (f commentFilter) doFilter(v interface{}, chain *filterChain) error {
	f.removeComment(v)
	return chain.doFilter(v)
}

func (f commentFilter) removeComment(v interface{}) {
	switch v.(type) {
	case myjson.Object:
		delete(v.(myjson.Object), dComment)
	case myjson.Array:
		for _, _v := range v.(myjson.Array) {
			f.removeComment(_v)
		}
	}
}

type templateFilter struct {
	templateCache  map[string]*Template
	varsSliceCache map[string][]*Vars
}

func (f templateFilter) doFilter(v interface{}, chain *filterChain) error {
	renderV, err := f.rewrite(v)
	if err != nil {
		return err
	}
	return chain.doFilter(renderV)
}

func (f templateFilter) rewrite(v interface{}) (interface{}, error) {
	var renderV interface{}
	var err error
	switch v.(type) {
	case myjson.Object:
		renderV, err = f.rewriteObject(v.(myjson.Object))
	case myjson.Array:
		renderV, err = f.rewriteArray(v.(myjson.Array))
	default:
		renderV, err = v, nil
	}
	return renderV, err
}

func (f templateFilter) rewriteObject(v myjson.Object) (interface{}, error) {
	if v.Has(dTemplate) {
		template, ctx, err := f.getTemplateFromDTemplate(v)
		if err != nil {
			return nil, err
		}
		varsSlice, err := f.getVarsFromDTemplate(v)
		if err != nil {
			return nil, err
		}

		renderV, err := template.render(ctx, varsSlice)
		if err != nil {
			return nil, err
		}
		return fromTemplate{renderV: renderV}, nil
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

func (f templateFilter) rewriteArray(v myjson.Array) (myjson.Array, error) {
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

func (f templateFilter) getTemplateFromDTemplate(v myjson.Object) (*Template, *renderContext, error) {
	filename, err := v.GetString(dTemplate)
	if err != nil {
		return nil, nil, errors.New("cannot read the name of template file")
	}

	var template *Template
	var ok bool
	_filename := string(filename)
	if template, ok = f.templateCache[_filename]; !ok {
		tParser := &templateParser{parser: parser{filename: _filename}}
		template, err = tParser.parse()
		if err != nil {
			return nil, &renderContext{filename: _filename}, err
		}
	}
	return template, nil, nil
}

func (f templateFilter) getVarsFromDTemplate(v myjson.Object) ([]*Vars, error) {
	var varsSlice []*Vars
	var err error
	if v.Has(tVars) {
		varsSlice, err = new(varsParser).parseVars(v)
	} else {
		filename, err := v.GetString(dVars)
		if err != nil {
			return nil, errors.New("cannot read the name of vars file")
		}

		_filename := string(filename)
		if varsSlice, ok := f.varsSliceCache[_filename]; !ok {
			vParser := &varsParser{parser: parser{filename: _filename}}
			varsSlice, err = vParser.parse()
			f.varsSliceCache[_filename] = varsSlice
		}
	}
	return varsSlice, err
}

type fromTemplate struct {
	renderV []interface{}
}

func (ft fromTemplate) forObject() interface{} {
	if len(ft.renderV) == 0 {
		return nil
	} else if len(ft.renderV) == 1 {
		return ft.renderV[0]
	} else {
		return ft.renderV
	}
}

func (ft fromTemplate) forArray() []interface{} {
	return ft.renderV
}
