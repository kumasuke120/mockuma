package mckmaps

import (
	"fmt"
	"strings"

	"github.com/kumasuke120/mockuma/internal/myjson"
)

type renderError struct {
	filename string
	jsonPath *myjson.Path
}

func (e *renderError) Error() string {
	result := ""
	if e.jsonPath == nil {
		result += "cannot render template"
	} else {
		result += fmt.Sprintf("cannot render template on json-path '%v'", e.jsonPath)
	}

	if e.filename != "" {
		result += fmt.Sprintf(" in the file '%s'", e.filename)
	}

	return result
}

type renderContext struct {
	filename string
}

type objectContext struct {
	dstJson myjson.Object
	dstKey  string
	*renderContext
}

type arrayContext struct {
	dstJson myjson.Array
	dstIdx  int
	*renderContext
}

func (t *Template) replaceInObject(ctx *objectContext, vars []*Vars) (myjson.Object, error) {
	result := make(myjson.Object)

	v, err := t.render(ctx.renderContext, vars)
	if err != nil {
		return nil, err
	}

	for name, value := range ctx.dstJson {
		var _value interface{}

		if name == ctx.dstKey {
			if len(v) == 0 {
				continue
			} else if len(v) == 1 {
				_value = v[0]
			} else {
				_value = v
			}
		} else {
			_value = value
		}

		result[name] = _value
	}

	return result, nil
}

func (t *Template) replaceInArray(ctx *arrayContext, vars []*Vars) (myjson.Array, error) {
	result := ctx.dstJson[:ctx.dstIdx]

	v, err := t.render(ctx.renderContext, vars)
	if err != nil {
		return nil, err
	}

	for _, _v := range v {
		result = append(result, _v)
	}

	result = append(result, ctx.dstJson[ctx.dstIdx+1:]...)

	return result, nil
}

func (t *Template) render(ctx *renderContext, vars []*Vars) ([]interface{}, error) {
	if len(vars) == 0 {
		return []interface{}{}, nil
	}

	result := make([]interface{}, len(vars))
	for idx, _var := range vars {
		v, err := render(ctx, nil, t.content, _var)
		if err != nil {
			return nil, err
		}
		result[idx] = v
	}
	return result, nil
}

func render(ctx *renderContext, jsonPath *myjson.Path, v interface{}, vars *Vars) (interface{}, error) {
	if jsonPath == nil {
		jsonPath = myjson.NewPath()
	}

	var result interface{}
	var err error
	switch v.(type) {
	case myjson.Object:
		result, err = renderObject(ctx, jsonPath, v.(myjson.Object), vars)
	case myjson.Array:
		result, err = renderArray(ctx, jsonPath, v.(myjson.Array), vars)
	case myjson.Number:
		result = v
	case myjson.String:
		result, err = renderString(ctx, jsonPath, v.(myjson.String), vars)
	case myjson.Boolean:
		result = v
	}

	if err != nil {
		return nil, err
	}
	return result, nil
}

func renderObject(ctx *renderContext, jsonPath *myjson.Path,
	v myjson.Object, vars *Vars) (myjson.Object, error) {
	jsonPath.Append("")

	result := make(myjson.Object)
	for name, value := range v {
		jsonPath.SetLast(name)

		rValue, err := render(ctx, jsonPath, value, vars)
		if err != nil {
			return nil, err
		}
		result[name] = rValue
	}

	jsonPath.RemoveLast()
	return result, nil
}

func renderArray(ctx *renderContext, jsonPath *myjson.Path,
	v myjson.Array, vars *Vars) (myjson.Array, error) {
	jsonPath.Append(0)

	result := make(myjson.Array, len(v))
	for idx, value := range v {
		jsonPath.SetLast(idx)

		rValue, err := render(ctx, jsonPath, value, vars)
		if err != nil {
			return nil, err
		}
		result[idx] = rValue
	}

	jsonPath.RemoveLast()
	return result, nil
}

const (
	rsReady = iota
	rsMaybeDVar
	rsMaybeDVar2
	rsMaybeSVar
	rsInVar
)

func renderString(ctx *renderContext, jsonPath *myjson.Path,
	v myjson.String, vars *Vars) (interface{}, error) {
	s := rsReady

	runes := []rune(v)

	var fromBegin, toEnd bool

	var builder strings.Builder
	var nameBuilder strings.Builder
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		doWrite := true
		doWriteName := false

		switch s {
		case rsReady:
			if r == '$' {
				s = rsMaybeSVar

			} else if r == '@' {
				s = rsMaybeDVar
			}

			if r == '$' || r == '@' {
				if i == 0 {
					fromBegin = true
				} else {
					fromBegin = false
				}
				doWrite = false
			}
		case rsMaybeDVar:
			if string(runes[i:i+3]) == "var" {
				s = rsMaybeDVar2
				i += 3
				doWrite = false
			} else {
				s = rsReady
				if r != '@' { // replaces "@@" to "@"
					builder.WriteRune('@')
				}
				if fromBegin {
					fromBegin = false
				}
			}
		case rsMaybeDVar2:
			if r == '{' {
				s = rsInVar
				doWrite = false
			} else {
				s = rsReady
				builder.WriteString("@var")
				if fromBegin {
					fromBegin = false
				}
			}
		case rsMaybeSVar:
			if r == '{' {
				s = rsInVar
				doWrite = false
			} else {
				s = rsReady
				if r != '$' { // replaces "$$" to "$"
					builder.WriteString("$")
				}
				if fromBegin {
					fromBegin = false
				}
			}
		case rsInVar:
			doWrite = false
			if r == '}' {
				s = rsReady
				if fromBegin && i == len(runes)-1 {
					toEnd = true
				} else {
					varName := nameBuilder.String()
					v, err := renderTextString(ctx, jsonPath, vars, varName)
					if err != nil {
						return "", err
					}
					builder.WriteString(v)
					nameBuilder.Reset()
				}
			} else {
				doWriteName = true
			}
		}

		if doWrite {
			builder.WriteRune(r)
		}
		if doWriteName {
			nameBuilder.WriteRune(r)
		}
	}

	if s != rsReady {
		return "", &renderError{filename: ctx.filename, jsonPath: jsonPath}
	}

	if fromBegin && toEnd {
		varName := nameBuilder.String()
		if varName == "" {
			return "", &renderError{filename: ctx.filename, jsonPath: jsonPath}
		}

		varV := vars.table[varName]
		return varV, nil
	}

	return myjson.String(builder.String()), nil
}

func renderTextString(ctx *renderContext, jsonPath *myjson.Path,
	vars *Vars, varName string) (string, error) {
	if varName == "" {
		return "", &renderError{filename: ctx.filename, jsonPath: jsonPath}
	}

	varV := vars.table[varName]
	switch varV.(type) {
	case myjson.String:
		return fmt.Sprintf("%v", varV), nil
	case myjson.Number:
		return fmt.Sprintf("%v", varV), nil
	case myjson.Boolean:
		return fmt.Sprintf("%v", varV), nil
	default:
		return "", &renderError{filename: ctx.filename, jsonPath: jsonPath}
	}
}
