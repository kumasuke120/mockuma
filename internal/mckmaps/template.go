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
		result += fmt.Sprintf("cannot render the template on json-path '%v'", e.jsonPath)
	}

	if e.filename != "" {
		result += fmt.Sprintf(" in the file '%s'", e.filename)
	}

	return result
}

type renderContext struct {
	filename string
}

type template struct {
	content interface{}
}

type vars struct {
	table map[string]interface{}
}

func (t *template) render(ctx *renderContext, varsSlice []*vars) (myjson.Array, error) {
	if len(varsSlice) == 0 {
		return myjson.Array{}, nil
	}

	result := make(myjson.Array, len(varsSlice))
	for idx, _var := range varsSlice {
		v, err := render(ctx, nil, t.content, _var)
		if err != nil {
			return nil, err
		}
		result[idx] = v
	}
	return result, nil
}

func render(ctx *renderContext, jsonPath *myjson.Path, v interface{}, varsSlice *vars) (interface{}, error) {
	if jsonPath == nil {
		jsonPath = myjson.NewPath()
	}

	var result interface{}
	var err error
	switch v.(type) {
	case myjson.Object:
		result, err = renderObject(ctx, jsonPath, v.(myjson.Object), varsSlice)
	case myjson.Array:
		result, err = renderArray(ctx, jsonPath, v.(myjson.Array), varsSlice)
	case myjson.String:
		result, err = renderString(ctx, jsonPath, v.(myjson.String), varsSlice)
	default:
		result, err = v, nil
	}

	return result, err
}

func renderObject(ctx *renderContext, jsonPath *myjson.Path,
	v myjson.Object, vars *vars) (myjson.Object, error) {
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
	v myjson.Array, vars *vars) (myjson.Array, error) {
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

// states for rendering string
const (
	rsReady = iota
	rsMaybeVar
	rsInVar
)

// tokens for rendering string
const (
	placeholderPrefix = '@'
	placeholderLeft   = '{'
	placeholderRight  = '}'
)

func renderString(ctx *renderContext, jsonPath *myjson.Path,
	v myjson.String, vars *vars) (interface{}, error) {
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
			if r == placeholderPrefix {
				s = rsMaybeVar
				if i == 0 {
					fromBegin = true
				} else {
					fromBegin = false
				}
				doWrite = false
			}
		case rsMaybeVar:
			if r == placeholderLeft {
				s = rsInVar
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
		case rsInVar:
			doWrite = false
			if r == placeholderRight {
				s = rsReady
				if fromBegin && i == len(runes)-1 {
					toEnd = true
				} else {
					varName := nameBuilder.String()
					v, err := renderTextString(ctx, jsonPath, vars, varName)
					if err != nil {
						return nil, err
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

	if s != rsReady { // placeholder is not complete
		return nil, &renderError{filename: ctx.filename, jsonPath: jsonPath}
	}

	if fromBegin && toEnd { // if the whole string is a placeholder
		varName := nameBuilder.String()
		if varName == "" {
			return nil, &renderError{filename: ctx.filename, jsonPath: jsonPath}
		}

		varV := vars.table[varName]
		return varV, nil
	}

	return myjson.String(builder.String()), nil
}

func renderTextString(ctx *renderContext, jsonPath *myjson.Path,
	vars *vars, varName string) (string, error) {
	if varName == "" {
		return "", &renderError{filename: ctx.filename, jsonPath: jsonPath}
	}

	varV := vars.table[varName]
	switch varV.(type) {
	case myjson.String:
		return fmt.Sprintf("%s", string(varV.(myjson.String))), nil
	case myjson.Number:
		return fmt.Sprintf("%v", varV), nil
	case myjson.Boolean:
		return fmt.Sprintf("%v", varV), nil
	default:
		return "", &renderError{filename: ctx.filename, jsonPath: jsonPath}
	}
}
