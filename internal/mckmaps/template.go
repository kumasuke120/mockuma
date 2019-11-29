package mckmaps

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/kumasuke120/mockuma/internal/myjson"
	"github.com/kumasuke120/mockuma/internal/typeutil"
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

		rName, err := renderPlainString(ctx, jsonPath, name, vars)
		if err != nil {
			return nil, err
		}
		rValue, err := render(ctx, jsonPath, value, vars)
		if err != nil {
			return nil, err
		}
		result[rName] = rValue
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

func renderPlainString(ctx *renderContext, jsonPath *myjson.Path,
	v string, vars *vars) (string, error) {
	r, err := renderString(ctx, jsonPath, myjson.String(v), vars)
	if err != nil {
		return "", err
	} else {
		switch r.(type) {
		case myjson.String:
			return string(r.(myjson.String)), nil
		default:
			return typeutil.ToString(r), nil
		}
	}
}

func renderString(ctx *renderContext, jsonPath *myjson.Path,
	v myjson.String, vars *vars) (interface{}, error) {
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
						return "", &renderError{filename: ctx.filename, jsonPath: jsonPath}
					}

					v, err := renderTextString(vars, varName, varFormat)
					if err != nil {
						return nil, &renderError{filename: ctx.filename, jsonPath: jsonPath}
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

var validVarFormat = regexp.MustCompile("^%([-+@0 ])?(\\d+)?\\.?(\\d+)?[tdeEfgsqxX]$")

func renderTextString(vars *vars, varName string, varFormat string) (string, error) {
	varV := vars.table[varName]

	if varFormat != "" && !validVarFormat.MatchString(varFormat) {
		return "", errors.New("invalid format for var")
	}

	switch varV.(type) {
	case myjson.String:
		if varFormat == "" {
			varFormat = "%s"
		}
		return fmt.Sprintf(varFormat, string(varV.(myjson.String))), nil
	case myjson.Number:
		if varFormat == "" {
			return fmt.Sprintf("%v", varV), nil
		} else if varFormat[len(varFormat)-1] == 'd' {
			return fmt.Sprintf(varFormat, int(float64(varV.(myjson.Number)))), nil
		} else {
			return fmt.Sprintf(varFormat, float64(varV.(myjson.Number))), nil
		}
	case myjson.Boolean:
		if varFormat == "" {
			varFormat = "%v"
		}
		return fmt.Sprintf(varFormat, bool(varV.(myjson.Boolean))), nil
	default:
		return "", errors.New("invalid json type for template rendering")
	}
}
