package server

import (
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
		time.Sleep(3 * time.Second)
		s.shutdown()
	}()
	s.Start(mappings)
}
