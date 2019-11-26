package mckmaps

import (
	"regexp"

	"github.com/kumasuke120/mockuma/internal/myhttp"
)

type MockuMappings struct {
	Mappings []*Mapping
}

type Mapping struct {
	Uri      string
	Method   myhttp.HttpMethod
	Policies []*Policy
}

type Policy struct {
	When    *When
	Returns *Returns
}

type When struct {
	Headers       []*NameValuesPair
	Params        []*NameValuesPair
	HeaderRegexps []*NameRegexpPair
	ParamRegexps  []*NameRegexpPair
}

type Returns struct {
	StatusCode myhttp.StatusCode
	Headers    []*NameValuesPair
	Body       []byte
}

type NameValuesPair struct {
	Name   string
	Values []string
}

type NameRegexpPair struct {
	Name   string
	Regexp *regexp.Regexp
}

func (m *MockuMappings) GetUriWithItsMethods() map[string][]myhttp.HttpMethod {
	result := make(map[string][]myhttp.HttpMethod)
	for _, m := range m.Mappings {
		mappingsOfUri := result[m.Uri]
		mappingsOfUri = append(mappingsOfUri, m.Method)
		result[m.Uri] = mappingsOfUri
	}
	return result
}
