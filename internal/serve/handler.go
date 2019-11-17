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

	theMapping := h.mappings.Match(r)
	if theMapping != nil {
		policy := theMapping.Policies.MatchFirst(r)

		if policy != nil {
			returns = policy.Returns
		} else {
			returns = &mapping.PolicyReturnsNoPolicyMatch
		}
	}

	err := returns.Render(w)
	if err != nil {
		log.Println("Fail to write response: ", err)
	}

	log.Println(returns.StatusCode, r.URL)
}
