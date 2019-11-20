package mckmaps

import (
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
	tTemplate = "template"
)

const (
	dType    = "@type"
	dInclude = "@include"
	dFile    = "@file"
)

const (
	mapUri           = "uri"
	mapMethod        = "method"
	mapPolicies      = "policies"
	mapPolicyWhen    = "when"
	mapPolicyReturns = "returns"
)

const (
	pStatusCode = "statusCode"
	pHeaders    = "headers"
	pParams     = "params"
	pBody       = "body"
)

var emptyWhen = new(When)

type parserError struct {
	filename string
	jsonPath *myjson.Path
}

func (e *parserError) Error() string {
	result := ""
	if e.jsonPath == nil {
		result += "cannot parse json data"
	} else {
		result += fmt.Sprintf("cannot parse value on json-path '%v'", e.jsonPath)
	}

	if e.filename != "" {
		result += fmt.Sprintf(" in the file '%s'", e.filename)
	}

	return result
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
	json     interface{}
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
	commentProcessor{v: json}.process()

	switch json.(type) {
	case myjson.Object:
		parser := &mainParser{json: json.(myjson.Object), parser: *p}
		return parser.parse()
	case myjson.Array:
		parser := &mappingsParser{json: json, parser: *p}
		mappings, err := parser.parse()
		if err != nil {
			return nil, err
		}
		return &MockuMappings{mappings: mappings}, nil
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

	return nil
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

		parser := &mappingsParser{parser: parser{filename: string(_filename), chdir: p.chdir}}
		partOfMappings, err := parser.parse()
		if err != nil {
			return nil, err
		}

		mappings = append(mappings, partOfMappings...)
	}

	return &MockuMappings{mappings: mappings}, nil
}

func (p *mappingsParser) parse() ([]*Mapping, error) {
	if p.json == nil {
		json, err := p.load()
		if err != nil {
			return nil, err
		}
		commentProcessor{v: json}.process()
		p.json = json
	}

	var rawMappings myjson.Array
	switch p.json.(type) {
	case myjson.Object:
		p.jsonPath = myjson.NewPath("")
		jsonObject := p.json.(myjson.Object)

		p.jsonPath.SetLast(dType)
		_type, err := jsonObject.GetString(dType)
		if err != nil || _type != tMappings {
			return nil, newParserError(p.filename, p.jsonPath)
		}

		p.jsonPath.SetLast(tMappings)
		rawMappings = ensureJsonArray(jsonObject.Get(tMappings))
	case myjson.Array:
		p.jsonPath = myjson.NewPath()
		rawMappings = p.json.(myjson.Array)
	default:
		p.jsonPath = myjson.NewPath()
		return nil, newParserError(p.filename, p.jsonPath)
	}

	p.jsonPath.Append(0)
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
	p.jsonPath.RemoveLast()

	return mappings, nil
}

func (p *mappingsParser) parseMapping(v myjson.Object) (*Mapping, error) {
	p.jsonPath.Append("")

	mapping := new(Mapping)

	p.jsonPath.SetLast(mapUri)
	uri, err := v.GetString(mapUri)
	if err != nil {
		return nil, newParserError(p.filename, p.jsonPath)
	}
	mapping.uri = string(uri)

	p.jsonPath.SetLast(mapMethod)
	method, err := v.GetString(mapMethod)
	if err != nil {
		return nil, newParserError(p.filename, p.jsonPath)
	}
	mapping.method = myhttp.ToHttpMethod(string(method))

	p.jsonPath.SetLast(mapPolicies)
	p.jsonPath.Append(0)
	var policies []*Policy
	for idx, rp := range ensureJsonArray(v.Get(mapPolicies)) {
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
	p.jsonPath.RemoveLast()
	mapping.policies = policies

	p.jsonPath.RemoveLast()
	return mapping, nil
}

func (p *mappingsParser) parsePolicy(v myjson.Object) (*Policy, error) {
	p.jsonPath.Append("")

	policy := new(Policy)

	p.jsonPath.SetLast(mapPolicyWhen)
	var when *When
	if v.Has(mapPolicyWhen) {
		rawWhen, err := v.GetObject(mapPolicyWhen)
		if err != nil {
			return nil, newParserError(p.filename, p.jsonPath)
		}
		when, err = p.parseWhen(rawWhen)
		if err != nil {
			return nil, err
		}
	} else {
		when = emptyWhen
	}
	policy.when = when

	p.jsonPath.SetLast(mapPolicyReturns)
	rawReturns, err := v.GetObject(mapPolicyReturns)
	if err != nil {
		return nil, newParserError(p.filename, p.jsonPath)
	}
	returns, err := p.parseReturns(rawReturns)
	if err != nil {
		return nil, err
	}
	policy.returns = returns

	p.jsonPath.RemoveLast()
	return policy, nil
}

func (p *mappingsParser) parseWhen(v myjson.Object) (*When, error) {
	p.jsonPath.Append("")

	when := new(When)

	p.jsonPath.SetLast(pHeaders)
	if v.Has(pHeaders) {
		rawHeaders, err := v.GetObject(pHeaders)
		if err != nil {
			return nil, newParserError(p.filename, p.jsonPath)
		}
		when.headers = parseAsNameValuesPairs(rawHeaders)
	}

	p.jsonPath.SetLast(pParams)
	if v.Has(pParams) {
		rawParams, err := v.GetObject(pParams)
		if err != nil {
			return nil, newParserError(p.filename, p.jsonPath)
		}
		when.params = parseAsNameValuesPairs(rawParams)
	}

	p.jsonPath.RemoveLast()
	return when, nil
}

func (p *mappingsParser) parseReturns(v myjson.Object) (*Returns, error) {
	p.jsonPath.Append("")

	returns := new(Returns)

	p.jsonPath.SetLast(pStatusCode)
	if v.Has(pStatusCode) {
		statusCode, err := v.GetNumber(pStatusCode)
		if err != nil {
			return nil, newParserError(p.filename, p.jsonPath)
		}
		returns.statusCode = myhttp.StatusCode(int(statusCode))
	}

	p.jsonPath.SetLast(pHeaders)
	if v.Has(pHeaders) {
		rawHeaders, err := v.GetObject(pHeaders)
		if err != nil {
			return nil, newParserError(p.filename, p.jsonPath)
		}
		returns.headers = parseAsNameValuesPairs(rawHeaders)
	}

	p.jsonPath.SetLast(pBody)
	rawBody := v.Get(pBody)
	body, err := p.parseBody(rawBody)
	if err != nil {
		return nil, newParserError(p.filename, p.jsonPath)
	}
	returns.body = body

	p.jsonPath.RemoveLast()
	return returns, nil
}

func (p *mappingsParser) parseBody(v interface{}) ([]byte, error) {
	switch v.(type) {
	case nil:
		return nil, nil
	case string:
		return []byte(v.(string)), nil
	case myjson.Object:
		if ok, bytes, err := p.parseDirectiveFile(v.(myjson.Object)); ok {
			if err != nil {
				return nil, err
			} else {
				return bytes, nil
			}
		} else {
			bytes, err := myjson.Marshal(v)
			if err != nil {
				return nil, newParserError(p.filename, p.jsonPath)
			}
			return bytes, nil
		}
	case myjson.Array:
		bytes, err := myjson.Marshal(v)
		if err != nil {
			return nil, newParserError(p.filename, p.jsonPath)
		}
		return bytes, nil
	}

	return nil, newParserError(p.filename, p.jsonPath)
}

func (p *mappingsParser) parseDirectiveFile(v myjson.Object) (bool, []byte, error) {
	dFileValue, err := v.GetString(dFile)
	if err == nil {
		bytes, err := ioutil.ReadFile(string(dFileValue))
		if err != nil {
			return true, nil, err
		}
		return true, bytes, nil
	}
	return false, nil, nil
}

func ensureJsonArray(v interface{}) myjson.Array {
	switch v.(type) {
	case myjson.Array:
		return v.(myjson.Array)
	default:
		return myjson.NewArray(v)
	}
}

func parseAsNameValuesPairs(o myjson.Object) []*NameValuesPair {
	var pairs []*NameValuesPair
	for name, rawValues := range o {
		p := parseAsNameValuesPair(name, ensureJsonArray(rawValues))
		pairs = append(pairs, p)
	}
	return pairs
}

func parseAsNameValuesPair(n string, v myjson.Array) *NameValuesPair {
	pair := new(NameValuesPair)

	pair.name = n

	values := make([]string, len(v))
	for i, p := range v {
		values[i] = typeutil.ToString(p)
	}
	pair.values = values

	return pair
}

func newParserError(filename string, jsonPath *myjson.Path) *parserError {
	return &parserError{filename: filename, jsonPath: jsonPath}
}

type templateParser struct {
	json     myjson.Object
	jsonPath *myjson.Path
	parser
}

func (p *templateParser) parse() (*Template, error) {
	json, err := p.load()
	if err != nil {
		return nil, err
	}
	commentProcessor{v: json}.process()

	p.jsonPath = myjson.NewPath()
	switch json.(type) {
	case myjson.Object:
		p.jsonPath.Append("")
		p.json = json.(myjson.Object)
	default:
		return nil, newParserError(p.filename, p.jsonPath)
	}

	p.jsonPath.SetLast(dType)
	_type, err := p.json.GetString(dType)
	if err != nil || _type != tTemplate {
		return nil, newParserError(p.filename, p.jsonPath)
	}

	template := new(Template)

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

	return template, nil
}
