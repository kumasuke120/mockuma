package mckmaps

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/kumasuke120/mockuma/internal/myhttp"
	"github.com/kumasuke120/mockuma/internal/myjson"
	"github.com/kumasuke120/mockuma/internal/types"
)

type Mapping struct {
	URI      string
	Method   myhttp.HTTPMethod
	Policies []*Policy
}

type Policy struct {
	When     *When
	CmdType  CmdType
	Returns  *Returns
	Forwards *Forwards
}

type When struct {
	Headers       []*NameValuesPair
	HeaderRegexps []*NameRegexpPair
	HeaderJSONs   []*NameJSONPair

	Params       []*NameValuesPair
	ParamRegexps []*NameRegexpPair
	ParamJSONs   []*NameJSONPair

	PathVars       []*NameValuesPair
	PathVarRegexps []*NameRegexpPair

	Body       []byte
	BodyRegexp *regexp.Regexp
	BodyJSON   *myjson.ExtJSONMatcher
}

type CmdType string

const (
	CmdTypeReturns   = CmdType(mapPolicyReturns)
	CmdTypeForwards  = CmdType(mapPolicyForwards)
	CmdTypeRedirects = CmdType(mapPolicyRedirects)
)

type Returns struct {
	StatusCode myhttp.StatusCode
	Headers    []*NameValuesPair
	Body       []byte
	Latency    *Interval
}

type Forwards struct {
	Path    string
	Latency *Interval
}

type NameValuesPair struct {
	Name   string
	Values []string
}

type NameRegexpPair struct {
	Name   string
	Regexp *regexp.Regexp
}

type NameJSONPair struct {
	Name string
	JSON myjson.ExtJSONMatcher
}

type Interval struct {
	Min int64
	Max int64
}

var (
	// refers to: https://tools.ietf.org/html/rfc7230#section-3.2.6
	methodRegexp      = regexp.MustCompile("(?i)^[-!#$%&'*+._`|~\\da-z]+$")
	forwardPathRegexp = regexp.MustCompile("^(?:https?://)?.+$")
	pathVarRegexp     = regexp.MustCompile("{[^}]*}")
)

type mappingsParser struct {
	json     interface{}
	jsonPath *myjson.Path
	Parser
}

func (p *mappingsParser) parse() ([]*Mapping, error) {
	if p.json == nil {
		// file has been recorded in mainParser
		json, err := p.load(false, ppRemoveComment, ppRenderTemplate)
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

		p.jsonPath.SetLast(aType)
		_type, err := jsonObject.GetString(aType)
		if err != nil || string(_type) != tMappings {
			return nil, newParserError(p.filename, p.jsonPath)
		}

		p.jsonPath.SetLast(tMappings)
		rawMappings = ensureJSONArray(jsonObject.Get(tMappings))
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

	p.jsonPath.SetLast(aMapURI)
	uri, err := v.GetString(aMapURI)
	if err != nil {
		return nil, newParserError(p.filename, p.jsonPath)
	}
	_uri, err := encodeURI(string(uri))
	if err != nil {
		return nil, &parserError{filename: p.filename, jsonPath: p.jsonPath, err: err}
	}
	mapping.URI = _uri

	p.jsonPath.SetLast(aMapMethod)
	if v.Has(aMapMethod) {
		method, err := v.GetString(aMapMethod)
		if err != nil {
			return nil, newParserError(p.filename, p.jsonPath)
		}
		_method := string(method)
		if methodRegexp.MatchString(_method) {
			mapping.Method = myhttp.ToHTTPMethod(_method)
		} else {
			return nil, newParserError(p.filename, p.jsonPath)
		}
	} else {
		mapping.Method = myhttp.MethodAny
	}

	p.jsonPath.SetLast(aMapPolicies)
	p.jsonPath.Append(0)
	var policies []*Policy
	for idx, rp := range ensureJSONArray(v.Get(aMapPolicies)) {
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

	p.renamePathVars(mapping)

	p.jsonPath.RemoveLast()
	return mapping, nil
}

func encodeURI(uri string) (string, error) {
	if strings.HasPrefix(uri, "/") {
		indices := pathVarRegexp.FindAllStringIndex(uri, -1)
		if indices == nil {
			return doEncodeURI(uri), nil
		} else {
			var builder strings.Builder
			for i, loc := range indices {
				var startPos int
				if i == 0 {
					startPos = 0
				} else {
					startPos = indices[i-1][1]
				}
				builder.WriteString(doEncodeURI(uri[startPos:loc[0]]))

				builder.WriteString(uri[loc[0]:loc[1]])

				if i == len(indices)-1 {
					builder.WriteString(doEncodeURI(uri[loc[1]:]))
				}
			}
			return builder.String(), nil
		}
	} else {
		return "", errors.New("uri must start with '/'")
	}
}

func doEncodeURI(uri string) string {
	encoded := &url.URL{Path: uri}
	return encoded.String()
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

	cntCommands := p.countCommands(v, mapPolicyCommands...)
	if cntCommands == 0 {
		policy.Returns = &Returns{
			StatusCode: myhttp.StatusOk,
		}
		policy.CmdType = CmdTypeReturns
	} else if cntCommands == 1 {
		err := p.parseCommand(policy, v)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, &parserError{
			filename: p.filename,
			jsonPath: p.jsonPath,
			err: errors.New(fmt.Sprintf("commands%v can only be used one at a time",
				mapPolicyCommands)),
		}
	}

	p.jsonPath.RemoveLast()
	return policy, nil
}

func (p *mappingsParser) parseCommand(dst *Policy, v myjson.Object) error {
	if v.Has(mapPolicyReturns) {
		p.jsonPath.SetLast(mapPolicyReturns)
		rawReturns, err := v.GetObject(mapPolicyReturns)
		if err != nil {
			return newParserError(p.filename, p.jsonPath)
		}
		returns, err := p.parseReturns(rawReturns)
		if err != nil {
			return err
		}
		dst.Returns = returns
		dst.CmdType = CmdTypeReturns
	} else if v.Has(mapPolicyRedirects) {
		p.jsonPath.SetLast(mapPolicyRedirects)
		rawRedirects, err := v.GetObject(mapPolicyRedirects)
		if err != nil {
			return newParserError(p.filename, p.jsonPath)
		}
		redirects, err := p.parseRedirects(rawRedirects)
		if err != nil {
			return err
		}
		dst.Returns = redirects
		dst.CmdType = CmdTypeRedirects
	} else {
		p.jsonPath.SetLast(mapPolicyForwards)
		rawForwards, err := v.GetObject(mapPolicyForwards)
		if err != nil {
			return newParserError(p.filename, p.jsonPath)
		}
		forwards, err := p.parseForwards(rawForwards)
		if err != nil {
			return err
		}
		dst.Forwards = forwards
		dst.CmdType = CmdTypeForwards
	}

	return nil
}

func (p *mappingsParser) countCommands(v myjson.Object, names ...string) int {
	cnt := 0
	for _, name := range names {
		if v.Has(name) {
			cnt++
		}
	}
	return cnt
}

func (p *mappingsParser) parseWhen(v myjson.Object) (*When, error) {
	p.jsonPath.Append("")

	_v, err := types.DoFiltersOnV(v, ppToJSONMatcher, ppParseRegexp, ppLoadFile)
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
		when.HeaderJSONs = parseAsNameJSONPairs(jsonMHeaders)
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
		when.ParamJSONs = parseAsNameJSONPairs(jsonMHeaders)
	}

	p.jsonPath.SetLast(pPathVars)
	if v.Has(pPathVars) {
		rawPathVars, err := v.GetObject(pPathVars)
		if err != nil {
			return nil, newParserError(p.filename, p.jsonPath)
		}

		normalPathVars, regexpPathVars, jsonMPathVars := divideIntoWhenMatchers(rawPathVars)
		when.PathVars = parseAsNameValuesPairs(normalPathVars)
		when.PathVarRegexps = parseAsNameRegexpPairs(regexpPathVars)
		if len(jsonMPathVars) != 0 {
			return nil, newParserError(p.filename, p.jsonPath)
		}
	}

	p.jsonPath.SetLast(pBody)
	if v.Has(pBody) {
		rawBody := v.Get(pBody)
		bytes, bodyRegexp, jMatcher := p.parseWhenBody(rawBody)
		when.Body = bytes
		when.BodyRegexp = bodyRegexp
		when.BodyJSON = jMatcher
	}

	p.jsonPath.RemoveLast()
	return when, nil
}

func (p *mappingsParser) parseWhenBody(v interface{}) ([]byte, myjson.ExtRegexp, *myjson.ExtJSONMatcher) {
	switch v.(type) {
	case myjson.String:
		return []byte(v.(myjson.String)), nil, nil
	case myjson.ExtRegexp:
		return nil, v.(myjson.ExtRegexp), nil
	case myjson.ExtJSONMatcher:
		_v := v.(myjson.ExtJSONMatcher)
		return nil, nil, &_v
	default:
		return []byte(types.ToString(v)), nil, nil
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
		returns.StatusCode = myhttp.StatusOk
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

func (p *mappingsParser) parseJSONToBytes(v interface{}) ([]byte, error) {
	bytes, err := myjson.Marshal(v)
	if err != nil {
		return nil, newParserError(p.filename, p.jsonPath)
	}
	return bytes, nil
}

func (p *mappingsParser) parseReturnsBody(v interface{}) ([]byte, error) {
	v, err := types.DoFiltersOnV(v, ppLoadFile)
	if err != nil {
		return nil, &loadError{filename: p.filename, err: err}
	}

	switch v.(type) {
	case nil:
		return nil, nil
	case myjson.String:
		return []byte(string(v.(myjson.String))), nil
	case myjson.Object:
		return p.parseJSONToBytes(v)
	case myjson.Array:
		return p.parseJSONToBytes(v)
	}

	return nil, newParserError(p.filename, p.jsonPath)
}

func (p *mappingsParser) parseForwards(v myjson.Object) (*Forwards, error) {
	p.jsonPath.Append("")

	forwards := new(Forwards)

	p.jsonPath.SetLast(pPath)
	path, err := v.GetString(pPath)
	if err != nil || !forwardPathRegexp.MatchString(string(path)) {
		return nil, newParserError(p.filename, p.jsonPath)
	}
	forwards.Path = string(path)

	p.jsonPath.SetLast(pLatency)
	if v.Has(pLatency) {
		rawLatency := v.Get(pLatency)
		latency, err := p.parseLatency(rawLatency)
		if err != nil {
			return nil, newParserError(p.filename, p.jsonPath)
		}
		forwards.Latency = latency
	}

	p.jsonPath.RemoveLast()
	return forwards, nil
}

func (p *mappingsParser) parseRedirects(v myjson.Object) (*Returns, error) {
	p.jsonPath.Append("")

	returns := &Returns{StatusCode: myhttp.StatusFound}

	p.jsonPath.SetLast(pPath)
	path, err := v.GetString(pPath)
	if err != nil || len(string(path)) == 0 {
		return nil, newParserError(p.filename, p.jsonPath)
	}
	returns.Headers = []*NameValuesPair{
		{
			Name:   myhttp.HeaderLocation,
			Values: []string{string(path)},
		},
	}

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

func (p *mappingsParser) renamePathVars(mapping *Mapping) {
	newURI, var2Idx := numberPathVars(mapping.URI)
	mapping.URI = newURI

	for _, pol := range mapping.Policies {
		when := pol.When
		if when != nil {
			l := len(when.PathVars)
			if l != 0 {
				newPVars := p.numberForPathVars(l, when, var2Idx)
				p.sortPathVars(newPVars)
				when.PathVars = newPVars
			}

			l = len(when.PathVarRegexps)
			if l != 0 {
				newPVarRegexps := p.numberForPathVarRegexps(when, var2Idx)
				p.sortPathVarRegexps(newPVarRegexps)
				when.PathVarRegexps = newPVarRegexps
			}
		}
	}
}

func (p *mappingsParser) numberForPathVars(l int, when *When, var2Idx map[string]int) []*NameValuesPair {
	newPVars := make([]*NameValuesPair, l)
	for i, v := range when.PathVars {
		if idx, ok := var2Idx[v.Name]; ok {
			newPVars[i] = &NameValuesPair{
				Name:   strconv.Itoa(idx),
				Values: v.Values,
			}
		} else {
			newPVars[i] = v
		}
	}
	return newPVars
}

func (p *mappingsParser) numberForPathVarRegexps(when *When, var2Idx map[string]int) []*NameRegexpPair {
	newPVarRegexps := make([]*NameRegexpPair, len(when.PathVarRegexps))
	for i, v := range when.PathVarRegexps {
		if idx, ok := var2Idx[v.Name]; ok {
			newPVarRegexps[i] = &NameRegexpPair{
				Name:   strconv.Itoa(idx),
				Regexp: v.Regexp,
			}
		} else {
			newPVarRegexps[i] = v
		}
	}
	return newPVarRegexps
}

func (p *mappingsParser) sortPathVars(newPVars []*NameValuesPair) {
	sort.Slice(newPVars, func(i, j int) bool {
		var err error
		var a, b int
		a, err = strconv.Atoi(newPVars[i].Name)
		if err != nil {
			a = 0
		}
		b, err = strconv.Atoi(newPVars[j].Name)
		if err != nil {
			b = 0
		}
		return a < b
	})
}

func (p *mappingsParser) sortPathVarRegexps(newPVars []*NameRegexpPair) {
	sort.Slice(newPVars, func(i, j int) bool {
		var err error
		var a, b int
		a, err = strconv.Atoi(newPVars[i].Name)
		if err != nil {
			a = 0
		}
		b, err = strconv.Atoi(newPVars[j].Name)
		if err != nil {
			b = 0
		}
		return a < b
	})
}

func numberPathVars(uri string) (n string, m map[string]int) {
	m = make(map[string]int)
	n = pathVarRegexp.ReplaceAllStringFunc(uri, (func() func(string) string {
		idx := 0
		return func(s string) string {
			var i int
			var ok bool
			name := s[1 : len(s)-1]
			if i, ok = m[name]; !ok {
				i = idx
				idx++
				m[name] = i
			}
			return fmt.Sprintf("{%d}", i)
		}
	})())
	return
}

func divideIntoWhenMatchers(v myjson.Object) (myjson.Object,
	map[string]myjson.ExtRegexp, map[string]myjson.ExtJSONMatcher) {

	direct := make(myjson.Object)
	regexps := make(map[string]myjson.ExtRegexp)
	jsonMatchers := make(map[string]myjson.ExtJSONMatcher)

	for name, rawValue := range v {
		var normV myjson.Array

		for _, rV := range ensureJSONArray(rawValue) { // divides normals and regexps
			switch rV.(type) {
			case myjson.ExtRegexp:
				_rV := rV.(myjson.ExtRegexp)
				if _, ok := regexps[name]; !ok { // only first @regexp is effective
					regexps[name] = _rV
				}
				continue
			case myjson.ExtJSONMatcher:
				_rV := rV.(myjson.ExtJSONMatcher)
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
		p := parseAsNameValuesPair(name, ensureJSONArray(rawValues))
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

func parseAsNameJSONPairs(o map[string]myjson.ExtJSONMatcher) []*NameJSONPair {
	var pairs []*NameJSONPair
	for name, value := range o {
		pair := new(NameJSONPair)
		pair.Name = name
		pair.JSON = value

		pairs = append(pairs, pair)
	}
	return pairs
}
