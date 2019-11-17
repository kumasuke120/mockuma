package serve

import (
	"log"
	"net/http"
	"strconv"

	"github.com/kumasuke120/mockuma/internal/mapping"
)

type MockServer struct {
	Port     int
	Mappings *mapping.MockuMappings
}

func (s *MockServer) Start() error {
	handler := &mockHandler{mappings: s.Mappings}
	handler.listAllMappings()

	portStr := strconv.Itoa(s.Port)
	log.Println("[server] listening on " + portStr + "...")

	addr := ":" + portStr
	err := http.ListenAndServe(addr, handler)
	if err != nil {
		return err
	}

	return nil
}
