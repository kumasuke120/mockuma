package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMockServer(t *testing.T) {
	s := NewMockServer(1234, mappings)
	assert.Equal(t, s.port, 1234)
}

func TestMockServer_SetNameAndVersion(t *testing.T) {
	s := NewMockServer(1234, mappings)
	s.SetNameAndVersion("a", "1")
	assert.Equal(t, "a/1", s.handler.serverHeader)
}
