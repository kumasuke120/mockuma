package server

import (
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/kumasuke120/mockuma/internal/mckmaps"
	"github.com/kumasuke120/mockuma/internal/myhttp"
	"github.com/kumasuke120/mockuma/internal/myjson"
	"github.com/stretchr/testify/assert"
)

var mappings = &mckmaps.MockuMappings{
	Mappings: []*mckmaps.Mapping{
		{
			Uri:      "/hello",
			Method:   myhttp.Post,
			Policies: []*mckmaps.Policy{newStatusJsonPolicy(myhttp.Ok, "OK")},
		},
		{
			Uri:    "/mp1",
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
	},
}

func TestNewPathMatcher(t *testing.T) {
	matcher := newPathMatcher(mappings)
	expected := map[string][]*mckmaps.Mapping{
		"/hello": {mappings.Mappings[0]},
		"/mp1":   {mappings.Mappings[1]},
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

	bound1 := matcher.bind(httptest.NewRequest("", "/mp1?p1=v1&p2=v1&p2=v2", nil))
	assert.True(bound1.matches())
	assert.Equal(mappings.Mappings[1].Policies[0], bound1.matchPolicy())

	bound2 := matcher.bind(httptest.NewRequest("", "/mp1?r1=123", nil))
	assert.True(bound2.matches())
	assert.Equal(mappings.Mappings[1].Policies[1], bound2.matchPolicy())
	bound2 = matcher.bind(httptest.NewRequest("", "/mp1?r1=12", nil))
	assert.True(bound2.matches())
	assert.Nil(bound2.matchPolicy())

	bound3 := matcher.bind(httptest.NewRequest("", "/mp1?j={}", nil))
	assert.True(bound3.matches())
	assert.Equal(mappings.Mappings[1].Policies[2], bound3.matchPolicy())
	bound3 = matcher.bind(httptest.NewRequest("", "/mp1?j=123", nil))
	assert.True(bound3.matches())
	assert.Nil(bound3.matchPolicy())
}
