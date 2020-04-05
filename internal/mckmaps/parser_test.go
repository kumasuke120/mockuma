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
	require.Nil(myos.Chdir(filepath.Join(oldWd, "testdata")))

	expectedMappings := []*Mapping{
		{
			URI:    "/m1",
			Method: myhttp.HTTPMethod("RESET"),
			Policies: []*Policy{
				{
					CmdType: mapPolicyReturns,
					Returns: &Returns{
						StatusCode: myhttp.StatusOk,
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
						StatusCode: myhttp.StatusOk,
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
						StatusCode: myhttp.StatusOk,
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
						StatusCode: myhttp.StatusOk,
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
	actual1 := p1.sortMappings(testdata1)
	assert.Equal(expected1, actual1)
}
