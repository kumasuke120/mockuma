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

var emptyJsonMatcher = myjson.MakeExtJsonMatcher(myjson.Object{})
var mappings = &mckmaps.MockuMappings{
	Mappings: []*mckmaps.Mapping{
		{
			Uri:      "/hello",
			Method:   myhttp.Post,
			Policies: []*mckmaps.Policy{newStatusJsonPolicy(myhttp.Ok, "OK")},
		},
		{
			Uri:    "/m",
			Method: myhttp.Get,
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
				},
				{
					When: &mckmaps.When{
						ParamJsons: []*mckmaps.NameJsonPair{
							{
								Name: "j",
								Json: myjson.MakeExtJsonMatcher(myjson.Object{}),
							},
						},
					},
				},
			},
		},
		{
			Uri:    "/m",
			Method: myhttp.Get,
			Policies: []*mckmaps.Policy{
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
						HeaderJsons: []*mckmaps.NameJsonPair{
							{
								Name: "X-J1",
								Json: myjson.MakeExtJsonMatcher(myjson.Object{}),
							},
						},
					},
				},
			},
		},
		{
			Uri:    "/m",
			Method: myhttp.Post,
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
				},
				{
					When: &mckmaps.When{
						BodyJson: &emptyJsonMatcher,
					},
				},
			},
		},
	},
}

func TestNewPathMatcher(t *testing.T) {
	matcher := newPathMatcher(mappings)

	expected := map[string][]*mckmaps.Mapping{
		"/hello": {mappings.Mappings[0]},
		"/m": {
			&mckmaps.Mapping{
				Uri:    "/m",
				Method: myhttp.Get,
				Policies: []*mckmaps.Policy{
					mappings.Mappings[1].Policies[0],
					mappings.Mappings[1].Policies[1],
					mappings.Mappings[1].Policies[2],
					mappings.Mappings[2].Policies[0],
					mappings.Mappings[2].Policies[1],
					mappings.Mappings[2].Policies[2],
				},
			},
			mappings.Mappings[3],
		},
	}
	assert.Equal(t, expected, matcher.uri2mappings)
}

func TestPathMatcher_matches(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	matcher := newPathMatcher(mappings)
	bound := matcher.bind(httptest.NewRequest("POST", "/hello", nil))
	assert.True(bound.matches())
	assert.Equal(mappings.Mappings[0], bound.matchedMapping)
}

func TestPathMatcher_matchPolicy(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	matcher := newPathMatcher(mappings)

	bound1 := matcher.bind(httptest.NewRequest("", "/m?p1=v1&p2=v1&p2=v2", nil))
	assert.True(bound1.matches())
	assert.Equal(mappings.Mappings[1].Policies[0], bound1.matchPolicy())

	bound2 := matcher.bind(httptest.NewRequest("", "/m?r1=123", nil))
	assert.True(bound2.matches())
	assert.Equal(mappings.Mappings[1].Policies[1], bound2.matchPolicy())
	bound2 = matcher.bind(httptest.NewRequest("", "/m?r1=12", nil))
	assert.True(bound2.matches())
	assert.Nil(bound2.matchPolicy())

	bound3 := matcher.bind(httptest.NewRequest("", "/m?j={}", nil))
	assert.True(bound3.matches())
	assert.Equal(mappings.Mappings[1].Policies[2], bound3.matchPolicy())
	bound3 = matcher.bind(httptest.NewRequest("", "/m?j=120", nil))
	assert.True(bound3.matches())
	assert.Nil(bound3.matchPolicy())

	req4 := httptest.NewRequest("", "/m", nil)
	req4.Header.Add("X-H1", "v1")
	req4.Header.Add("X-H2", "v1")
	req4.Header.Add("X-H2", "v2")
	bound4 := matcher.bind(req4)
	assert.True(bound4.matches())
	assert.Equal(mappings.Mappings[2].Policies[0], bound4.matchPolicy())

	req5p1 := httptest.NewRequest("", "/m", nil)
	req5p1.Header.Add("X-R1", "123")
	bound5 := matcher.bind(req5p1)
	assert.True(bound5.matches())
	assert.Equal(mappings.Mappings[2].Policies[1], bound5.matchPolicy())
	req5p2 := httptest.NewRequest("", "/m", nil)
	req5p2.Header.Add("X-R1", "12")
	bound5 = matcher.bind(req5p2)
	assert.True(bound5.matches())
	assert.Nil(bound5.matchPolicy())

	req6p1 := httptest.NewRequest("", "/m", nil)
	req6p1.Header.Add("X-J1", "{}")
	bound6 := matcher.bind(req6p1)
	assert.True(bound6.matches())
	assert.Equal(mappings.Mappings[2].Policies[2], bound6.matchPolicy())
	req6p2 := httptest.NewRequest("", "/m", nil)
	req6p2.Header.Add("X-J1", "120")
	bound6 = matcher.bind(req6p2)
	assert.True(bound6.matches())
	assert.Nil(bound6.matchPolicy())

	bound7 := matcher.bind(httptest.NewRequest("POST", "/m", strings.NewReader("123")))
	assert.True(bound7.matches())
	assert.Equal(mappings.Mappings[3].Policies[0], bound7.matchPolicy())

	bound8 := matcher.bind(httptest.NewRequest("POST", "/m", strings.NewReader("120")))
	assert.True(bound8.matches())
	assert.Equal(mappings.Mappings[3].Policies[1], bound8.matchPolicy())
	bound8 = matcher.bind(httptest.NewRequest("POST", "/m", strings.NewReader("95")))
	assert.True(bound8.matches())
	assert.Nil(bound8.matchPolicy())

	bound9 := matcher.bind(httptest.NewRequest("POST", "/m", strings.NewReader("{}")))
	assert.True(bound9.matches())
	assert.Equal(mappings.Mappings[3].Policies[2], bound9.matchPolicy())
}
