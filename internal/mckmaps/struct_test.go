package mckmaps

import (
	"testing"

	"github.com/kumasuke120/mockuma/internal/myhttp"
	"github.com/stretchr/testify/assert"
)

var mappings = &MockuMappings{Mappings: []*Mapping{
	{
		URI:      "/a1",
		Method:   myhttp.MethodPut,
		Policies: nil,
	},
	{
		URI:      "/a2",
		Method:   myhttp.MethodPost,
		Policies: nil,
	},
	{
		URI:      "/a1",
		Method:   myhttp.MethodGet,
		Policies: nil,
	},
}}

func TestMockuMappings_GroupMethodsByURI(t *testing.T) {
	expected := map[string][]myhttp.HTTPMethod{
		"/a1": {myhttp.MethodPut, myhttp.MethodGet},
		"/a2": {myhttp.MethodPost},
	}
	actual := mappings.GroupMethodsByURI()

	assert.Equal(t, expected, actual)
}

func TestMockuMappings_IsEmpty(t *testing.T) {
	assert.False(t, mappings.IsEmpty())
	assert.True(t, new(MockuMappings).IsEmpty())
}
