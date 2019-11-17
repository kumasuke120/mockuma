package mapping

import (
	"fmt"
	"go/types"

	"kumasuke.app/mockuma/myhttp"
)

func parseAsMockuMappingMap(data []interface{}) (map[string][]*MockuMapping, error) {
	mappings := make(map[string][]*MockuMapping)
	for _, mappingData := range data {
		mapping := parseAsMockuMapping(mappingData.(map[string]interface{}))
		mappingsOfUri := mappings[mapping.Uri]
		mappingsOfUri = append(mappingsOfUri, mapping)
		mappings[mapping.Uri] = mappingsOfUri
	}
	return mappings, nil
}

func parseAsMockuMapping(mappingData map[string]interface{}) *MockuMapping {
	mapping := new(MockuMapping)
	mapping.Uri = mappingData["uri"].(string)
	method := myhttp.ToHttpMethod(mappingData["method"].(string))
	mapping.Method = method

	var policies []*Policy
	policiesData := mappingData["policies"].([]interface{})
	for _, policyData := range policiesData {
		policy := parseAsMockPolicy(policyData.(map[string]interface{}))
		policies = append(policies, policy)
	}
	mapping.Policies = &Policies{policies: policies}

	return mapping
}

func parseAsMockPolicy(policyData map[string]interface{}) *Policy {
	when := parseAsPolicyWhen(policyData)
	returns := parseAsPolicyReturns(policyData)

	policy := new(Policy)
	policy.When = when
	policy.Returns = returns

	return policy
}

func parseAsPolicyWhen(policyData map[string]interface{}) *PolicyWhen {
	if policyData["when"] == nil {
		return nil
	}

	whenData := policyData["when"].(map[string]interface{})
	paramsData := whenData["params"].(map[string]interface{})

	params := make(map[string][]string)
	for name, rawValue := range paramsData {
		params[name] = parseAsValues(rawValue)
	}

	when := new(PolicyWhen)
	when.Params = params
	return when
}

func parseAsPolicyReturns(policyData map[string]interface{}) *PolicyReturns {
	returnsData := policyData["returns"].(map[string]interface{})

	statusCode := returnsData["statusCode"]
	if statusCode == nil {
		statusCode = int(myhttp.Ok)
	} else {
		statusCode = int(statusCode.(float64))
	}

	headersData := returnsData["headers"].(map[string]interface{})
	headers := make(map[string][]string)
	for name, rawValue := range headersData {
		headers[name] = parseAsValues(rawValue)
	}

	body := returnsData["body"].(string)

	returns := new(PolicyReturns)
	returns.StatusCode = myhttp.StatusCode(statusCode.(int))
	returns.Headers = &Headers{headers: headers}
	returns.Body = body
	return returns
}

func parseAsValues(rawValue interface{}) []string {
	var result []string

	switch rawValue.(type) {
	case types.Nil:
		result = append(result, "")
	case []interface{}:
		result = append(result, toStringSlice(rawValue.([]interface{}))...)
	default:
		result = append(result, toString(rawValue))
	}

	return result
}

func toStringSlice(value []interface{}) []string {
	var result []string

	for _, i := range value {
		result = append(result, toString(i))
	}

	return result
}

func toString(value interface{}) string {
	switch value.(type) {
	case string:
		return value.(string)
	case fmt.Stringer:
		return value.(fmt.Stringer).String()
	default:
		return fmt.Sprintf("%v", value)
	}
}
