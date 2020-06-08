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

var HeaderValueServer = fmt.Sprintf("%s/%s", internal.AppName, internal.VersionNumber)

type mockHandler struct {
	mappings    *mckmaps.MockuMappings
	pathMatcher *pathMatcher
}

func newMockHandler(mappings *mckmaps.MockuMappings) http.Handler {
	h := new(mockHandler)
	h.mappings = mappings
	h.pathMatcher = newPathMatcher(mappings)

	h.listAllMappings()

	corsOption := mappings.Config.CORS
	if corsOption.Enabled {
		log.Println("[handler ] enabled  : cors handler")
		return corsOption.ToCors().Handler(h)
	} else {
		return h
	}
}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(myhttp.HeaderServer, HeaderValueServer)

	executor := h.matchNewExecutor(r, w)
	if err := executor.execute(); err != nil {
		h.handleExecuteError(w, r, err)
	}
}

func (h *mockHandler) matchNewExecutor(r *http.Request, w http.ResponseWriter) *policyExecutor {
	executor := &policyExecutor{h: h, r: r, w: &w}

	matcher := h.pathMatcher.bind(r)
	if matcher.matches() {
		executor.returnHead = matcher.headMatches()
		executor.policy = matcher.matchPolicy()
	} else {
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
