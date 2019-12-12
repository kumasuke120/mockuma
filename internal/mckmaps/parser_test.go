package mckmaps

import (
	"io/ioutil"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/kumasuke120/mockuma/internal/myhttp"
	"github.com/kumasuke120/mockuma/internal/myjson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
					Uri:    "/",
					Method: myhttp.Get,
					Policies: []*Policy{
						{
							Returns: &Returns{
								StatusCode: myhttp.Ok,
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
					Uri:    "/",
					Method: myhttp.Get,
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
								HeaderJsons: []*NameJsonPair{
									{
										Name: "X-J",
										Json: myjson.MakeExtJsonMatcher(myjson.Object(map[string]interface{}{
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
								ParamJsons: []*NameJsonPair{
									{
										Name: "j",
										Json: myjson.MakeExtJsonMatcher(myjson.Object(map[string]interface{}{
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
									Max: 3000,
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
		_, e3 := m3.parse()
		assert.NotNil(e3)
		assert.NotEmpty(e3.Error())
	}
}

func TestTemplateParser_parse(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	t0 := &templateParser{parser: parser{filename: filepath.Join("testdata", "template-0.json")}}
	_, e0 := t0.parse()
	assert.NotNil(e0)

	t1 := &templateParser{parser: parser{filename: filepath.Join("testdata", "template-1.json")}}
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

	t0 := &varsParser{parser: parser{filename: filepath.Join("testdata", "vars-0.json")}}
	_, e0 := t0.parse()
	assert.NotNil(e0)

	t1 := &varsParser{parser: parser{filename: filepath.Join("testdata", "vars-1.json")}}
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
