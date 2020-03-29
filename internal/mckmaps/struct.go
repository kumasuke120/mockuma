package mckmaps

import (
	"regexp"

	"github.com/kumasuke120/mockuma/internal/myhttp"
	"github.com/kumasuke120/mockuma/internal/myjson"
)

type MockuMappings struct {
	Mappings  []*Mapping
	Filenames []string
}

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

func (m *MockuMappings) IsEmpty() bool {
	return len(m.Mappings) == 0 && len(m.Filenames) == 0
}

func (m *MockuMappings) GroupMethodsByURI() map[string][]myhttp.HTTPMethod {
	result := make(map[string][]myhttp.HTTPMethod)
	for _, m := range m.Mappings {
		mappingsOfURI := result[m.URI]
		mappingsOfURI = append(mappingsOfURI, m.Method)
		result[m.URI] = mappingsOfURI
	}
	return result
}
