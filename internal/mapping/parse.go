package mapping

import (
	"encoding/json"
	"fmt"

	"github.com/kumasuke120/mockuma/internal/myhttp"
)

type JsonParseError struct {
	jsonpath string
}

func (e *JsonParseError) Error() string {
	return fmt.Sprintf("Cannot parse value on jsonpath '%s", e.jsonpath)
}

func parseFromJson(jsonData []byte) (*MockuMappings, error) {
	var v interface{}
	err := json.Unmarshal(jsonData, &v)
	if err != nil {
		return nil, err
	}

	data, ok := v.([]interface{})
	if !ok {
		return nil, &JsonParseError{jsonpath: "$"}
	}
	mappingsMap, err := parseAsMockuMappingMap(data)
	if err != nil {
		return nil, err
	}

	return &MockuMappings{mappings: mappingsMap}, nil
}

func parseAsMockuMappingMap(data []interface{}) (map[string][]*MockuMapping, error) {
	mappings := make(map[string][]*MockuMapping)
	for i, mappingData := range data {
		_mappingData, ok := mappingData.(map[string]interface{})
		if !ok {
			return nil, &JsonParseError{jsonpath: fmt.Sprintf("$[%d]", i)}
		}
		mapping, err := parseAsMockuMapping(i, _mappingData)
		if err != nil {
			return nil, err
		}

		mappingsOfUri := mappings[mapping.Uri]
		mappingsOfUri = append(mappingsOfUri, mapping)
		mappings[mapping.Uri] = mappingsOfUri
	}

	return mappings, nil
}

func parseAsMockuMapping(i int, mappingData map[string]interface{}) (*MockuMapping, error) {
	var ok bool

	mapping := new(MockuMapping)
	if mapping.Uri, ok = mappingData["uri"].(string); !ok {
		return nil, &JsonParseError{jsonpath: fmt.Sprintf("$[%d].uri", i)}
	}
	method := myhttp.ToHttpMethod(mappingData["method"])
	mapping.Method = method

	var policies []*Policy
	var policiesData []interface{}
	if policiesData, ok = mappingData["policies"].([]interface{}); !ok {
		return nil, &JsonParseError{jsonpath: fmt.Sprintf("$[%d].policies", i)}
	}
	for j, policyData := range policiesData {
		var _policyData map[string]interface{}
		if _policyData, ok = policyData.(map[string]interface{}); !ok {
			return nil, &JsonParseError{jsonpath: fmt.Sprintf("$[%d].policies[%d]", i, j)}
		}

		policy, err := parseAsPolicy([]interface{}{i, j}, _policyData)
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}
	mapping.Policies = &Policies{policies: policies}

	return mapping, nil
}

func parseAsPolicy(idx []interface{}, policyData map[string]interface{}) (*Policy, error) {
	when, err := parseAsPolicyWhen(idx, policyData)
	if err != nil {
		return nil, err
	}
	returns, err := parseAsPolicyReturns(idx, policyData)
	if err != nil {
		return nil, err
	}

	policy := new(Policy)
	policy.When = when
	policy.Returns = returns
	return policy, nil
}

func parseAsPolicyWhen(idx []interface{}, policyData map[string]interface{}) (*PolicyWhen, error) {
	var ok bool

	if policyData["when"] == nil {
		return nil, nil
	}

	var whenData map[string]interface{}
	if whenData, ok = policyData["when"].(map[string]interface{}); !ok {
		return nil, &JsonParseError{jsonpath: fmt.Sprintf("$[%d].policies[%d].when", idx...)}
	}
	var paramsData map[string]interface{}
	if whenData["params"] == nil {
		paramsData = make(map[string]interface{}, 0)
	} else {
		if paramsData, ok = whenData["params"].(map[string]interface{}); !ok {
			return nil, &JsonParseError{jsonpath: fmt.Sprintf("$[%d].policies[%d].when.params", idx...)}
		}
	}

	params := make(map[string][]string)
	for name, rawValue := range paramsData {
		params[name] = parseAsValues(rawValue)
	}

	when := new(PolicyWhen)
	when.Params = params
	return when, nil
}

func parseAsPolicyReturns(idx []interface{}, policyData map[string]interface{}) (*PolicyReturns, error) {
	var ok bool

	var returnsData map[string]interface{}
	if policyData["returns"] == nil {
		returnsData = make(map[string]interface{}, 0)
	} else {
		if returnsData, ok = policyData["returns"].(map[string]interface{}); !ok {
			return nil, &JsonParseError{jsonpath: fmt.Sprintf("$[%d].policies[%d].returns", idx...)}
		}
	}

	var statusCode int
	if returnsData["statusCode"] == nil {
		statusCode = int(myhttp.Ok)
	} else {
		var _statusCode float64
		if _statusCode, ok = returnsData["statusCode"].(float64); !ok {
			return nil, &JsonParseError{jsonpath: fmt.Sprintf("$[%d].policies[%d].statusCode", idx...)}
		}
		statusCode = int(_statusCode)
	}

	var headersData map[string]interface{}
	if returnsData["headers"] == nil {
		headersData = make(map[string]interface{}, 0)
	} else {
		if headersData, ok = returnsData["headers"].(map[string]interface{}); !ok {
			return nil, &JsonParseError{jsonpath: fmt.Sprintf("$[%d].policies[%d].returns.headers", idx...)}
		}
	}
	headers := make(map[string][]string)
	for name, rawValue := range headersData {
		headers[name] = parseAsValues(rawValue)
	}

	body, err := parseAsBody(idx, returnsData)
	if err != nil {
		return nil, err
	}

	returns := new(PolicyReturns)
	returns.StatusCode = myhttp.StatusCode(statusCode)
	returns.Headers = &Headers{headers: headers}
	returns.Body = body
	return returns, nil
}

func parseAsBody(idx []interface{}, returnsData map[string]interface{}) (string, error) {
	bodyData := returnsData["body"]

	switch bodyData.(type) {
	case nil:
		return "", nil
	case string:
		return bodyData.(string), nil
	case map[string]interface{}:
		return marshalJsonBodyData(bodyData)
	case []interface{}:
		return marshalJsonBodyData(bodyData)
	}

	return "", &JsonParseError{jsonpath: fmt.Sprintf("$[%d].policies[%d].returns.body", idx...)}
}

func marshalJsonBodyData(bodyData interface{}) (string, error) {
	bytes, err := json.Marshal(bodyData)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func parseAsValues(rawValue interface{}) []string {
	var result []string

	switch rawValue.(type) {
	case nil:
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
