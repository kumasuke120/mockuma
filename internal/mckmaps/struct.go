package mckmaps

import (
	"regexp"

	"github.com/kumasuke120/mockuma/internal/myhttp"
	"github.com/kumasuke120/mockuma/internal/myjson"
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
	HeaderRegexps []*NameRegexpPair
	HeaderJsons   []*NameJsonPair

	Params       []*NameValuesPair
	ParamRegexps []*NameRegexpPair
	ParamJsons   []*NameJsonPair

	Body       []byte
	BodyRegexp *regexp.Regexp
	BodyJson   *myjson.ExtJsonMatcher
}

type Returns struct {
	StatusCode myhttp.StatusCode
	Headers    []*NameValuesPair
	Body       []byte
	Latency    *Interval
}

type NameValuesPair struct {
	Name   string
	Values []string
}

type NameRegexpPair struct {
	Name   string
	Regexp *regexp.Regexp
}

type NameJsonPair struct {
	Name string
	Json myjson.ExtJsonMatcher
}

type Interval struct {
	Min int64
	Max int64
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
