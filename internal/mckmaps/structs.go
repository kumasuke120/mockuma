package mckmaps

import "github.com/kumasuke120/mockuma/internal/myhttp"

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
	Headers []*NameValuesPair
	Params  []*NameValuesPair
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
