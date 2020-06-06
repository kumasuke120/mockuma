package mckmaps

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/kumasuke120/mockuma/internal/myjson"
	"github.com/kumasuke120/mockuma/internal/types"
)

type template struct {
	content  interface{}
	defaults *vars
	filename string
}

type templateParser struct {
	json     myjson.Object
	jsonPath *myjson.Path
	Parser
}

var parsingTemplates []string

func (p *templateParser) parse() (*template, error) {
	needLoading := p.json == nil

	if needLoading { // adds the current file
		parsingTemplates = append(parsingTemplates, p.filename)
		err := p.checkCyclicReference()
		if err != nil {
			return nil, err
		}
	}

	if needLoading { // loads file raw content
		json, err := p.load(true, ppRemoveComment, ppRenderTemplate)
		if err != nil {
			return nil, err
		}

		switch json.(type) {
		case myjson.Object:
			p.json = json.(myjson.Object)
		default:
			return nil, p.newJSONParseError(p.jsonPath)
		}
	}

	p.jsonPath = myjson.NewPath("")
	p.jsonPath.SetLast(aType)
	_type, err := p.json.GetString(aType)
	if err != nil || string(_type) != tTemplate {
		return nil, p.newJSONParseError(p.jsonPath)
	}

	template := new(template)

	p.jsonPath.SetLast(tTemplate)
	v := p.json.Get(tTemplate)
	content, err := p.parseContent(v)
	if err != nil {
		return nil, err
	}
	template.content = content

	p.jsonPath.SetLast(tVars)
	var defaults *vars
	if p.json.Has(tVars) {
		vo, err := p.json.GetObject(tVars)
		if err != nil {
			return nil, p.newJSONParseError(p.jsonPath)
		}
		defaults, err = p.parseDefaults(vo)
		if err != nil {
			return nil, err
		}
	} else {
		defaults = &vars{table: map[string]interface{}{}}
	}
	template.defaults = defaults

	template.filename = p.filename

	if needLoading { // removes the current file with checking
		if parsingTemplates[len(parsingTemplates)-1] == p.filename {
			parsingTemplates = parsingTemplates[:len(parsingTemplates)-1]
		} else {
			panic("Shouldn't happen")
		}
	}
	return template, nil
}

func (p *templateParser) parseContent(v interface{}) (interface{}, error) {
	switch v.(type) {
	case myjson.Object:
		return v, nil
	case myjson.Array:
		return v, nil
	case myjson.String:
		return v, nil
	default:
		return nil, p.newJSONParseError(p.jsonPath)
	}
}

func (p *templateParser) parseDefaults(v myjson.Object) (*vars, error) {
	r, err := parseVars(v)
	if err != nil {
		return nil, p.newJSONParseError(p.jsonPath)
	}
	return r, nil
}

func (p *templateParser) checkCyclicReference() error {
	found := make(map[string]bool)
	for _, t := range parsingTemplates {
		if _, ok := found[t]; ok {
			return &loadError{
				filename: p.filename,
				err: errors.New("found a cyclic template application : " +
					strings.Join(parsingTemplates, " -> ")),
			}
		} else {
			found[t] = true
		}
	}
	return nil
}

type renderError struct {
	filename string
	jsonPath *myjson.Path
}

func (e *renderError) Error() string {
	result := ""
	if e.jsonPath == nil {
		result += "cannot render template"
	} else {
		result += fmt.Sprintf("cannot render the template on json-path \"%v\"", e.jsonPath)
	}

	if e.filename != "" {
		result += fmt.Sprintf(" in the file '%s'", e.filename)
	}

	return result
}

func (t *template) renderAll(varsSlice []*vars) (myjson.Array, error) {
	if len(varsSlice) == 0 {
		return myjson.Array{}, nil
	}

	result := make(myjson.Array, len(varsSlice))
	for idx, _var := range varsSlice {
		v, err := t.render(nil, t.content, _var)
		if err != nil {
			return nil, err
		}
		result[idx] = v
	}
	return result, nil
}

func (t *template) render(jsonPath *myjson.Path, v interface{}, varsSlice *vars) (interface{}, error) {
	if jsonPath == nil {
		jsonPath = myjson.NewPath()
	}

	var result interface{}
	var err error
	switch v.(type) {
	case myjson.Object:
		result, err = t.renderObject(jsonPath, v.(myjson.Object), varsSlice)
	case myjson.Array:
		result, err = t.renderArray(jsonPath, v.(myjson.Array), varsSlice)
	case myjson.String:
		result, err = t.renderString(jsonPath, v.(myjson.String), varsSlice)
	default:
		result, err = v, nil
	}

	return result, err
}

func (t *template) renderObject(jsonPath *myjson.Path,
	v myjson.Object, vars *vars) (myjson.Object, error) {
	jsonPath.Append("")

	result := make(myjson.Object)
	for name, value := range v {
		jsonPath.SetLast(name)

		rName, err := t.renderPlainString(jsonPath, name, vars)
		if err != nil {
			return nil, err
		}
		rValue, err := t.render(jsonPath, value, vars)
		if err != nil {
			return nil, err
		}
		result[rName] = rValue
	}

	jsonPath.RemoveLast()
	return result, nil
}

func (t *template) renderArray(jsonPath *myjson.Path,
	v myjson.Array, vars *vars) (myjson.Array, error) {
	jsonPath.Append(0)

	result := make(myjson.Array, len(v))
	for idx, value := range v {
		jsonPath.SetLast(idx)

		rValue, err := t.render(jsonPath, value, vars)
		if err != nil {
			return nil, err
		}
		result[idx] = rValue
	}

	jsonPath.RemoveLast()
	return result, nil
}

// states for rendering string
const (
	rsReady = iota
	rsMaybePlaceholder
	rsInPlaceholder
	rsMaybePlaceHolderFormat
	rsInPlaceHolderFormat
)

// tokens for rendering string
const (
	placeholderPrefix          = '@'
	placeholderLeft            = '{'
	placeholderRight           = '}'
	placeholderFormatSeparator = ':'
)

func (t *template) renderPlainString(jsonPath *myjson.Path,
	v string, vars *vars) (string, error) {
	r, err := t.renderString(jsonPath, myjson.String(v), vars)
	if err != nil {
		return "", err
	} else {
		switch r.(type) {
		case myjson.String:
			return string(r.(myjson.String)), nil
		default:
			return types.ToString(r), nil
		}
	}
}

func (t *template) renderString(jsonPath *myjson.Path, v myjson.String, vars *vars) (interface{}, error) {
	s := rsReady

	runes := []rune(v)

	var fromBegin, toEnd bool

	var builder strings.Builder
	var nameBuilder strings.Builder
	var formatBuilder strings.Builder

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		doWrite := true
		doWriteName := false
		doWriteFormat := false

		switch s {
		case rsReady:
			if r == placeholderPrefix {
				s = rsMaybePlaceholder
				if i == 0 {
					fromBegin = true
				} else {
					fromBegin = false
				}
				doWrite = false
			}
		case rsMaybePlaceholder:
			if r == placeholderLeft {
				s = rsInPlaceholder
				doWrite = false
			} else {
				s = rsReady
				if r != placeholderPrefix { // replaces "@@" to "@"
					builder.WriteString(string(placeholderPrefix))
				}
				if fromBegin {
					fromBegin = false
				}
			}
		case rsInPlaceholder:
			doWrite = false
			if r == placeholderRight {
				s = rsReady
				if fromBegin && i == len(runes)-1 {
					toEnd = true
				} else {
					varName := nameBuilder.String()
					varFormat := formatBuilder.String()
					if varName == "" {
						return "", &renderError{filename: t.filename, jsonPath: jsonPath}
					}

					v, err := t.renderTextString(vars, varName, varFormat)
					if err != nil {
						return nil, &renderError{filename: t.filename, jsonPath: jsonPath}
					}
					builder.WriteString(v)
					nameBuilder.Reset()
					formatBuilder.Reset()
				}
			} else if r == placeholderFormatSeparator {
				s = rsMaybePlaceHolderFormat
			} else {
				doWriteName = true
			}
		case rsMaybePlaceHolderFormat:
			doWrite = false
			if r == placeholderRight { // same as empty format, state rolls back
				s = rsInPlaceholder
			} else {
				s = rsInPlaceHolderFormat
			}
			i -= 1 // goes back for other state to process
		case rsInPlaceHolderFormat:
			doWrite = false
			if r == placeholderRight { // end of placeholder
				s = rsInPlaceholder
				i -= 1
			} else {
				doWriteFormat = true
			}
		}

		if doWrite {
			builder.WriteRune(r)
		}
		if doWriteName {
			nameBuilder.WriteRune(r)
		}
		if doWriteFormat {
			formatBuilder.WriteRune(r)
		}
	}

	if s != rsReady { // placeholder is not complete
		return nil, &renderError{filename: t.filename, jsonPath: jsonPath}
	}

	if fromBegin && toEnd { // if the whole string is a placeholder
		varName := nameBuilder.String()
		if varName == "" {
			return nil, &renderError{filename: t.filename, jsonPath: jsonPath}
		}

		if vVal, vSet := t.getVarValue(vars, varName); vSet {
			return vVal, nil
		} else {
			// undefined var
			return nil, &renderError{filename: t.filename, jsonPath: jsonPath}
		}
	}

	return myjson.String(builder.String()), nil
}

var validVarFormat = regexp.MustCompile("^%([-+@0 ])?(\\d+)?\\.?(\\d+)?[tdeEfgsqxX]$")

func (t *template) renderTextString(vars *vars, varName string, varFormat string) (string, error) {
	if varFormat != "" && !validVarFormat.MatchString(varFormat) {
		return "", errors.New("invalid format for var")
	}

	vVal, vSet := t.getVarValue(vars, varName)
	if vSet && vVal == nil { // value set but is nil, e.g. { "key": null }
		vVal = myjson.String("")
	}
	switch vVal.(type) {
	case myjson.String:
		if varFormat == "" {
			varFormat = "%s"
		}
		return fmt.Sprintf(varFormat, string(vVal.(myjson.String))), nil
	case myjson.Number:
		if varFormat == "" {
			return fmt.Sprintf("%v", vVal), nil
		} else if varFormat[len(varFormat)-1] == 'd' {
			return fmt.Sprintf(varFormat, int(float64(vVal.(myjson.Number)))), nil
		} else {
			return fmt.Sprintf(varFormat, float64(vVal.(myjson.Number))), nil
		}
	case myjson.Boolean:
		if varFormat == "" {
			varFormat = "%v"
		}
		return fmt.Sprintf(varFormat, bool(vVal.(myjson.Boolean))), nil
	default:
		return "", errors.New("invalid json type for template rendering")
	}
}

func (t *template) getVarValue(vars *vars, varName string) (vVal interface{}, vSet bool) {
	if vVal, vSet = vars.table[varName]; !vSet {
		// finds in defaults if not found
		vVal, vSet = t.defaults.table[varName]
	}
	return
}
