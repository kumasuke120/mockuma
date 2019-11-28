package mckmaps

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/kumasuke120/mockuma/internal/myhttp"
	"github.com/kumasuke120/mockuma/internal/myjson"
)

type loadError struct {
	filename string
	err      error
}

func (e *loadError) Error() string {
	return fmt.Sprintf("cannot load the file '%s': %s", e.filename, e.err)
}

type parserError struct {
	filename string
	jsonPath *myjson.Path
	err      error
}

func (e *parserError) Error() string {
	result := ""
	if e.jsonPath == nil {
		result += "cannot parse json data"
	} else {
		result += fmt.Sprintf("cannot parse the value on json-path '%v'", e.jsonPath)
	}

	if e.filename != "" {
		result += fmt.Sprintf(" in the file '%s'", e.filename)
	}

	if e.err != nil {
		result += ": " + e.err.Error()
	}

	return result
}

type parser struct {
	filename string
}

func (p *parser) parse(chdir bool) (*MockuMappings, error) {
	var json interface{}
	var err error
	if json, err = p.load(ppRemoveComment, ppRenderTemplate); err != nil {
		return nil, err
	}

	var result *MockuMappings
	switch json.(type) {
	case myjson.Object: // parses in multi-file mode
		parser := &mainParser{json: json.(myjson.Object), parser: *p}
		result, err = parser.parse()
	case myjson.Array: // parses in single-file mode
		parser := &mappingsParser{json: json, parser: *p}
		mappings, _err := parser.parse()
		if _err == nil {
			result, err = &MockuMappings{Mappings: mappings}, _err
		} else {
			result, err = nil, _err
		}
	default:
		result, err = nil, newParserError(p.filename, nil)
	}

	p.reset()
	return result, err
}

func (p *parser) load(preprocessors ...filter) (interface{}, error) {
	bytes, err := ioutil.ReadFile(p.filename)
	if err != nil {
		return nil, err
	}

	json, err := myjson.Unmarshal(bytes)
	if err != nil {
		return nil, newParserError(p.filename, nil)
	}

	v, err := doFiltersOnV(json, preprocessors...) // runs given preprocessors
	if err != nil {
		return nil, &loadError{filename: p.filename, err: err}
	}
	return v, nil
}

func (p *parser) reset() {
	ppRenderTemplate.reset()
	ppLoadFile.reset()
	ppParseRegexp.reset()
}

type mainParser struct {
	json myjson.Object
	parser
}

func (p *mainParser) parse() (*MockuMappings, error) {
	_type, err := p.json.GetString(dType)
	if err != nil || string(_type) != tMain {
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

	var mappings []*Mapping
	for idx, filename := range filenamesOfMappings {
		_filename, err := myjson.ToString(filename)
		if err != nil {
			return nil, newParserError(p.filename, myjson.NewPath(dInclude, tMappings, idx))
		}

		parser := &mappingsParser{parser: parser{filename: string(_filename)}}
		partOfMappings, err := parser.parse() // parses mappings for each included file
		if err != nil {
			return nil, err
		}

		mappings = append(mappings, partOfMappings...)
	}

	return &MockuMappings{Mappings: mappings}, nil
}

type mappingsParser struct {
	json     interface{}
	jsonPath *myjson.Path
	parser
}

func (p *mappingsParser) parse() ([]*Mapping, error) {
	if p.json == nil {
		json, err := p.load(ppRemoveComment, ppRenderTemplate)
		if err != nil {
			return nil, err
		}
		p.json = json
	}

	var rawMappings myjson.Array
	switch p.json.(type) {
	case myjson.Object: // parses in multi-file mode
		p.jsonPath = myjson.NewPath("")
		jsonObject := p.json.(myjson.Object)

		p.jsonPath.SetLast(dType)
		_type, err := jsonObject.GetString(dType)
		if err != nil || string(_type) != tMappings {
			return nil, newParserError(p.filename, p.jsonPath)
		}

		p.jsonPath.SetLast(tMappings)
		rawMappings = ensureJsonArray(jsonObject.Get(tMappings))
	case myjson.Array: // parses in single-file mode
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
	mapping.Uri = string(uri)

	p.jsonPath.SetLast(mapMethod)
	if v.Has(mapMethod) {
		method, err := v.GetString(mapMethod)
		if err != nil {
			return nil, newParserError(p.filename, p.jsonPath)
		}
		mapping.Method = myhttp.ToHttpMethod(string(method))
	} else {
		mapping.Method = myhttp.Any
	}

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
	mapping.Policies = policies

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
		policy.When = when
	}

	p.jsonPath.SetLast(mapPolicyReturns)
	rawReturns, err := v.GetObject(mapPolicyReturns)
	if err != nil {
		return nil, newParserError(p.filename, p.jsonPath)
	}
	returns, err := p.parseReturns(rawReturns)
	if err != nil {
		return nil, err
	}
	policy.Returns = returns

	p.jsonPath.RemoveLast()
	return policy, nil
}

func (p *mappingsParser) parseWhen(v myjson.Object) (*When, error) {
	p.jsonPath.Append("")

	_v, err := doFiltersOnV(v, ppParseRegexp, ppLoadFile)
	if err != nil {
		return nil, &loadError{filename: p.filename, err: err}
	}
	switch _v.(type) {
	case myjson.Object:
		v = _v.(myjson.Object)
	default:
		return nil, newParserError(p.filename, p.jsonPath)
	}

	when := new(When)

	p.jsonPath.SetLast(pHeaders)
	if v.Has(pHeaders) {
		rawHeaders, err := v.GetObject(pHeaders)
		if err != nil {
			return nil, newParserError(p.filename, p.jsonPath)
		}

		normalHeaders, regexpHeaders := divideNormalsAndRegexps(rawHeaders)
		when.Headers = parseAsNameValuesPairs(normalHeaders)
		regexpPairs, err := parseAsNameRegexpPairs(regexpHeaders)
		if err != nil {
			newErr := newParserError(p.filename, p.jsonPath)
			newErr.err = err
			return nil, newErr
		}
		when.HeaderRegexps = regexpPairs
	}

	p.jsonPath.SetLast(pParams)
	if v.Has(pParams) {
		rawParams, err := v.GetObject(pParams)
		if err != nil {
			return nil, newParserError(p.filename, p.jsonPath)
		}

		normalParams, regexpParams := divideNormalsAndRegexps(rawParams)
		when.Params = parseAsNameValuesPairs(normalParams)
		regexpPairs, err := parseAsNameRegexpPairs(regexpParams)
		if err != nil {
			newErr := newParserError(p.filename, p.jsonPath)
			newErr.err = err
			return nil, newErr
		}
		when.ParamRegexps = regexpPairs
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
		returns.StatusCode = myhttp.StatusCode(int(statusCode))
	} else {
		returns.StatusCode = myhttp.Ok
	}

	p.jsonPath.SetLast(pHeaders)
	if v.Has(pHeaders) {
		rawHeaders, err := v.GetObject(pHeaders)
		if err != nil {
			return nil, newParserError(p.filename, p.jsonPath)
		}
		returns.Headers = parseAsNameValuesPairs(rawHeaders)
	}

	p.jsonPath.SetLast(pBody)
	rawBody := v.Get(pBody)
	body, err := p.parseBody(rawBody)
	if err != nil {
		return nil, newParserError(p.filename, p.jsonPath)
	}
	returns.Body = body

	p.jsonPath.RemoveLast()
	return returns, nil
}

func (p *mappingsParser) parseBody(v interface{}) ([]byte, error) {
	v, err := doFiltersOnV(v, ppLoadFile)
	if err != nil {
		return nil, &loadError{filename: p.filename, err: err}
	}

	switch v.(type) {
	case nil:
		return nil, nil
	case myjson.String:
		return []byte(string(v.(myjson.String))), nil
	case myjson.Object:
		return p.parseJsonToBytes(v)
	case myjson.Array:
		return p.parseJsonToBytes(v)
	}

	return nil, newParserError(p.filename, p.jsonPath)
}

func (p *mappingsParser) parseJsonToBytes(v interface{}) ([]byte, error) {
	bytes, err := myjson.Marshal(v)
	if err != nil {
		return nil, newParserError(p.filename, p.jsonPath)
	}
	return bytes, nil
}

type templateParser struct {
	json     myjson.Object
	jsonPath *myjson.Path
	parser
}

func (p *templateParser) parse() (*template, error) {
	json, err := p.load(ppRemoveComment)
	if err != nil {
		return nil, err
	}

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

	return template, nil
}

type varsParser struct {
	json     myjson.Object
	jsonPath *myjson.Path
	parser
}

func (p *varsParser) parse() ([]*vars, error) {
	json, err := p.load(ppRemoveComment)
	if err != nil {
		return nil, err
	}

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

func (p *varsParser) parseVars(v myjson.Object) ([]*vars, error) {
	rawVarsArray := ensureJsonArray(v.Get(tVars))
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

func newParserError(filename string, jsonPath *myjson.Path) *parserError {
	return &parserError{filename: filename, jsonPath: jsonPath}
}

func ensureJsonArray(v interface{}) myjson.Array {
	switch v.(type) {
	case myjson.Array:
		return v.(myjson.Array)
	default:
		return myjson.NewArray(v)
	}
}

func ensureSlice(v interface{}) []interface{} {
	switch v.(type) {
	case myjson.Array:
		return v.(myjson.Array)
	default:
		return []interface{}{v}
	}
}

func divideNormalsAndRegexps(v myjson.Object) (myjson.Object, map[string]*regexp.Regexp) {
	normals := make(myjson.Object)
	regexps := make(map[string]*regexp.Regexp)

	for name, rawValue := range v {
		var normV myjson.Array

		for _, rV := range ensureSlice(rawValue) { // divides normals and regexps
			switch rV.(type) {
			case *regexp.Regexp:
				_rV := rV.(*regexp.Regexp)
				if _, ok := regexps[name]; !ok { // only first @regexp is effective
					regexps[name] = _rV
				}
				continue
			}
			normV = append(normV, rV)
		}

		if len(normV) != 0 {
			normals[name] = normV
		}
	}

	return normals, regexps
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

	pair.Name = n

	values := make([]string, len(v))
	for i, p := range v {
		switch p.(type) {
		case nil:
			values[i] = ""
		default:
			str, err := myjson.ToString(p)
			if err != nil {
				panic("Shouldn't happen")
			}
			values[i] = string(str)
		}

	}
	pair.Values = values

	return pair
}

func parseAsNameRegexpPairs(o map[string]*regexp.Regexp) ([]*NameRegexpPair, error) {
	var pairs []*NameRegexpPair
	for name, value := range o {
		pair := new(NameRegexpPair)
		pair.Name = name
		pair.Regexp = value

		pairs = append(pairs, pair)
	}
	return pairs, nil
}

var varNameRegexp = regexp.MustCompile("(?i)[a-z][a-z\\d]*")

func parseVars(v myjson.Object) (*vars, error) {
	vars := new(vars)
	table := make(map[string]interface{})
	for name, value := range v {
		if !varNameRegexp.Match([]byte(name)) {
			return nil, errors.New("invalid name for var")
		}
		table[name] = value
	}
	vars.table = table
	return vars, nil
}
