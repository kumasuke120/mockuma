package server

import (
	"log"
	"net/http"

	"github.com/kumasuke120/mockuma/internal/mckmaps"
)

type policyExecutor struct {
	r      *http.Request
	w      *http.ResponseWriter
	policy *mckmaps.Policy
}

func (e *policyExecutor) execute() error {
	returns := e.policy.Returns

	e.writeHeaders(returns.Headers)
	(*e.w).WriteHeader(int(returns.StatusCode)) // statusCode must be written after headers
	err := e.writeBody(returns.Body)
	if err != nil {
		return err
	}

	log.Printf("[server] (%d) %s %s\n", returns.StatusCode, e.r.Method, e.r.URL)
	return nil
}

func (e *policyExecutor) writeHeaders(headers []*mckmaps.NameValuesPair) {
	outHeader := (*e.w).Header()

	for _, pair := range headers {
		if _, ok := outHeader[pair.Name]; ok {
			outHeader.Del(pair.Name)
		}

		for _, value := range pair.Values {
			outHeader.Add(pair.Name, value)
		}
	}
}

func (e *policyExecutor) writeBody(body []byte) error {
	var err error
	if body != nil && len(body) != 0 {
		_, err = (*e.w).Write(body)
	}
	return err
}
