package mckmaps

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/kumasuke120/mockuma/internal/myhttp"
	"github.com/kumasuke120/mockuma/internal/myjson"
	"github.com/kumasuke120/mockuma/internal/myos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestError(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	err0 := &loadError{filename: "test.json", err: errors.New("\n\n")}
	assert.NotNil(err0)
	assert.Contains(err0.Error(), "test.json")
	assert.Contains(err0.Error(), "\n\t\n\t")

	err1 := &parserError{
		jsonPath: myjson.NewPath("testPath"),
		filename: "test.json",
		err:      errors.New("test_error\n"),
	}
	assert.NotNil(err1)
	assert.Contains(err1.Error(), "$.testPath")
	assert.Contains(err1.Error(), "test.json")
	assert.Contains(err1.Error(), "test_error")
	assert.Contains(err1.Error(), "test_error\n\t")
}

func TestNewParser(t *testing.T) {
	p1 := NewParser("123")
	assert.Equal(t, "123", p1.filename)
}

//noinspection GoImportUsedAsName
func TestParser_Parse(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	loadedFilenames = nil

	require.Nil(myos.InitWd())
	oldWd := myos.GetWd()
	require.Nil(myos.Chdir(filepath.Join(oldWd, "testdata", "parser")))

	expectedMappings := []*Mapping{
		{
			URI:    "/m1",
			Method: myhttp.HTTPMethod("RESET"),
			Policies: []*Policy{
				{
					CmdType: mapPolicyReturns,
					Returns: &Returns{
						StatusCode: myhttp.StatusOK,
						Body:       []byte("m1"),
					},
				},
			},
		},
		{
			URI:    "/m2",
			Method: myhttp.MethodPost,
			Policies: []*Policy{
				{
					When: &When{
						Params: []*NameValuesPair{
							{
								Name:   "p",
								Values: []string{"1"},
							},
						},
					},
					CmdType: mapPolicyReturns,
					Returns: &Returns{
						StatusCode: myhttp.StatusOK,
						Body:       []byte("m2:1"),
					},
				},
				{
					When: &When{
						Params: []*NameValuesPair{
							{
								Name:   "p",
								Values: []string{"2"},
							},
						},
					},
					CmdType: mapPolicyReturns,
					Returns: &Returns{
						StatusCode: myhttp.StatusOK,
						Body:       []byte("m2:2"),
					},
				},
				{
					When: &When{
						Params: []*NameValuesPair{
							{
								Name:   "p",
								Values: []string{"3"},
							},
						},
					},
					CmdType: mapPolicyReturns,
					Returns: &Returns{
						StatusCode: myhttp.StatusOK,
						Body:       []byte("m2:3"),
					},
				},
			},
		},
	}

	fn1 := "parser-single.json"
	path1, e1 := filepath.Abs(fn1)
	require.Nil(e1)
	expected1 := &MockuMappings{
		Mappings:  expectedMappings,
		Filenames: []string{fn1},
		Config:    defaultConfig(),
	}
	parser1 := NewParser(path1)
	actual1, e1 := parser1.Parse()
	if assert.Nil(e1) {
		assert.Equal(expected1, actual1)
	}

	fn2 := "parser-multi.json"
	expected2 := &MockuMappings{
		Mappings:  expectedMappings,
		Filenames: []string{fn2, fn1},
		Config:    defaultConfig(),
	}
	parser2 := NewParser(fn2)
	actual2, e2 := parser2.Parse()
	if assert.Nil(e2) {
		assert.Equal(expected2, actual2)
	}

	require.Nil(myos.Chdir(oldWd))
}

func TestParser_sortMappings(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	testdata1 := &MockuMappings{Mappings: []*Mapping{
		{
			URI:    "/",
			Method: myhttp.MethodGet,
			Policies: []*Policy{
				{},
			},
		},
		{
			URI:    "/",
			Method: myhttp.MethodGet,
			Policies: []*Policy{
				{},
			},
		},
		{
			URI:    "/",
			Method: myhttp.MethodPost,
			Policies: []*Policy{
				{},
			},
		},
	}}
	expected1 := &MockuMappings{Mappings: []*Mapping{
		{
			URI:    "/",
			Method: myhttp.MethodGet,
			Policies: []*Policy{
				testdata1.Mappings[0].Policies[0],
				testdata1.Mappings[1].Policies[0],
			},
		},
		{
			URI:    "/",
			Method: myhttp.MethodPost,
			Policies: []*Policy{
				testdata1.Mappings[2].Policies[0],
			},
		},
	}}
	p1 := &Parser{filename: ""}
	p1.sortMappings(testdata1)
	assert.Equal(expected1, testdata1)
}
