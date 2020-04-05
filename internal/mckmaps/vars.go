package mckmaps

import (
	"encoding/csv"
	"errors"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/kumasuke120/mockuma/internal/myjson"
)

var varNameRegexp = regexp.MustCompile("(?i)[a-z][a-z\\d]*")

type vars struct {
	table map[string]interface{}
}

type varsJSONParser struct {
	json     myjson.Object
	jsonPath *myjson.Path
	Parser
}

func (p *varsJSONParser) parse() ([]*vars, error) {
	if p.json == nil {
		json, err := p.load(true, ppRemoveComment)
		if err != nil {
			return nil, err
		}

		p.jsonPath = myjson.NewPath()
		switch json.(type) {
		case myjson.Object:
			p.json = json.(myjson.Object)
		default:
			return nil, newParserError(p.filename, p.jsonPath)
		}
	}

	p.jsonPath = myjson.NewPath("")
	p.jsonPath.SetLast(aType)
	_type, err := p.json.GetString(aType)
	if err != nil || string(_type) != tVars {
		return nil, newParserError(p.filename, p.jsonPath)
	}

	p.jsonPath.SetLast(tVars)
	p.jsonPath.Append(0)
	varsSlice, err := p.parseVars(p.json)
	if err != nil {
		return nil, err
	}
	p.jsonPath.RemoveLast()

	return varsSlice, nil
}

func (p *varsJSONParser) parseVars(v myjson.Object) ([]*vars, error) {
	rawVarsArray := ensureJSONArray(v.Get(tVars))
	varsSlice := make([]*vars, len(rawVarsArray))
	for idx, rawVars := range rawVarsArray {
		if p.json != nil {
			p.jsonPath.SetLast(idx)
		}
		rVars, err := myjson.ToObject(rawVars)
		if err != nil {
			return nil, newParserError(p.filename, p.jsonPath)
		}
		varsSlice[idx], err = parseVars(rVars)
		if err != nil {
			return nil, newParserError(p.filename, p.jsonPath)
		}
	}
	return varsSlice, nil
}

func parseVars(v myjson.Object) (*vars, error) {
	vars := new(vars)
	table := make(map[string]interface{})
	for name, value := range v {
		if !varNameRegexp.MatchString(name) {
			return nil, errors.New("invalid name for var")
		}
		table[name] = value
	}
	vars.table = table
	return vars, nil
}

type varsCSVParser struct {
	rdr *csv.Reader
	Parser
}

func (p *varsCSVParser) parse() ([]*vars, error) {
	if p.rdr == nil {
		file, err := os.Open(p.filename)
		if err != nil {
			return nil, &loadError{filename: p.filename, err: err}
		}
		defer func() {
			_ = file.Close()
		}()
		p.rdr = csv.NewReader(file)
	}

	var result []*vars
	var varNames []string
	for {
		line, err := p.rdr.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, &loadError{filename: p.filename, err: err}
		}

		if varNames == nil {
			varNames = line
			if len(varNames) != 0 {
				varNames[0] = cleanBom(varNames[0])
			}
			continue
		}

		table := make(map[string]interface{}, len(varNames))
		for i, c := range line {
			if i < len(varNames) {
				var col interface{}
				json, err := myjson.Unmarshal([]byte(c))
				if err != nil {
					col = myjson.String(c) // treats the non-valid json as a pure string
				} else {
					col = json
				}
				table[varNames[i]] = col
			}
		}

		result = append(result, &vars{table: table})
	}
	return result, nil
}

const bom = "\xef\xbb\xbf"

func cleanBom(s string) string {
	if strings.HasPrefix(s, bom) {
		return s[3:]
	}
	return s
}
