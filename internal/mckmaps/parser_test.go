package mckmaps

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/kumasuke120/mockuma/internal/myhttp"
	"github.com/kumasuke120/mockuma/internal/myjson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewParser(t *testing.T) {
	p1 := NewParser("123")
	assert.Equal(t, "123", p1.filename)
}

//noinspection GoImportUsedAsName
func TestParser_Parse(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	loadedFilenames = nil

	oldWd, err := os.Getwd()
	require.Nil(err)

	require.Nil(os.Chdir("testdata"))

	expectedMappings := []*Mapping{
		{
			URI:    "/m1",
			Method: myhttp.HTTPMethod("RESET"),
			Policies: []*Policy{
				{
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

	require.Nil(os.Chdir(oldWd))
}

//noinspection GoImportUsedAsName
func TestMappingsParser_parse(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fb0, e0 := ioutil.ReadFile(filepath.Join("testdata", "mappings-0.json"))
	require.Nil(e0)
	j0, e0 := myjson.Unmarshal(fb0)
	if assert.Nil(e0) {
		m0 := &mappingsParser{json: j0}
		_, e0 := m0.parse()
		assert.NotNil(e0)
	}

	fb1, e1 := ioutil.ReadFile(filepath.Join("testdata", "mappings-1.json"))
	require.Nil(e1)
	j1, e1 := myjson.Unmarshal(fb1)
	if assert.Nil(e1) {
		m1 := &mappingsParser{json: j1}
		p1, e1 := m1.parse()
		if assert.Nil(e1) {
			expected1 := []*Mapping{
				{
					URI:    "/",
					Method: myhttp.MethodAny,
					Policies: []*Policy{
						{
							Returns: &Returns{
								StatusCode: myhttp.StatusOk,
								Headers: []*NameValuesPair{
									{
										Name:   myhttp.HeaderContentType,
										Values: []string{"application/json; charset=utf8"},
									},
								},
								Body: []byte("abc123"),
								Latency: &Interval{
									Min: 100,
									Max: 3000,
								},
							},
						},
					},
				},
			}
			assert.Equal(expected1, p1)
		}
	}

	fb2, e2 := ioutil.ReadFile(filepath.Join("testdata", "mappings-2.json"))
	require.Nil(e2)
	j2, e2 := myjson.Unmarshal(fb2)
	if assert.Nil(e2) {
		m2 := &mappingsParser{json: j2}
		p2, e2 := m2.parse()
		if assert.Nil(e2) {
			expected2 := []*Mapping{
				{
					URI:    "/",
					Method: myhttp.MethodGet,
					Policies: []*Policy{
						{
							When: &When{
								Headers: []*NameValuesPair{
									{
										Name:   "X-BC",
										Values: []string{"2", "3"},
									},
								},
								HeaderRegexps: []*NameRegexpPair{
									{
										Name:   "X-R",
										Regexp: regexp.MustCompile("^.+$"),
									},
								},
								HeaderJSONs: []*NameJSONPair{
									{
										Name: "X-J",
										JSON: myjson.MakeExtJSONMatcher(myjson.Object(map[string]interface{}{
											"v": myjson.String("v"),
										})),
									},
								},
								Params: []*NameValuesPair{
									{
										Name:   "bc",
										Values: []string{"2", "3"},
									},
								},
								ParamRegexps: []*NameRegexpPair{
									{
										Name:   "r",
										Regexp: regexp.MustCompile("^.+$"),
									},
								},
								ParamJSONs: []*NameJSONPair{
									{
										Name: "j",
										JSON: myjson.MakeExtJSONMatcher(myjson.Object(map[string]interface{}{
											"v": myjson.String("v"),
										})),
									},
								},
								Body: []byte("123"),
							},
							Returns: &Returns{
								StatusCode: myhttp.StatusCode(201),
								Headers: []*NameValuesPair{
									{
										Name:   myhttp.HeaderContentType,
										Values: []string{"application/json; charset=utf8"},
									},
								},
								Body: []byte(`{"v":"v"}`),
								Latency: &Interval{
									Min: 100,
									Max: 100,
								},
							},
						},
						{
							Returns: &Returns{
								StatusCode: myhttp.StatusOk,
								Body:       []byte(""),
								Latency: &Interval{
									Min: 100,
									Max: 100,
								},
							},
						},
					},
				},
			}
			assert.Equal(expected2, p2)
		}
	}

	fb3, e3 := ioutil.ReadFile(filepath.Join("testdata", "mappings-3.json"))
	require.Nil(e3)
	j3, e3 := myjson.Unmarshal(fb3)
	if assert.Nil(e3) {
		m3 := &mappingsParser{json: j3}
		_, e3 = m3.parse()
		assert.NotNil(e3)
		assert.NotEmpty(e3.Error())
	}

	fb4, e4 := ioutil.ReadFile(filepath.Join("testdata", "mappings-4.json"))
	require.Nil(e4)
	j4, e4 := myjson.Unmarshal(fb4)
	if assert.Nil(e4) {
		m4 := &mappingsParser{json: j4}
		_, e4 = m4.parse()
		assert.NotNil(e4)
		assert.NotEmpty(e4.Error())
	}

	fb5, e5 := ioutil.ReadFile(filepath.Join("testdata", "mappings-5.json"))
	require.Nil(e5)
	j5, e5 := myjson.Unmarshal(fb5)
	if assert.Nil(e5) {
		m5 := &mappingsParser{json: j5}
		p5, e5 := m5.parse()
		if assert.Nil(e5) {
			expected5 := []*Mapping{
				{
					URI:    "/",
					Method: myhttp.MethodGet,
					Policies: []*Policy{
						{
							When: &When{
								BodyRegexp: regexp.MustCompile("^.+$"),
							},
							Returns: &Returns{StatusCode: myhttp.StatusOk},
						},
						{
							When: &When{
								BodyJSON: myjson.NewExtJSONMatcher(myjson.Object{
									"v": myjson.String("v"),
								}),
							},
							Returns: &Returns{StatusCode: myhttp.StatusOk},
						},
					},
				},
			}
			assert.Equal(expected5, p5)
		}
	}
}

func TestTemplateParser_parse(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	t0 := &templateParser{Parser: Parser{filename: filepath.Join("testdata", "template-0.json")}}
	_, e0 := t0.parse()
	assert.NotNil(e0)

	t1 := &templateParser{Parser: Parser{filename: filepath.Join("testdata", "template-1.json")}}
	p1, e1 := t1.parse()
	if assert.Nil(e1) {
		expected1 := &template{
			content: myjson.Object{
				"v":   myjson.String("@{v}"),
				"v-v": myjson.String("@{v}-@{v}"),
			},
		}
		assert.Equal(expected1, p1)
	}
}

func TestVarsParser_parse(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	t0 := &varsParser{Parser: Parser{filename: filepath.Join("testdata", "vars-0.json")}}
	_, e0 := t0.parse()
	assert.NotNil(e0)

	t1 := &varsParser{Parser: Parser{filename: filepath.Join("testdata", "vars-1.json")}}
	p1, e1 := t1.parse()
	if assert.Nil(e1) {
		expected1 := []*vars{
			{
				table: map[string]interface{}{
					"v": myjson.String("1"),
				},
			},
			{
				table: map[string]interface{}{
					"v": myjson.String("2"),
					"a": myjson.String("b"),
				},
			},
		}
		assert.Equal(expected1, p1)
	}
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
