package server

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/kumasuke120/mockuma/internal/mckmaps"
	"github.com/kumasuke120/mockuma/internal/myhttp"
	"github.com/kumasuke120/mockuma/internal/myjson"
)

type pathMatcher struct {
	mappings    *mckmaps.MockuMappings
	directPath  map[string][]*mckmaps.Mapping
	patternPath map[*regexp.Regexp][]*mckmaps.Mapping
}

var pathVarRegexp = regexp.MustCompile(`{(\d+)}`)

func newPathMatcher(mappings *mckmaps.MockuMappings) *pathMatcher {
	directPath := make(map[string][]*mckmaps.Mapping)
	patternPath := make(map[*regexp.Regexp][]*mckmaps.Mapping)
	for _, m := range mappings.Mappings {
		if theURI := pathVarRegexp.ReplaceAllString(m.URI, "(?P<v$1>.*?)"); theURI == m.URI {
			mappingsOfURI := directPath[theURI]
			mappingsOfURI = append(mappingsOfURI, m)
			directPath[theURI] = mappingsOfURI
		} else {
			regexpURI := regexp.MustCompile("^" + theURI + "$")
			mappingsOfURI := patternPath[regexpURI]
			mappingsOfURI = append(mappingsOfURI, m)
			patternPath[regexpURI] = mappingsOfURI
		}
	}

	return &pathMatcher{
		mappings:    mappings,
		directPath:  directPath,
		patternPath: patternPath,
	}
}

func (m *pathMatcher) bind(r *http.Request) *boundMatcher {
	return &boundMatcher{m: m, r: r, conf: m.mappings.Config}
}

type boundMatcher struct {
	m          *pathMatcher
	conf       *mckmaps.Config
	r          *http.Request
	method     myhttp.HTTPMethod
	uri        string
	uriPattern *regexp.Regexp

	matchedMapping *mckmaps.Mapping
	matchState     matchState
	bodyCache      []byte
}

type matchState int

const (
	matchExact = iota
	matchURI
	matchHead
	matchCORSOptions
	matchNone
)

func (bm *boundMatcher) matches() bool {
	bm.uri = bm.r.URL.Path
	bm.method = myhttp.ToHTTPMethod(bm.r.Method)

	var possibleMappings []*mckmaps.Mapping
	possibleMappings = bm.matchURIDirect()

	var possibleURIPattern *regexp.Regexp
	if len(possibleMappings) == 0 {
		possibleMappings, possibleURIPattern = bm.matchURIPattern()
	}

	if len(possibleMappings) != 0 { // if finds any mapping
		if matched := bm.matchByMethod(possibleMappings); matched != nil {
			bm.matchedMapping = matched
			bm.uriPattern = possibleURIPattern
			bm.matchState = matchExact
		} else if matched := bm.matchHead(possibleMappings); matched != nil {
			bm.matchedMapping = matched
			bm.matchState = matchHead
		} else if bm.matchCORSOptions() {
			bm.matchedMapping = nil
			bm.matchState = matchCORSOptions
		} else {
			bm.matchedMapping = nil
			bm.matchState = matchURI
		}
	} else {
		bm.matchedMapping = nil
		bm.matchState = matchNone
	}

	return bm.matchState != matchNone
}

func (bm *boundMatcher) matchURIDirect() (pm []*mckmaps.Mapping) {
	if mappingsOfURI, ok := bm.m.directPath[bm.uri]; ok { // matching for direct path
		pm = mappingsOfURI
	} else if bm.conf.MatchTrailingSlash { // matches /path to /path/
		if mappingsOfURI, ok := bm.m.directPath[bm.uri+"/"]; ok {
			bm.uri += "/"
			pm = mappingsOfURI
		}
	}
	return
}

func (bm *boundMatcher) matchURIPattern() (pm []*mckmaps.Mapping, pp *regexp.Regexp) {
	for pattern, mappingsOfURI := range bm.m.patternPath { // matching for pattern path
		if pattern.MatchString(bm.uri) {
			pm = mappingsOfURI
			pp = pattern
		} else if bm.conf.MatchTrailingSlash && pattern.MatchString(bm.uri+"/") {
			bm.uri += "/"
			pm = mappingsOfURI
			pp = pattern
		}
	}
	return
}

func (bm *boundMatcher) headMatches() bool {
	return bm.matchState == matchHead
}

func (bm *boundMatcher) matchByMethod(mappings []*mckmaps.Mapping) *mckmaps.Mapping {
	return matchByMethod(mappings, bm.method)
}

func (bm *boundMatcher) matchHead(mappings []*mckmaps.Mapping) *mckmaps.Mapping {
	if bm.method != myhttp.MethodHead {
		return nil
	}

	return matchByMethod(mappings, myhttp.MethodGet)
}

func (bm *boundMatcher) matchCORSOptions() bool {
	return bm.conf.CORS.Enabled && bm.method == myhttp.MethodOptions
}

func matchByMethod(mappings []*mckmaps.Mapping, method myhttp.HTTPMethod) *mckmaps.Mapping {
	for _, m := range mappings {
		if m.Method.Matches(method) {
			return m
		}
	}
	return nil
}

func (bm *boundMatcher) matchPolicy() *mckmaps.Policy {
	switch bm.matchState {
	case matchHead:
		fallthrough
	case matchExact:
		p := bm.matchExactPolicy()
		if p == nil {
			return pNoPolicyMatched
		} else {
			return p
		}
	case matchCORSOptions:
		return pEmptyOK
	case matchURI:
		return pMethodNotAllowed
	}
	return pNotFound
}

func (bm *boundMatcher) matchExactPolicy() *mckmaps.Policy {
	bm.cacheBody()

	err := bm.r.ParseForm()
	if err != nil {
		log.Println("[server  ] fail to parse form:", err)
		return nil
	}

	var policy *mckmaps.Policy
	for _, p := range bm.matchedMapping.Policies {
		when := p.When

		if when != nil {
			if bm.uriPattern != nil && !bm.pathVarsMatch(when) {
				continue
			}

			if !bm.paramsMatch(when) {
				continue
			}

			if !bm.headersMatch(when) {
				continue
			}

			if !bm.bodyMatches(when) {
				continue
			}
		}

		policy = p
		break
	}

	// resets body for later use in executor
	bm.resetBodyFromCache()

	return policy
}

func (bm *boundMatcher) cacheBody() {
	body, err := ioutil.ReadAll(bm.r.Body)
	if err == nil {
		bm.bodyCache = body
		bm.resetBodyFromCache()
	}
}

func (bm *boundMatcher) resetBodyFromCache() {
	if bm.bodyCache != nil {
		bm.r.Body = ioutil.NopCloser(bytes.NewReader(bm.bodyCache))
	}
}

func (bm *boundMatcher) pathVarsMatch(when *mckmaps.When) bool {
	pathVars := bm.extractPathVars()

	if !valuesMatch(when.PathVars, pathVars) {
		return false
	}
	if !regexpsMatch(when.PathVarRegexps, pathVars) {
		return false
	}

	return true
}

func (bm *boundMatcher) extractPathVars() map[string][]string {
	mValues := bm.uriPattern.FindStringSubmatch(bm.uri)
	if len(mValues) == 0 {
		panic("Shouldn't happen")
	}
	mNames := bm.uriPattern.SubexpNames()
	pathVars := make(map[string][]string, len(mValues))
	for i := 1; i < len(mValues); i++ {
		pathVars[mNames[i][1:]] = []string{mValues[i]} // [1:] to remove the prefix v
	}
	return pathVars
}

func (bm *boundMatcher) paramsMatch(when *mckmaps.When) bool {
	if !valuesMatch(when.Params, bm.r.Form) {
		return false
	}
	if !regexpsMatch(when.ParamRegexps, bm.r.Form) {
		return false
	}
	if !asJSONsMatch(when.ParamJSONs, bm.r.Form) {
		return false
	}

	return true
}

func (bm *boundMatcher) headersMatch(when *mckmaps.When) bool {
	if !valuesMatch(when.Headers, bm.r.Header) {
		return false
	}
	if !regexpsMatch(when.HeaderRegexps, bm.r.Header) {
		return false
	}
	if !asJSONsMatch(when.HeaderJSONs, bm.r.Header) {
		return false
	}

	return true
}

func (bm *boundMatcher) bodyMatches(when *mckmaps.When) bool {
	body := bm.bodyCache
	if when.Body != nil {
		return bytes.Equal(when.Body, body)
	} else if when.BodyRegexp != nil {
		return when.BodyRegexp.Match(body)
	} else if when.BodyJSON != nil {
		json, err := myjson.Unmarshal(body)
		if err != nil {
			return false
		}
		return when.BodyJSON.Matches(json)
	} else {
		return true
	}
}

func valuesMatch(expected []*mckmaps.NameValuesPair, actual map[string][]string) bool {
	for _, e := range expected {
		formValues := actual[e.Name]

		if !stringSlicesEqualIgnoreOrder(e.Values, formValues) {
			return false
		}
	}

	return true
}

// tests if two []string share same elements, ignoring the order
func stringSlicesEqualIgnoreOrder(l, r []string) bool {
	if len(l) != len(r) {
		return false
	}

	diff := make(map[string]int, len(l))
	for _, _x := range l {
		diff[_x]++
	}

	for _, _y := range r {
		if _, ok := diff[_y]; !ok {
			return false
		}

		diff[_y] -= 1
		if diff[_y] == 0 {
			delete(diff, _y)
		}
	}

	return len(diff) == 0
}

func regexpsMatch(expected []*mckmaps.NameRegexpPair, actual map[string][]string) bool {
	for _, e := range expected {
		formValues := actual[e.Name]

		if !regexpAnyMatches(e.Regexp, formValues) {
			return false
		}
	}

	return true
}

func regexpAnyMatches(r *regexp.Regexp, values []string) bool {
	for _, v := range values {
		if r.Match([]byte(v)) {
			return true
		}
	}
	return false
}

func asJSONsMatch(expected []*mckmaps.NameJSONPair, actual map[string][]string) bool {
	for _, e := range expected {
		formValues := actual[e.Name]

		if !asJSONAnyMatches(e.JSON, formValues) {
			return false
		}
	}

	return true
}

func asJSONAnyMatches(m myjson.ExtJSONMatcher, values []string) bool {
	for _, v := range values {
		json, err := myjson.Unmarshal([]byte(v))
		if err != nil {
			continue
		}

		if m.Matches(json) {
			return true
		}
	}
	return false
}
