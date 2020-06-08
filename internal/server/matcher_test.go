package server

import (
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/kumasuke120/mockuma/internal/mckmaps"
	"github.com/kumasuke120/mockuma/internal/myhttp"
	"github.com/kumasuke120/mockuma/internal/myjson"
	"github.com/stretchr/testify/assert"
)

var emptyJSONMatcher = myjson.MakeExtJSONMatcher(myjson.Object{})
var mappings = &mckmaps.MockuMappings{
	Mappings: []*mckmaps.Mapping{
		{
			URI:      "/hello",
			Method:   myhttp.MethodPost,
			Policies: []*mckmaps.Policy{newJSONPolicy(myhttp.StatusOK, "OK")},
		},
		{
			URI:    "/m",
			Method: myhttp.MethodGet,
			Policies: []*mckmaps.Policy{
				{
					When: &mckmaps.When{
						Params: []*mckmaps.NameValuesPair{
							{
								Name:   "p1",
								Values: []string{"v1"},
							},
							{
								Name:   "p2",
								Values: []string{"v2", "v1"},
							},
						},
					},
					CmdType: mckmaps.CmdTypeForwards,
					Forwards: &mckmaps.Forwards{
						Path: "http://localhost:8080",
					},
				},
				{
					When: &mckmaps.When{
						ParamRegexps: []*mckmaps.NameRegexpPair{
							{
								Name:   "r1",
								Regexp: regexp.MustCompile("^\\d{3}$"),
							},
						},
					},
					CmdType: mckmaps.CmdTypeReturns,
					Returns: &mckmaps.Returns{
						StatusCode: myhttp.StatusOK,
						Body:       []byte(""),
					},
				},
				{
					When: &mckmaps.When{
						ParamJSONs: []*mckmaps.NameJSONPair{
							{
								Name: "j",
								JSON: myjson.MakeExtJSONMatcher(myjson.Object{}),
							},
						},
					},
				},
				{
					When: &mckmaps.When{
						Headers: []*mckmaps.NameValuesPair{
							{
								Name:   "X-H1",
								Values: []string{"v1"},
							},
							{
								Name:   "X-H2",
								Values: []string{"v2", "v1"},
							},
						},
					},
				},
				{
					When: &mckmaps.When{
						HeaderRegexps: []*mckmaps.NameRegexpPair{
							{
								Name:   "X-R1",
								Regexp: regexp.MustCompile("^\\d{3}$"),
							},
						},
					},
				},
				{
					When: &mckmaps.When{
						HeaderJSONs: []*mckmaps.NameJSONPair{
							{
								Name: "X-J1",
								JSON: myjson.MakeExtJSONMatcher(myjson.Object{}),
							},
						},
					},
				},
			},
		},
		{
			URI:    "/m",
			Method: myhttp.MethodPost,
			Policies: []*mckmaps.Policy{
				{
					When: &mckmaps.When{
						Body: []byte("123"),
					},
				},
				{
					When: &mckmaps.When{
						BodyRegexp: regexp.MustCompile("^\\d{3}$"),
					},
					CmdType: mckmaps.CmdTypeReturns,
					Returns: &mckmaps.Returns{
						StatusCode: myhttp.StatusOK,
						Headers: []*mckmaps.NameValuesPair{
							{
								Name:   "Server",
								Values: []string{"TEST/v1"},
							},
						},
					},
				},
				{
					When: &mckmaps.When{
						BodyJSON: &emptyJSONMatcher,
					},
				},
			},
		},
		{
			URI:    "/p/{0}/m{1}",
			Method: myhttp.MethodPut,
			Policies: []*mckmaps.Policy{
				{
					When: &mckmaps.When{
						PathVars: []*mckmaps.NameValuesPair{
							{
								Name:   "0",
								Values: []string{"v0"},
							},
						},
						PathVarRegexps: []*mckmaps.NameRegexpPair{
							{
								Name:   "1",
								Regexp: regexp.MustCompile("^\\d+$"),
							},
						},
					},
				},
			},
		},
	},
	Config: &mckmaps.Config{
		CORS: &mckmaps.CORSOptions{
			Enabled: false,
		},
	},
}

var mappingsWithCORS = &mckmaps.MockuMappings{
	Mappings: mappings.Mappings,
	Config: &mckmaps.Config{
		CORS: &mckmaps.CORSOptions{
			Enabled:          true,
			AllowCredentials: true,
			MaxAge:           1800,
			AllowedOrigins:   []string{"*"},
			AllowedMethods: []myhttp.HTTPMethod{
				myhttp.MethodGet,
				myhttp.MethodPost,
				myhttp.MethodHead,
				myhttp.MethodOptions,
			},
			AllowedHeaders: []string{
				myhttp.HeaderOrigin,
				myhttp.HeaderAccept,
				myhttp.HeaderXRequestWith,
				myhttp.HeaderContentType,
				myhttp.HeaderAccessControlRequestMethod,
				myhttp.HeaderAccessControlRequestHeaders,
			},
			ExposedHeaders: nil,
		},
		MatchTrailingSlash: false,
	},
}

func TestNewPathMatcher(t *testing.T) {
	matcher := newPathMatcher(mappings)

	expectedDirectPath := map[string][]*mckmaps.Mapping{
		"/hello": {mappings.Mappings[0]},
		"/m": {
			&mckmaps.Mapping{
				URI:      "/m",
				Method:   myhttp.MethodGet,
				Policies: mappings.Mappings[1].Policies,
			},
			mappings.Mappings[2],
		},
	}
	assert.Equal(t, expectedDirectPath, matcher.directPath)
	expectedPatternPath := map[*regexp.Regexp][]*mckmaps.Mapping{
		regexp.MustCompile("^/p/(?P<v0>.*?)/m(?P<v1>.*?)$"): {
			mappings.Mappings[3],
		},
	}
	assert.Equal(t, formatRegexpKeyMap(expectedPatternPath), formatRegexpKeyMap(matcher.patternPath))
}

func formatRegexpKeyMap(m map[*regexp.Regexp][]*mckmaps.Mapping) map[string][]*mckmaps.Mapping {
	result := make(map[string][]*mckmaps.Mapping, len(m))
	for r, ms := range m {
		result[r.String()] = ms
	}
	return result
}

func TestPathMatcher_matches(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	matcher := newPathMatcher(mappings)
	bound := matcher.bind(httptest.NewRequest("POST", "/hello", nil))
	assert.True(bound.matches())
	assert.Equal(matchState(matchExact), bound.matchState)
	assert.Equal(mappings.Mappings[0], bound.matchedMapping)
}

func TestPathMatcher_matchPolicy(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	matcher := newPathMatcher(mappings)

	bound1 := matcher.bind(httptest.NewRequest("", "/m?p1=v1&p2=v1&p2=v2", nil))
	assert.True(bound1.matches())
	assert.Equal(matchState(matchExact), bound1.matchState)
	assert.Equal(mappings.Mappings[1].Policies[0], bound1.matchPolicy())

	bound2 := matcher.bind(httptest.NewRequest("", "/m?r1=123", nil))
	assert.True(bound2.matches())
	assert.Equal(matchState(matchExact), bound2.matchState)
	assert.Equal(mappings.Mappings[1].Policies[1], bound2.matchPolicy())
	bound2 = matcher.bind(httptest.NewRequest("", "/m?r1=12", nil))
	assert.True(bound2.matches())
	assert.Equal(matchState(matchExact), bound2.matchState)
	assert.Equal(pNoPolicyMatched, bound2.matchPolicy())

	bound3 := matcher.bind(httptest.NewRequest("", "/m?j={}", nil))
	assert.True(bound3.matches())
	assert.Equal(matchState(matchExact), bound3.matchState)
	assert.Equal(mappings.Mappings[1].Policies[2], bound3.matchPolicy())
	bound3 = matcher.bind(httptest.NewRequest("", "/m?j=120", nil))
	assert.True(bound3.matches())
	assert.Equal(matchState(matchExact), bound3.matchState)
	assert.Equal(pNoPolicyMatched, bound3.matchPolicy())

	req4 := httptest.NewRequest("", "/m", nil)
	req4.Header.Add("X-H1", "v1")
	req4.Header.Add("X-H2", "v1")
	req4.Header.Add("X-H2", "v2")
	bound4 := matcher.bind(req4)
	assert.True(bound4.matches())
	assert.Equal(matchState(matchExact), bound4.matchState)
	assert.Equal(mappings.Mappings[1].Policies[3], bound4.matchPolicy())

	req5p1 := httptest.NewRequest("", "/m", nil)
	req5p1.Header.Add("X-R1", "123")
	bound5 := matcher.bind(req5p1)
	assert.True(bound5.matches())
	assert.Equal(matchState(matchExact), bound5.matchState)
	assert.Equal(mappings.Mappings[1].Policies[4], bound5.matchPolicy())
	req5p2 := httptest.NewRequest("", "/m", nil)
	req5p2.Header.Add("X-R1", "12")
	bound5 = matcher.bind(req5p2)
	assert.True(bound5.matches())
	assert.Equal(matchState(matchExact), bound5.matchState)
	assert.Equal(pNoPolicyMatched, bound5.matchPolicy())

	req6p1 := httptest.NewRequest("", "/m", nil)
	req6p1.Header.Add("X-J1", "{}")
	bound6 := matcher.bind(req6p1)
	assert.True(bound6.matches())
	assert.Equal(matchState(matchExact), bound6.matchState)
	assert.Equal(mappings.Mappings[1].Policies[5], bound6.matchPolicy())
	req6p2 := httptest.NewRequest("", "/m", nil)
	req6p2.Header.Add("X-J1", "120")
	bound6 = matcher.bind(req6p2)
	assert.True(bound6.matches())
	assert.Equal(matchState(matchExact), bound6.matchState)
	assert.Equal(pNoPolicyMatched, bound6.matchPolicy())

	bound7 := matcher.bind(httptest.NewRequest("POST", "/m", strings.NewReader("123")))
	assert.True(bound7.matches())
	assert.Equal(matchState(matchExact), bound7.matchState)
	assert.Equal(mappings.Mappings[2].Policies[0], bound7.matchPolicy())

	bound8 := matcher.bind(httptest.NewRequest("POST", "/m", strings.NewReader("120")))
	assert.True(bound8.matches())
	assert.Equal(matchState(matchExact), bound8.matchState)
	assert.Equal(mappings.Mappings[2].Policies[1], bound8.matchPolicy())
	bound8 = matcher.bind(httptest.NewRequest("POST", "/m", strings.NewReader("95")))
	assert.True(bound8.matches())
	assert.Equal(matchState(matchExact), bound8.matchState)
	assert.Equal(pNoPolicyMatched, bound8.matchPolicy())

	bound9 := matcher.bind(httptest.NewRequest("POST", "/m", strings.NewReader("{}")))
	assert.True(bound9.matches())
	assert.Equal(matchState(matchExact), bound9.matchState)
	assert.Equal(mappings.Mappings[2].Policies[2], bound9.matchPolicy())

	bound10 := matcher.bind(httptest.NewRequest("", "/hello", nil))
	assert.True(bound10.matches())
	assert.Equal(matchState(matchURI), bound10.matchState)

	bound11 := matcher.bind(httptest.NewRequest("PUT", "/p/v0/m1", nil))
	assert.True(bound11.matches())
	assert.Equal(matchState(matchExact), bound11.matchState)
	assert.Equal(mappings.Mappings[3].Policies[0], bound11.matchPolicy())

	bound12 := matcher.bind(httptest.NewRequest("PUT", "/p/v0/ma", nil))
	assert.True(bound12.matches())
	assert.Equal(matchState(matchExact), bound12.matchState)
	assert.Equal(pNoPolicyMatched, bound12.matchPolicy())

	bound13 := matcher.bind(httptest.NewRequest("", "/p/v0/m1", nil))
	assert.True(bound13.matches())
	assert.Equal(matchState(matchURI), bound13.matchState)

	bound14 := matcher.bind(httptest.NewRequest("PUT", "/p/v1/m1", nil))
	assert.True(bound14.matches())
	assert.Equal(matchState(matchExact), bound14.matchState)
	assert.Equal(pNoPolicyMatched, bound14.matchPolicy())
}
