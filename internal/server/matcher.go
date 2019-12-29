package server

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"

	"github.com/kumasuke120/mockuma/internal/mckmaps"
	"github.com/kumasuke120/mockuma/internal/myjson"
)

type pathMatcher struct {
	directPath  map[string][]*mckmaps.Mapping
	patternPath map[*regexp.Regexp][]*mckmaps.Mapping
}

var pathVarRegexp = regexp.MustCompile(`{(\d+)}`)

func newPathMatcher(mappings *mckmaps.MockuMappings) *pathMatcher {
	directPath := make(map[string][]*mckmaps.Mapping)
	patternPath := make(map[*regexp.Regexp][]*mckmaps.Mapping)
	for _, m := range mappings.Mappings {
		if theURI := pathVarRegexp.ReplaceAllString(m.URI, "(?P<v$1>.+?)"); theURI == m.URI {
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
		directPath:  directPath,
		patternPath: patternPath,
	}
}

func (m *pathMatcher) bind(r *http.Request) *boundMatcher {
	return &boundMatcher{m: m, r: r}
}

type boundMatcher struct {
	m          *pathMatcher
	r          *http.Request
	uri        string
	uriPattern *regexp.Regexp

	matchedMapping *mckmaps.Mapping
	is405          bool
	bodyCache      []byte
}

func (bm *boundMatcher) matches() bool {
	bm.uri = getURIWithoutQuery(bm.r.URL)

	// matching for direct path
	if mappingsOfURI, ok := bm.m.directPath[bm.uri]; ok {
		matched := bm.anyMethodMatches(mappingsOfURI)
		if matched != nil {
			bm.matchedMapping = matched
			return true
		}
		bm.is405 = true
	}

	// matching for pattern path
	for pattern, mappingsOfURI := range bm.m.patternPath {
		if pattern.MatchString(bm.uri) {
			matched := bm.anyMethodMatches(mappingsOfURI)
			if matched != nil {
				bm.uriPattern = pattern
				bm.matchedMapping = matched
				return true
			}
			bm.is405 = true
		}
	}

	return false
}

func (bm *boundMatcher) anyMethodMatches(mappingsOfURI []*mckmaps.Mapping) *mckmaps.Mapping {
	for _, mappingOfURI := range mappingsOfURI {
		if mappingOfURI.Method.Matches(bm.r.Method) {
			return mappingOfURI
		}
	}
	return nil
}

func (bm *boundMatcher) isMethodNotAllowed() bool {
	return bm.is405
}

func getURIWithoutQuery(url0 *url.URL) string {
	url1 := &url.URL{}
	*url1 = *url0

	url1.RawQuery = ""
	url1.ForceQuery = false
	return url1.Path
}

func (bm *boundMatcher) matchPolicy() *mckmaps.Policy {
	err := bm.r.ParseForm()
	if err != nil {
		log.Println("[server] fail to parse form:", err)
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
	return policy
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
	if body == nil {
		_body, err := ioutil.ReadAll(bm.r.Body)
		if err == nil {
			bm.bodyCache = _body
			body = _body
		}
	}

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
