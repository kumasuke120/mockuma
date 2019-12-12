package mckmaps

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/kumasuke120/mockuma/internal/myhttp"
	"github.com/kumasuke120/mockuma/internal/myjson"
	"github.com/kumasuke120/mockuma/internal/typeutil"
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
		result += fmt.Sprintf("cannot parse the value on json-path \"%v\"", e.jsonPath)
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

func (p *parser) parse() (*MockuMappings, error) {
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
	return p.sortMappings(result), err
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

func (p *parser) sortMappings(mappings *MockuMappings) *MockuMappings {
	if mappings == nil {
		return nil
	}

	uri2mappings := make(map[string][]*Mapping)
	var uriOrder []string
	for _, m := range mappings.Mappings {
		mappingsOfUri := uri2mappings[m.Uri]
		mappingsOfUri = appendToMappingsOfUri(mappingsOfUri, m)
		uri2mappings[m.Uri] = mappingsOfUri
		uriOrder = append(uriOrder, m.Uri)
	}

	ms := make([]*Mapping, 0, len(mappings.Mappings))
	for _, uri := range uriOrder {
		mappingsOfUri := uri2mappings[uri]
		ms = append(ms, mappingsOfUri...)
	}
	return &MockuMappings{Mappings: ms}
}

func appendToMappingsOfUri(dst []*Mapping, m *Mapping) []*Mapping {
	appended := false
	for _, dm := range dst {
		if dm.Uri == m.Uri && dm.Method == m.Method {
			dm.Policies = append(dm.Policies, m.Policies...)
			appended = true
		}
	}

	if !appended {
		dst = append(dst, m)
	}
	return dst
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

	_v, err := doFiltersOnV(v, ppToJsonMatcher, ppParseRegexp, ppLoadFile)
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

		normalHeaders, regexpHeaders, jsonMHeaders := divideIntoWhenMatchers(rawHeaders)
		when.Headers = parseAsNameValuesPairs(normalHeaders)
		when.HeaderRegexps = parseAsNameRegexpPairs(regexpHeaders)
		when.HeaderJsons = parseAsNameJsonPairs(jsonMHeaders)
	}

	p.jsonPath.SetLast(pParams)
	if v.Has(pParams) {
		rawParams, err := v.GetObject(pParams)
		if err != nil {
			return nil, newParserError(p.filename, p.jsonPath)
		}

		normalParams, regexpParams, jsonMHeaders := divideIntoWhenMatchers(rawParams)
		when.Params = parseAsNameValuesPairs(normalParams)
		when.ParamRegexps = parseAsNameRegexpPairs(regexpParams)
		when.ParamJsons = parseAsNameJsonPairs(jsonMHeaders)
	}

	p.jsonPath.SetLast(pBody)
	if v.Has(pBody) {
		rawBody := v.Get(pBody)
		bytes, bodyRegexp, jMatcher := p.parseWhenBody(rawBody)
		when.Body = bytes
		when.BodyRegexp = bodyRegexp
		when.BodyJson = jMatcher
	}

	p.jsonPath.RemoveLast()
	return when, nil
}

func (p *mappingsParser) parseWhenBody(v interface{}) ([]byte, myjson.ExtRegexp, *myjson.ExtJsonMatcher) {
	switch v.(type) {
	case myjson.String:
		return []byte(v.(myjson.String)), nil, nil
	case myjson.ExtRegexp:
		return nil, v.(myjson.ExtRegexp), nil
	case myjson.ExtJsonMatcher:
		_v := v.(myjson.ExtJsonMatcher)
		return nil, nil, &_v
	default:
		return []byte(typeutil.ToString(v.(myjson.Number))), nil, nil
	}
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
	body, err := p.parseReturnsBody(rawBody)
	if err != nil {
		return nil, newParserError(p.filename, p.jsonPath)
	}
	returns.Body = body

	p.jsonPath.SetLast(pLatency)
	if v.Has(pLatency) {
		rawLatency := v.Get(pLatency)
		latency, err := p.parseLatency(rawLatency)
		if err != nil {
			return nil, newParserError(p.filename, p.jsonPath)
		}
		returns.Latency = latency
	}

	p.jsonPath.RemoveLast()
	return returns, nil
}

func (p *mappingsParser) parseReturnsBody(v interface{}) ([]byte, error) {
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

func (p *mappingsParser) parseLatency(v interface{}) (*Interval, error) {
	switch v.(type) {
	case myjson.Number:
		_v := int64(v.(myjson.Number))
		return &Interval{
			Min: _v,
			Max: _v,
		}, nil
	case myjson.Array:
		va := v.(myjson.Array)
		if len(va) == 1 {
			va0 := va[0]
			switch va0.(type) {
			case myjson.Number:
				return p.parseLatency(va0)
			}
		} else if len(va) == 2 {
			if myjson.IsAllNumber(va) {
				va0 := int64(va[0].(myjson.Number))
				va1 := int64(va[1].(myjson.Number))
				if va1 >= va0 {
					return &Interval{
						Min: va0,
						Max: va1,
					}, nil
				}
			}
		}
	}

	return nil, newParserError(p.filename, p.jsonPath)
}

type templateParser struct {
	json     myjson.Object
	jsonPath *myjson.Path
	parser
}

func (p *templateParser) parse() (*template, error) {
	if p.json == nil {
		json, err := p.load(ppRemoveComment)
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

	return template, nil
}

type varsParser struct {
	json     myjson.Object
	jsonPath *myjson.Path
	parser
}

func (p *varsParser) parse() ([]*vars, error) {
	if p.json == nil {
		json, err := p.load(ppRemoveComment)
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

func divideIntoWhenMatchers(v myjson.Object) (myjson.Object,
	map[string]myjson.ExtRegexp, map[string]myjson.ExtJsonMatcher) {

	direct := make(myjson.Object)
	regexps := make(map[string]myjson.ExtRegexp)
	jsonMatchers := make(map[string]myjson.ExtJsonMatcher)

	for name, rawValue := range v {
		var normV myjson.Array

		for _, rV := range ensureJsonArray(rawValue) { // divides normals and regexps
			switch rV.(type) {
			case myjson.ExtRegexp:
				_rV := rV.(myjson.ExtRegexp)
				if _, ok := regexps[name]; !ok { // only first @regexp is effective
					regexps[name] = _rV
				}
				continue
			case myjson.ExtJsonMatcher:
				_rV := rV.(myjson.ExtJsonMatcher)
				if _, ok := jsonMatchers[name]; !ok { // only first @json is effective
					jsonMatchers[name] = _rV
				}
				continue
			}
			normV = append(normV, rV)
		}

		if len(normV) != 0 {
			direct[name] = normV
		}
	}

	return direct, regexps, jsonMatchers
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

func parseAsNameRegexpPairs(o map[string]myjson.ExtRegexp) []*NameRegexpPair {
	var pairs []*NameRegexpPair
	for name, value := range o {
		pair := new(NameRegexpPair)
		pair.Name = name
		pair.Regexp = value

		pairs = append(pairs, pair)
	}
	return pairs
}

func parseAsNameJsonPairs(o map[string]myjson.ExtJsonMatcher) []*NameJsonPair {
	var pairs []*NameJsonPair
	for name, value := range o {
		pair := new(NameJsonPair)
		pair.Name = name
		pair.Json = value

		pairs = append(pairs, pair)
	}
	return pairs
}

var varNameRegexp = regexp.MustCompile("(?i)[a-z][a-z\\d]*")

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
