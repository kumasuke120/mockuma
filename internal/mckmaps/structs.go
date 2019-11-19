package mckmaps

import "github.com/kumasuke120/mockuma/internal/myhttp"

type MockuMappings struct {
	mappings []*Mapping
}

type Mapping struct {
	uri      string
	method   myhttp.HttpMethod
	policies []*Policy
}

type Policy struct {
	when    *When
	returns *Returns
}

type When struct {
	headers []*NameValuesPair
	params  []*NameValuesPair
}

type Returns struct {
	statusCode myhttp.StatusCode
	headers    []*NameValuesPair
	body       []byte
}

type NameValuesPair struct {
	name   string
	values []string
}
