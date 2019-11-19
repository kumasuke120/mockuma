package mckmaps

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/kumasuke120/mockuma/internal/myhttp"
	"github.com/kumasuke120/mockuma/internal/myjson"
	"github.com/kumasuke120/mockuma/internal/typeutil"
)

const (
	tMain     = "main"
	tMappings = "mappings"
)

const (
	dType    = "@type"
	dInclude = "@include"
)

const (
	mapUri           = "uri"
	mapMethod        = "method"
	mapPolicies      = "policies"
	mapPolicyWhen    = "when"
	mapPolicyReturns = "returns"
)

const (
	pParams  = "params"
	pHeaders = "headers"
)

type JsonParseError struct {
	JsonPath string
}

func (e *JsonParseError) Error() string {
	if e.JsonPath == "" {
		return "cannot parse json data"
	} else {
		return fmt.Sprintf("cannot parse value on json-path '%s", e.JsonPath)
	}
}

type ParserError struct {
	Filename string
	JsonPath *myjson.Path
}

func (e *ParserError) Error() string {
	result := ""
	if e.JsonPath == nil {
		result += "cannot parse json data"
	} else {
		result += fmt.Sprintf("cannot parse value on json-path '%v'", e.JsonPath)
	}

	if e.Filename != "" {
		result += fmt.Sprintf(" in the file '%s'", e.Filename)
	}

	return result
}

func newParserError(filename string, jsonPath *myjson.Path) *ParserError {
	return &ParserError{Filename: filename, JsonPath: jsonPath}
}

type parser struct {
	filename string
	chdir    bool
}

type mainParser struct {
	json myjson.Object
	parser
}

type mappingsParser struct {
	json     myjson.Object
	jsonPath *myjson.Path
	parser
}

func (p *parser) load() (interface{}, error) {
	bytes, err := ioutil.ReadFile(p.filename)
	if err != nil {
		return nil, err
	}

	json, err := myjson.Unmarshal(bytes)
	if err != nil {
		return nil, newParserError(p.filename, nil)
	}

	if p.chdir {
		err := p.chdirBasedOnFilename()
		if err != nil {
			return nil, err
		}
	}

	return json, nil
}

func (p *parser) parse() (*MockuMappings, error) {
	json, err := p.load()
	if err != nil {
		return nil, err
	}

	switch json.(type) {
	case myjson.Object:
		parser := &mainParser{parser: *p, json: json.(myjson.Object)}
		return parser.parse()
	}

	return nil, newParserError(p.filename, nil)
}

func (p *parser) chdirBasedOnFilename() error {
	abs, err := filepath.Abs(p.filename)
	if err != nil {
		return err
	}

	dir := filepath.Dir(abs)
	err = os.Chdir(dir)
	if err != nil {
		return err
	}

	log.Println("[load] working directory has been changed to:", dir)
}

func (p *mainParser) parse() (*MockuMappings, error) {
	_type, err := p.json.GetString(dType)
	if err != nil || _type != tMain {
		return nil, newParserError(p.filename, myjson.NewPath(dType))
	}

	include, err := p.json.GetObject(dInclude)
	if err != nil {
		return nil, newParserError(p.filename, myjson.NewPath(dInclude))
	}

	filenamesOfMappings, err := include.GetArray(tMappings)
	if err != nil {
		return nil, newParserError(p.filename, myjson.NewPath(dInclude, tMappings))
	}

	var mappings = make([]*Mapping, len(filenamesOfMappings))
	for idx, filename := range filenamesOfMappings {
		_filename, err := myjson.ToString(filename)
		if err != nil {
			return nil, newParserError(p.filename, myjson.NewPath(dInclude, tMappings, idx))
		}

		parser := &mappingsParser{parser: parser{filename: string(_filename)}}
		partOfMappings, err := parser.parse()
		if err != nil {
			return nil, err
		}

		mappings = append(mappings, partOfMappings...)
	}

	return &MockuMappings{Mappings: mappings}, nil
}

func (p *mappingsParser) parse() ([]*Mapping, error) {
	_type, err := p.json.GetString(dType)
	if err != nil || _type != tMappings {
		return nil, newParserError(p.filename, myjson.NewPath(dType))
	}

	rawMappings, err := ensureJsonArray(p.json.Get(tMappings))
	if err != nil {
		return nil, newParserError(p.filename, myjson.NewPath(tMappings))
	}

	p.jsonPath = myjson.NewPath(tMappings, 0)
	var mappings []*Mapping
	for idx, rm := range rawMappings {
		p.jsonPath.SetLast(idx)

		switch rm.(type) {
		case myjson.Object:
			mapping, err := p.parseMapping(rm.(myjson.Object))
			if err != nil {
				return nil, err
			}
			mappings = append(mappings, mapping)
		default:
			return nil, newParserError(p.filename, p.jsonPath)
		}
	}

	return mappings, nil
}

func (p *mappingsParser) parseMapping(v myjson.Object) (*Mapping, error) {
	mapping := new(Mapping)

	p.jsonPath.Append(mapUri)
	uri, err := v.GetString(mapUri)
	if err != nil {
		return nil, newParserError(p.filename, p.jsonPath)
	}
	mapping.Uri = string(uri)

	p.jsonPath.SetLast(mapMethod)
	method, err := v.GetString(mapMethod)
	if err != nil {
		return nil, newParserError(p.filename, p.jsonPath)
	}
	mapping.Method = myhttp.ToHttpMethod(string(method))

	p.jsonPath.SetLast(mapPolicies)
	rawPolicies, err := ensureJsonArray(p.json.Get(mapPolicies))
	p.jsonPath.Append(0)
	var policies []*Policy
	for idx, rp := range rawPolicies {
		p.jsonPath.SetLast(idx)

		switch rp.(type) {
		case myjson.Object:
			policy, err := p.parsePolicy(rp.(myjson.Object))
			if err != nil {
				return nil, err
			}
			policies = append(policies, policy)
		default:
			return nil, newParserError(p.filename, p.jsonPath)
		}
	}
	mapping.Policies = policies

	return mapping, nil
}

func (p *mappingsParser) parsePolicy(v myjson.Object) (*Policy, error) {
	policy := new(Policy)

	p.jsonPath.Append(mapPolicyWhen)
	if v.Has(mapPolicyWhen) {
		rawWhen, err := v.GetObject(mapPolicyWhen)
		if err != nil {
			return nil, newParserError(p.filename, p.jsonPath)
		}
		policy.When, err = p.parseWhen(rawWhen)
		if err != nil {
			return nil, err
		}
	}

}

func (p *mappingsParser) parseWhen(v myjson.Object) (*When, error) {
	p.jsonPath.Append(pParams)
	if v.Has(pParams) {
		rawParams, err := v.GetObject(pParams)
		if err != nil {
			return nil, newParserError(p.filename, p.jsonPath)
		}
		var params []*NameValuesPair
		for name, rawValues := range rawParams {
			p := parseAsNameValuesPair(name, rawValues)
			params = append(params, p)
		}
	}
}

func ensureJsonArray(v interface{}) (myjson.Array, error) {
	switch v.(type) {
	case myjson.Object:
		return myjson.NewArray(v), nil
	case myjson.Array:
		return v.(myjson.Array), nil
	default:
		return nil, errors.New("cannot convert to myjson.Array")
	}
}

func parseAsNameValuesPair(n string, v myjson.Array) *NameValuesPair {
	pair := new(NameValuesPair)

	pair.Name = n

	values := make([]string, len(v))
	for i, p := range v {
		values[i] = typeutil.ToString(p)
	}
	pair.Values = values

	return pair
}
