package mapping

import (
	"log"
	"net/http"
	"net/url"
)

func (m *MockuMappings) Match(r *http.Request) *MockuMapping {
	uri := getUriWithoutQuery(r.URL)

	mappingsOfUri := m.Mappings[uri]
	if mappingsOfUri != nil {
		for _, mappingOfUri := range mappingsOfUri {
			if mappingOfUri.Method.Matches(r.Method) {
				return mappingOfUri
			}
		}
	}

	return nil
}

func getUriWithoutQuery(url0 *url.URL) string {
	url1 := &url.URL{}
	*url1 = *url0

	url1.RawQuery = ""
	url1.ForceQuery = false
	return url1.Path
}

func (p *Policies) MatchFirst(r *http.Request) *Policy {
	err := r.ParseForm()
	if err != nil {
		log.Println("http: Fail to parse form: ", err)
		return nil
	}

	form := r.Form

	var policy *Policy
	for _, p := range p.policies {
		when := p.When

		if when != nil {
			if !paramsMatches(when.Params, form) {
				continue
			}
		}

		policy = p
		break
	}

	return policy
}

func paramsMatches(params map[string][]string, form url.Values) bool {
	for name, values := range params {
		formValues := form[name]

		if !valuesMatches(values, formValues) {
			return false
		}
	}

	return true
}

func valuesMatches(l, r []string) bool {
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
