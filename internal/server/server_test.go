package server

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMockServer(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	s := NewMockServer(1234)
	assert.Equal(s.port, 1234)
}

func TestMockServer_Start(t *testing.T) {
	s := NewMockServer(3214)
	go func() {
		time.Sleep(2 * time.Second)
		s.shutdown()
	}()
	s.ListenAndServe(mappings)
}

func TestMockServer_SetMappings(t *testing.T) {
	s := NewMockServer(3214)
	s.SetMappings(mappings)

	var wg sync.WaitGroup
	wg.Add(1)
	go s.ListenAndServe(mappings)
	go func() {
		defer wg.Done()

		time.Sleep(1 * time.Second)
		s.SetMappings(mappings)
		time.Sleep(2 * time.Second)
		s.shutdown()
	}()
	wg.Wait()
}
