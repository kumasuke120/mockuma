package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/kumasuke120/mockuma/internal/mckmaps"
)

type MockServer struct {
	port      int
	server    *http.Server
	serverMux sync.Mutex
}

func NewMockServer(port int) *MockServer {
	s := new(MockServer)
	s.port = port
	return s
}

func (s *MockServer) ListenAndServe(mappings *mckmaps.MockuMappings) {
	if mappings == nil {
		panic("parameter 'mappings' should not be nil")
	}

	handler := newMockHandler(mappings)
	addr := fmt.Sprintf(":%d", s.port)
	server := &http.Server{Addr: addr, Handler: handler}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Println("[server  ] listening on " + strconv.Itoa(s.port) + "...")
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatalln("[server  ] cannot start:", err)
			}
		}
	}()

	s.setServer(server)
	wg.Wait()
}

func (s *MockServer) SetMappings(mappings *mckmaps.MockuMappings) {
	if mappings == nil {
		panic("parameter 'mappings' should not be nil")
	}

	if ok := s.shutdown(); ok {
		log.Println("[server  ] restarting with the new mockuMappings...")
		go s.ListenAndServe(mappings)
	}
}

func (s *MockServer) shutdown() bool {
	server := s.getServer()
	if server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalln("[server  ] cannot shutdown server in order to restart with the new mappings:", err)
		}

		return true
	}

	return false
}

func (s *MockServer) getServer() *http.Server {
	s.serverMux.Lock()
	defer s.serverMux.Unlock()
	return s.server
}

func (s *MockServer) setServer(server *http.Server) {
	s.serverMux.Lock()
	defer s.serverMux.Unlock()
	s.server = server
}
