package mckmaps

import (
	"errors"
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

//noinspection GoImportUsedAsName
func TestError(t *testing.T) {
	assert := assert.New(t)

	err0 := &loadError{filename: "test.json"}
	assert.NotNil(err0)
	assert.Contains(err0.Error(), "test.json")

	err1 := &parserError{
		jsonPath: myjson.NewPath("testPath"),
		filename: "test.json",
		err:      errors.New("test_error"),
	}
	assert.NotNil(err1)
	assert.Contains(err1.Error(), "$.testPath")
	assert.Contains(err1.Error(), "test.json")
	assert.Contains(err1.Error(), "test_error")
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

	oldWd, err := os.Getwd()
	require.Nil(err)

	require.Nil(os.Chdir("testdata"))

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
							CmdType: mapPolicyReturns,
							Returns: &Returns{
								StatusCode: myhttp.StatusOk,
								Headers: []*NameValuesPair{
									{
										Name:   myhttp.HeaderContentType,
										Values: []string{"application/json; charset=utf-8"},
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
							CmdType: mapPolicyReturns,
							Returns: &Returns{
								StatusCode: myhttp.StatusCode(201),
								Headers: []*NameValuesPair{
									{
										Name:   myhttp.HeaderContentType,
										Values: []string{"application/json; charset=utf-8"},
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
							CmdType: mapPolicyReturns,
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
							CmdType: mapPolicyReturns,
							Returns: &Returns{StatusCode: myhttp.StatusOk},
						},
						{
							When: &When{
								BodyJSON: myjson.NewExtJSONMatcher(myjson.Object{
									"v": myjson.String("v"),
								}),
							},
							CmdType: mapPolicyReturns,
							Returns: &Returns{StatusCode: myhttp.StatusOk},
						},
					},
				},
			}
			assert.Equal(expected5, p5)
		}
	}

	fb6, e6 := ioutil.ReadFile(filepath.Join("testdata", "mappings-6.json"))
	require.Nil(e6)
	j6, e6 := myjson.Unmarshal(fb6)
	if assert.Nil(e6) {
		m6 := &mappingsParser{json: j6}
		p6, e6 := m6.parse()
		if assert.Nil(e6) {
			expected6 := []*Mapping{
				{
					URI:    "/{0}/{1}/{2}",
					Method: myhttp.MethodAny,
					Policies: []*Policy{
						{
							When: &When{
								PathVars: []*NameValuesPair{
									{
										Name:   "0",
										Values: []string{"1"},
									},
								},
								PathVarRegexps: []*NameRegexpPair{
									{
										Name:   "1",
										Regexp: regexp.MustCompile("\\d+"),
									}, {
										Name:   "2",
										Regexp: regexp.MustCompile("\\w+"),
									},
								},
							},
							CmdType: mapPolicyReturns,
							Returns: &Returns{StatusCode: myhttp.StatusOk},
						},
					},
				},
				{
					URI:    "/{0}/{1}/{0}",
					Method: myhttp.MethodAny,
					Policies: []*Policy{
						{
							When: &When{
								PathVars: []*NameValuesPair{
									{
										Name:   "0",
										Values: []string{"1"},
									},
									{
										Name:   "1",
										Values: []string{"2"},
									},
								},
							},
							CmdType: mapPolicyReturns,
							Returns: &Returns{StatusCode: myhttp.StatusOk},
						},
					},
				},
			}
			assert.Equal(expected6, p6)
		}
	}

	fb7, e7 := ioutil.ReadFile(filepath.Join("testdata", "mappings-7.json"))
	require.Nil(e7)
	j7, e7 := myjson.Unmarshal(fb7)
	if assert.Nil(e7) {
		m7 := &mappingsParser{json: j7}
		p7, e7 := m7.parse()
		if assert.Nil(e7) {
			expected7 := []*Mapping{
				{
					URI:    "/test-for-redirects",
					Method: myhttp.MethodAny,
					Policies: []*Policy{
						{
							When: &When{
								Params: []*NameValuesPair{
									{
										Name:   "no-latency",
										Values: []string{"true"},
									},
								},
							},
							CmdType: mapPolicyRedirects,
							Returns: &Returns{
								StatusCode: myhttp.StatusFound,
								Headers: []*NameValuesPair{
									{
										Name:   myhttp.HeaderLocation,
										Values: []string{"/"},
									},
								},
							},
						},
						{
							CmdType: mapPolicyRedirects,
							Returns: &Returns{
								StatusCode: myhttp.StatusFound,
								Headers: []*NameValuesPair{
									{
										Name:   myhttp.HeaderLocation,
										Values: []string{"/"},
									},
								},
								Latency: &Interval{
									Min: 1000,
									Max: 1000,
								},
							},
						},
					},
				},
			}
			assert.Equal(expected7, p7)
		}
	}

	fb8, e8 := ioutil.ReadFile(filepath.Join("testdata", "mappings-8.json"))
	require.Nil(e8)
	j8, e8 := myjson.Unmarshal(fb8)
	if assert.Nil(e7) {
		m8 := &mappingsParser{json: j8}
		_, e8 := m8.parse()
		assert.NotNil(e8)
	}

	fb9, e9 := ioutil.ReadFile(filepath.Join("testdata", "mappings-9.json"))
	require.Nil(e9)
	j9, e9 := myjson.Unmarshal(fb9)
	if assert.Nil(e9) {
		m9 := &mappingsParser{json: j9}
		p9, e9 := m9.parse()
		if assert.Nil(e9) {
			expected9 := []*Mapping{
				{
					URI:    "/test-for-forwards",
					Method: myhttp.MethodAny,
					Policies: []*Policy{
						{
							When: &When{
								Params: []*NameValuesPair{
									{
										Name:   "no-latency",
										Values: []string{"true"},
									},
								},
							},
							CmdType: mapPolicyForwards,
							Forwards: &Forwards{
								Path: "/",
							},
						},
						{
							CmdType: mapPolicyForwards,
							Forwards: &Forwards{
								Path: "/",
								Latency: &Interval{
									Min: 1000,
									Max: 2000,
								},
							},
						},
					},
				},
			}
			assert.Equal(expected9, p9)
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

	t2 := &templateParser{Parser: Parser{filename: filepath.Join("testdata", "template-2.json")}}
	p2, e2 := t2.parse()
	if assert.Nil(e2) {
		expected2 := &template{
			content: myjson.Array{
				myjson.Object{
					"v": myjson.String("@{v}"),
				},
				myjson.Object{
					"v-v": myjson.String("@{v}-@{v}"),
				},
			},
		}
		assert.Equal(expected2, p2)
	}
}

func TestVarsJSONParser_parse(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	t0 := &varsJSONParser{Parser: Parser{filename: filepath.Join("testdata", "vars-0.json")}}
	_, e0 := t0.parse()
	assert.NotNil(e0)

	t1 := &varsJSONParser{Parser: Parser{filename: filepath.Join("testdata", "vars-1.json")}}
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

func TestVarsCSVParser_parse(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	expected := []*vars{
		{
			table: map[string]interface{}{
				"a": myjson.Number(1),
				"b": myjson.Boolean(true),
				"c": myjson.String("one"),
			},
		},
		{
			table: map[string]interface{}{
				"a": myjson.Number(0),
				"b": myjson.Boolean(false),
				"c": myjson.String("zero"),
			},
		},
	}

	t0 := &varsCSVParser{Parser: Parser{filename: filepath.Join("testdata", "vars-0.csv")}}
	p0, e0 := t0.parse()
	if assert.Nil(e0) {
		assert.Equal(expected, p0)
	}

	t1 := &varsCSVParser{Parser: Parser{filename: filepath.Join("testdata", "vars-1.csv")}}
	p1, e1 := t1.parse()
	if assert.Nil(e1) {
		assert.Equal(expected, p1)
	}
}
