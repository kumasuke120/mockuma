package mckmaps

import (
	"testing"

	"github.com/kumasuke120/mockuma/internal/myhttp"
	"github.com/stretchr/testify/assert"
)

func TestMockuMappings_GetUriWithItsMethods(t *testing.T) {
	mappings := &MockuMappings{Mappings: []*Mapping{
		{
			Uri:      "/a1",
			Method:   myhttp.Put,
			Policies: nil,
		},
		{
			Uri:      "/a2",
			Method:   myhttp.Post,
			Policies: nil,
		},
		{
			Uri:      "/a1",
			Method:   myhttp.Get,
			Policies: nil,
		},
	}}
	expected := map[string][]myhttp.HttpMethod{
		"/a1": {myhttp.Put, myhttp.Get},
		"/a2": {myhttp.Post},
	}
	actual := mappings.GetUriWithItsMethods()

	assert.Equal(t, expected, actual)
}
