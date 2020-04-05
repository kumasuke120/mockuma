package mckmaps

import (
	"path/filepath"
	"testing"

	"github.com/kumasuke120/mockuma/internal/myjson"
	"github.com/kumasuke120/mockuma/internal/myos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//noinspection GoImportUsedAsName
func TestTemplateParser_parse(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp0 := filepath.Join("testdata", "template", "template-0.json")
	t0 := &templateParser{Parser: Parser{filename: fp0}}
	_, e0 := t0.parse()
	assert.NotNil(e0)

	fp1 := filepath.Join("testdata", "template", "template-1.json")
	t1 := &templateParser{Parser: Parser{filename: fp1}}
	p1, e1 := t1.parse()
	if assert.Nil(e1) {
		expected1 := &template{
			content: myjson.Object{
				"v":   myjson.String("@{v}"),
				"v-v": myjson.String("@{v}-@{v}"),
			},
			filename: fp1,
		}
		assert.Equal(expected1, p1)
	}

	fp2 := filepath.Join("testdata", "template", "template-2.json")
	t2 := &templateParser{Parser: Parser{filename: fp2}}
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
			filename: fp2,
		}
		assert.Equal(expected2, p2)
	}

	require.Nil(myos.InitWd())
	oldWd := myos.GetWd()
	require.Nil(myos.Chdir(filepath.Join(oldWd, "testdata", "template")))

	fp3 := "template-3.json"
	t3 := &templateParser{Parser: Parser{filename: fp3}}
	p3, e3 := t3.parse()
	if assert.Nil(e3) {
		expected3 := &template{
			content: myjson.Array{
				myjson.Object{
					"v": myjson.String("@{v}"),
				},
				myjson.Object{
					"v-v": myjson.String("@{v}-@{v}"),
				},
			},
			filename: fp3,
		}
		assert.Equal(expected3, p3)
	}

	fp4 := "template-cyclic-a.json"
	t4 := &templateParser{Parser: Parser{filename: fp4}}
	_, e4 := t4.parse()
	assert.NotNil(e4)
	assert.Contains(e4.Error(), "cyclic")

	require.Nil(myos.Chdir(oldWd))
}

var theTemplate = &template{filename: "test"}
var theVars = &vars{table: map[string]interface{}{
	"a":  myjson.String("123"),
	"b1": myjson.Number(123),
	"b2": myjson.Number(123.123),
	"c":  myjson.Boolean(true),
}}

func TestTemplate_renderString(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	t1 := myjson.String("a = \"@{a}\", b1 = @{b1}, b2 = @{b2}, c = @{c}")
	jp1 := myjson.NewPath()
	r1, e1 := theTemplate.renderString(jp1, t1, theVars)
	if assert.Nil(e1) {
		assert.IsType(myjson.String(""), r1)
		assert.EqualValues("a = \"123\", b1 = 123, b2 = 123.123, c = true", r1)
	}

	t2 := myjson.String("a = \"@{a:%8s}\", b1 = ORD@{b1:%010d}, b2 = @{b2:%.2f}, c = @{c}")
	jp2 := myjson.NewPath()
	r2, e2 := theTemplate.renderString(jp2, t2, theVars)
	if assert.Nil(e2) {
		assert.IsType(myjson.String(""), r2)
		assert.EqualValues("a = \"     123\", b1 = ORD0000000123, b2 = 123.12, c = true", r2)
	}

	t3 := myjson.String("a = @@{a}")
	jp3 := myjson.NewPath()
	r3, e3 := theTemplate.renderString(jp3, t3, theVars)
	if assert.Nil(e3) {
		assert.IsType(myjson.String(""), r3)
		assert.EqualValues("a = @{a}", r3)
	}

	t4 := myjson.String("a = @{}")
	jp4 := myjson.NewPath()
	_, e4 := theTemplate.renderString(jp4, t4, theVars)
	if assert.NotNil(e4) {
		assert.NotEmpty(e4.Error())
	}
}

func TestTemplate_renderObject(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	t1 := myjson.Object{
		"a": myjson.String("@{a}"),
		"b": myjson.Object{
			"b1": myjson.String("@{b1}"),
			"b2": myjson.String("@{b2}"),
		},
		"c": myjson.String("@{c}"),
	}
	jp1 := myjson.NewPath()
	r1, e1 := theTemplate.renderObject(jp1, t1, theVars)
	if assert.Nil(e1) {
		expected1 := myjson.Object{
			"a": myjson.String("123"),
			"b": myjson.Object{
				"b1": myjson.Number(123),
				"b2": myjson.Number(123.123),
			},
			"c": myjson.Boolean(true),
		}
		assert.Equal(expected1, r1)
	}
}

func TestTestTemplate_renderArray(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	t1 := myjson.Array{
		myjson.String("@{a}"),
		myjson.Array{
			myjson.String("@{b1}"),
			myjson.String("@{b2}"),
		},
		myjson.String("@{c}"),
	}
	jp1 := myjson.NewPath()
	r1, e1 := theTemplate.renderArray(jp1, t1, theVars)
	if assert.Nil(e1) {
		expected1 := myjson.Array{
			myjson.String("123"),
			myjson.Array{
				myjson.Number(123),
				myjson.Number(123.123),
			},
			myjson.Boolean(true),
		}
		assert.Equal(expected1, r1)
	}
}
