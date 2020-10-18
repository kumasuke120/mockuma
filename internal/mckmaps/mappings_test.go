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

	fb0, e0 := ioutil.ReadFile(filepath.Join("testdata", "mappings", "mappings-0.json"))
	require.Nil(e0)
	j0, e0 := myjson.Unmarshal(fb0)
	if assert.Nil(e0) {
		m0 := &mappingsParser{json: j0}
		_, e0 := m0.parse()
		assert.NotNil(e0)
	}

	fb1, e1 := ioutil.ReadFile(filepath.Join("testdata", "mappings", "mappings-1.json"))
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
								StatusCode: myhttp.StatusOK,
								Headers: []*NameValuesPair{
									{
										Name:   myhttp.HeaderContentType,
										Values: []string{"application/json; charset=utf-8"},
									},
								},
								Body: []byte(`["a","b","c"]`),
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

	fb2, e2 := ioutil.ReadFile(filepath.Join("testdata", "mappings", "mappings-2.json"))
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
								StatusCode: myhttp.StatusOK,
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

	fb3, e3 := ioutil.ReadFile(filepath.Join("testdata", "mappings", "mappings-3.json"))
	require.Nil(e3)
	j3, e3 := myjson.Unmarshal(fb3)
	if assert.Nil(e3) {
		m3 := &mappingsParser{json: j3}
		_, e3 = m3.parse()
		assert.NotNil(e3)
		assert.NotEmpty(e3.Error())
	}

	fb4, e4 := ioutil.ReadFile(filepath.Join("testdata", "mappings", "mappings-4.json"))
	require.Nil(e4)
	j4, e4 := myjson.Unmarshal(fb4)
	if assert.Nil(e4) {
		m4 := &mappingsParser{json: j4}
		_, e4 = m4.parse()
		assert.NotNil(e4)
		assert.NotEmpty(e4.Error())
	}

	fb5, e5 := ioutil.ReadFile(filepath.Join("testdata", "mappings", "mappings-5.json"))
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
							Returns: &Returns{StatusCode: myhttp.StatusOK},
						},
						{
							When: &When{
								BodyJSON: myjson.NewExtJSONMatcher(myjson.Object{
									"v": myjson.String("v"),
								}),
							},
							CmdType: mapPolicyReturns,
							Returns: &Returns{StatusCode: myhttp.StatusOK},
						},
						{
							When: &When{
								Body: []byte("123"),
							},
							CmdType: mapPolicyReturns,
							Returns: &Returns{StatusCode: myhttp.StatusOK},
						},
						{
							When: &When{
								Body: []byte("true"),
							},
							CmdType: mapPolicyReturns,
							Returns: &Returns{StatusCode: myhttp.StatusOK},
						},
					},
				},
			}
			assert.Equal(expected5, p5)
		}
	}

	fb6, e6 := ioutil.ReadFile(filepath.Join("testdata", "mappings", "mappings-6.json"))
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
							Returns: &Returns{StatusCode: myhttp.StatusOK},
						},
					},
				},
				{
					URI:    "/{0}/{1}/{0}/{2}",
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
									}, {
										Name:   "2",
										Values: []string{""},
									},
								},
							},
							CmdType: mapPolicyReturns,
							Returns: &Returns{StatusCode: myhttp.StatusOK},
						},
					},
				},
			}
			assert.Equal(expected6, p6)
		}
	}

	fb7, e7 := ioutil.ReadFile(filepath.Join("testdata", "mappings", "mappings-7.json"))
	require.Nil(e7)
	j7, e7 := myjson.Unmarshal(fb7)
	if assert.Nil(e7) {
		m7 := &mappingsParser{json: j7}
		p7, e7 := m7.parse()
		if assert.Nil(e7) {
			expected7 := []*Mapping{
				{
					URI:    "/%E8%B7%B3%E8%BD%AC%E6%B5%8B%E8%AF%95",
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

	fb8, e8 := ioutil.ReadFile(filepath.Join("testdata", "mappings", "mappings-8.json"))
	require.Nil(e8)
	j8, e8 := myjson.Unmarshal(fb8)
	if assert.Nil(e7) {
		m8 := &mappingsParser{json: j8}
		_, e8 := m8.parse()
		assert.NotNil(e8)
	}

	fb9, e9 := ioutil.ReadFile(filepath.Join("testdata", "mappings", "mappings-9.json"))
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

	fb10, e10 := ioutil.ReadFile(filepath.Join("testdata", "mappings", "mappings-10.json"))
	require.Nil(e10)
	j10, e10 := myjson.Unmarshal(fb10)
	if assert.Nil(e9) {
		m10 := &mappingsParser{json: j10}
		_, e10 := m10.parse()
		assert.NotNil(e10)
	}
}
