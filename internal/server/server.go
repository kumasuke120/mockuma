package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/kumasuke120/mockuma/internal/mckmaps"
)

type MockServer struct {
	port    int
	handler *mockHandler
}

func NewMockServer(port int, mappings *mckmaps.MockuMappings) *MockServer {
	s := new(MockServer)
	s.port = port
	s.handler = newMockHandler(mappings)
	return s
}

func (s *MockServer) SetNameAndVersion(name string, versionNumber string) {
	s.handler.serverHeader = fmt.Sprintf("%s/%s", name, versionNumber)
}

func (s *MockServer) Start() error {
	s.handler.listAllMappings()

	portStr := strconv.Itoa(s.port)
	log.Println("[server] listening on " + portStr + "...")

	addr := ":" + portStr
	err := http.ListenAndServe(addr, s.handler)
	if err != nil {
		return err
	}

	return nil
}
