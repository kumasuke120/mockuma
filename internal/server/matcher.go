package server

import (
	"log"
	"net/http"
	"net/url"
	"regexp"

	"github.com/kumasuke120/mockuma/internal/mckmaps"
)

type pathMatcher struct {
	uri2mappings map[string][]*mckmaps.Mapping
}

func newPathMatcher(mappings *mckmaps.MockuMappings) *pathMatcher {
	uri2mappings := make(map[string][]*mckmaps.Mapping)
	for _, m := range mappings.Mappings {
		mappingsOfUri := uri2mappings[m.Uri]
		mappingsOfUri = append(mappingsOfUri, m)
		uri2mappings[m.Uri] = mappingsOfUri
	}
	return &pathMatcher{uri2mappings: uri2mappings}
}

func (m *pathMatcher) bind(r *http.Request) *boundMatcher {
	return &boundMatcher{m: m, r: r}
}

type boundMatcher struct {
	m              *pathMatcher
	r              *http.Request
	matchedMapping *mckmaps.Mapping
}

func (bm *boundMatcher) matches() bool {
	uri := getUriWithoutQuery(bm.r.URL)

	if mappingsOfUri, ok := bm.m.uri2mappings[uri]; ok {
		for _, mappingOfUri := range mappingsOfUri {
			if mappingOfUri.Method.Matches(bm.r.Method) {
				bm.matchedMapping = mappingOfUri
				return true
			}
		}
	}

	return false
}

func getUriWithoutQuery(url0 *url.URL) string {
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
			if !allMatch(when.Params, bm.r.Form) {
				continue
			}
			if !allMatch(when.Headers, bm.r.Header) {
				continue
			}
			if !regexpAllMatch(when.ParamRegexps, bm.r.Form) {
				continue
			}
			if !regexpAllMatch(when.HeaderRegexps, bm.r.Header) {
				continue
			}
		}

		policy = p
		break
	}
	return policy
}

func allMatch(expected []*mckmaps.NameValuesPair, actual map[string][]string) bool {
	for _, e := range expected {
		formValues := actual[e.Name]

		if !valuesMatch(e.Values, formValues) {
			return false
		}
	}

	return true
}

func valuesMatch(l, r []string) bool { // tests if two []string share same elements, ignoring the order
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

func regexpAllMatch(expected []*mckmaps.NameRegexpPair, actual map[string][]string) bool {
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
