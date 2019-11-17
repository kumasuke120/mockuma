package mapping

import "github.com/kumasuke120/mockuma/internal/myhttp"

type MockuMappings struct {
	Mappings map[string][]*MockuMapping
}

type MockuMapping struct {
	Uri      string
	Method   myhttp.HttpMethod
	Policies *Policies
}

type Policies struct {
	policies []*Policy
}

type Policy struct {
	When    *PolicyWhen
	Returns *PolicyReturns
}

type PolicyWhen struct {
	Params map[string][]string
}

type Headers struct {
	headers map[string][]string
}

type PolicyReturns struct {
	StatusCode myhttp.StatusCode
	Headers    *Headers
	Body       []byte
}

type ReturnsBodyDirectiveType string

const (
	ReadFile = ReturnsBodyDirectiveType("@file")
)

type ReturnsBodyDirective struct {
	directiveType ReturnsBodyDirectiveType
	argument      interface{}
}

var PolicyReturnsNotFound = PolicyReturns{
	StatusCode: myhttp.NotFound,
	Headers: &Headers{headers: map[string][]string{
		myhttp.HeaderContentType: {myhttp.ContentTypeJson},
	}},
	Body: []byte(`{"statusCode": 404, "message": "Not Found"}`),
}

var PolicyReturnsNoPolicyMatch = PolicyReturns{
	StatusCode: myhttp.BadRequest,
	Headers: &Headers{headers: map[string][]string{
		myhttp.HeaderContentType: {myhttp.ContentTypeJson},
	}},
	Body: []byte(`{"statusCode": 400, "message": "No policy matched"}`),
}
