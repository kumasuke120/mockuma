package mapping

import "github.com/kumasuke120/mockuma/internal/myhttp"

type MockuMappings struct {
	mappings map[string][]*MockuMapping
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
	Body       string
}

var PolicyReturnsNotFound = PolicyReturns{
	StatusCode: myhttp.NotFound,
	Headers: &Headers{headers: map[string][]string{
		myhttp.HeaderContentType: {myhttp.ContentTypeJson},
	}},
	Body: `{"statusCode": 404, "message": "Not Found"}`,
}

var PolicyReturnsNoPolicyMatch = PolicyReturns{
	StatusCode: myhttp.BadRequest,
	Headers: &Headers{headers: map[string][]string{
		myhttp.HeaderContentType: {myhttp.ContentTypeJson},
	}},
	Body: `{"statusCode": 400, "message": "No policy matched"}`,
}
