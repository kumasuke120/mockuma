package server

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/kumasuke120/mockuma/internal"
	"github.com/kumasuke120/mockuma/internal/mckmaps"
	"github.com/kumasuke120/mockuma/internal/myhttp"
)

// default policies
var (
	pNotFound            = newStatusJSONPolicy(myhttp.StatusNotFound, "Not Found")
	pNoPolicyMatched     = newStatusJSONPolicy(myhttp.StatusBadRequest, "No policy matched")
	pMethodNotAllowed    = newStatusJSONPolicy(myhttp.StatusMethodNotAllowed, "Method Not Allowed")
	pInternalServerError = newStatusJSONPolicy(myhttp.StatusInternalServerError, "Internal Server Error")
	pBadGateway          = newStatusJSONPolicy(myhttp.StatusBadGateway, "Bad Gateway")
)

func newStatusJSONPolicy(statusCode myhttp.StatusCode, message string) *mckmaps.Policy {
	return &mckmaps.Policy{
		CmdType: mckmaps.CmdTypeReturns,
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

	executor := h.matchNewExecutor(r, w)
	if err := executor.execute(); err != nil {
		h.handleExecuteError(w, r, err)
	}
}

func (h *mockHandler) matchNewExecutor(r *http.Request, w http.ResponseWriter) *policyExecutor {
	matcher := h.pathMatcher.bind(r)
	executor := &policyExecutor{h: h, r: r, w: &w}

	switch matcher.match() {
	case MatchHead:
		executor.returnHead = true
		fallthrough
	case MatchExact:
		policy := matcher.matchPolicy()
		if policy != nil {
			executor.policy = policy
		} else {
			executor.policy = pNoPolicyMatched
		}
	case MatchURI:
		executor.policy = pMethodNotAllowed
	case MatchNone:
		executor.policy = pNotFound
	}

	return executor
}

func (h *mockHandler) handleExecuteError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("[handler ] error    : %s %s => %v\n", r.Method, r.URL, err)

	switch err.(type) {
	case *forwardError:
		executor := &policyExecutor{h: h, r: r, w: &w, policy: pBadGateway}
		err = executor.execute()
	default:
		executor := &policyExecutor{h: h, r: r, w: &w, policy: pInternalServerError}
		err = executor.execute()
	}

	if err != nil {
		log.Printf("[handler ] error    : %s %s => fail to render response: %v\n", r.Method, r.URL, err)
	}
}

func (h *mockHandler) listAllMappings() {
	uri2Methods := h.mappings.GroupMethodsByURI()

	var uris []string
	for uri := range uri2Methods {
		uris = append(uris, uri)
	}
	sort.Strings(uris)

	for _, uri := range uris {
		methods := uri2Methods[uri]
		log.Printf("[handler ] mapped   : %s, methods = %v\n", uri, methods)
	}
}
