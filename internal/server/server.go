package server

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/kumasuke120/mockuma/internal/mckmaps"
)

type MockServer struct {
	port       int
	serverChan chan *http.Server
}

func NewMockServer(port int) *MockServer {
	s := new(MockServer)
	s.port = port
	s.serverChan = make(chan *http.Server)
	return s
}

func (s *MockServer) SetMappings(mappings *mckmaps.MockuMappings) {
	if mappings == nil {
		panic("parameter 'mappings' should not be nil")
	}

	if ok := s.shutdown(); ok {
		log.Println("[server] restarting with the new MockuMappings...")
		s.Start(mappings)
	}
}

func (s *MockServer) Start(mappings *mckmaps.MockuMappings) {
	if mappings == nil {
		panic("parameter 'mappings' should not be nil")
	}

	handler := newMockHandler(mappings)
	handler.listAllMappings()

	portStr := strconv.Itoa(s.port)
	addr := ":" + strconv.Itoa(s.port)
	server := &http.Server{Addr: addr, Handler: handler}

	go func() {
		log.Println("[server] listening on " + portStr + "...")
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatalln("[server] fail to start:", err)
			}
		}
	}()

	s.serverChan <- server
}

func (s *MockServer) shutdown() bool {
	select {
	case server := <-s.serverChan:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalln("[server] cannot shutdown server to restart with new mappings:", err)
		}
		return true
	default:
		return false
	}
}
