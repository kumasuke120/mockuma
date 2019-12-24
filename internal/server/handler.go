package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kumasuke120/mockuma/internal"
	"github.com/kumasuke120/mockuma/internal/mckmaps"
	"github.com/kumasuke120/mockuma/internal/myhttp"
)

// default policies
var (
	pNotFound         = newStatusJSONPolicy(myhttp.StatusNotFound, "Not Found")
	pNoPolicyMatched  = newStatusJSONPolicy(myhttp.StatusBadRequest, "No policy matched")
	pMethodNotAllowed = newStatusJSONPolicy(myhttp.StatusMethodNotAllowed, "Method not allowed")
)

func newStatusJSONPolicy(statusCode myhttp.StatusCode, message string) *mckmaps.Policy {
	return &mckmaps.Policy{
		Returns: &mckmaps.Returns{
			StatusCode: statusCode,
			Headers: []*mckmaps.NameValuesPair{
				{Name: myhttp.HeaderContentType, Values: []string{myhttp.ContentTypeJSON}},
			},
			Body: []byte(fmt.Sprintf(`{"statusCode": %d, "message": "%s"}`, statusCode, message)),
		},
	}
}

var HeaderValueServer = fmt.Sprintf("%s/%s", internal.AppName, internal.VersionNumber)

type mockHandler struct {
	mappings    *mckmaps.MockuMappings
	pathMatcher *pathMatcher
}

func newMockHandler(mappings *mckmaps.MockuMappings) *mockHandler {
	handler := new(mockHandler)
	handler.mappings = mappings
	handler.pathMatcher = newPathMatcher(mappings)
	return handler
}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(myhttp.HeaderServer, HeaderValueServer)

	matcher := h.pathMatcher.bind(r)
	executor := &policyExecutor{r: r, w: &w}

	if matcher.matches() {
		policy := matcher.matchPolicy()
		if policy != nil {
			executor.policy = policy
		} else {
			executor.policy = pNoPolicyMatched
		}
	} else if matcher.isMethodNotAllowed() {
		executor.policy = pMethodNotAllowed
	} else {
		executor.policy = pNotFound
	}

	if err := executor.execute(); err != nil {
		log.Printf("[handler] %s %s: fail to render response: %v\n", r.Method, r.URL, err)
	}
}

func (h *mockHandler) listAllMappings() {
	for uri, methods := range h.mappings.GroupMethodsByURI() {
		log.Printf("[handler] mapped: %s, methods = %v\n", uri, methods)
	}
}
