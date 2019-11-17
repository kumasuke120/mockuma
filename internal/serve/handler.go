package serve

import (
	"log"
	"net/http"

	"github.com/kumasuke120/mockuma/internal/mapping"
	"github.com/kumasuke120/mockuma/internal/myhttp"
)

type mockHandler struct {
	mappings *mapping.MockuMappings
}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(myhttp.HeaderServer, serverName)
	returns := &mapping.PolicyReturnsNotFound

	_mapping := h.mappings.Match(r)
	if _mapping != nil {
		policy := _mapping.Policies.MatchFirst(r)

		if policy != nil {
			returns = policy.Returns
		} else {
			returns = &mapping.PolicyReturnsNoPolicyMatch
		}
	}

	err := returns.Render(&w)
	if err != nil {
		log.Printf("[handler] (err) %s %s: Fail to render response: %v\n",
			r.Method, r.URL, err)
	}

	log.Printf("[handler] (%d) %s %s\n", returns.StatusCode, r.Method, r.URL)
}

func (h *mockHandler) listAllMappings() {
	for uri, mappings := range h.mappings.Mappings {
		log.Printf("[handler] mapped: %s, methods = %s\n", uri, getSupportedMethods(mappings))
	}
}

func getSupportedMethods(mappingsOfUri []*mapping.MockuMapping) []myhttp.HttpMethod {
	if mappingsOfUri == nil {
		return []myhttp.HttpMethod{}
	} else {
		var result []myhttp.HttpMethod
		for _, _mapping := range mappingsOfUri {
			result = append(result, _mapping.Method)
		}
		return result
	}
}
