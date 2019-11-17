package serve

import (
	"log"
	"net/http"
	"strconv"

	"github.com/kumasuke120/mockuma/mapping"
)

type MockServer struct {
	Port     int
	Mappings *mapping.MockuMappings
}

func (s *MockServer) Start() error {
	handler := &mockHandler{mappings: s.Mappings}

	portStr := strconv.Itoa(s.Port)
	log.Println("Listening on " + portStr + "...")

	addr := ":" + portStr
	err := http.ListenAndServe(addr, handler)
	if err != nil {
		return err
	}

	return nil
}
