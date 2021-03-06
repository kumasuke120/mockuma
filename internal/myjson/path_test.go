package myjson

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPath(t *testing.T) {
	assert.Panics(t, func() {
		NewPath([]string{"a", "b"})
	})
}

func TestParsePath(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	p1Str := "$[0].name['@type']"
	p1, e1 := ParsePath(p1Str)
	assert.Nil(e1)
	assert.Equal(p1Str, p1.String())

	p2, e2 := ParsePath("$['first'][1].second")
	assert.Nil(e2)
	assert.Equal("$.first[1].second", p2.String())

	ps3 := "$$.first.second"
	_, e3 := ParsePath(ps3)
	assert.NotNil(e3)
	assert.Contains(e3.Error(), ps3)

	_, e4 := ParsePath("$.first[1")
	assert.NotNil(e4)

	_, e5 := ParsePath(" $")
	assert.NotNil(e5)

	_, e6 := ParsePath("")
	assert.NotNil(e6)

	_, e7 := ParsePath("$.!!")
	assert.NotNil(e7)

	_, e8 := ParsePath("$[.]")
	assert.NotNil(e8)

	_, e9 := ParsePath("$['abc'[")
	assert.NotNil(e9)

	_, e10 := ParsePath("$[1abc]")
	assert.NotNil(e10)
}

func TestObject_SetByPath(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	p := NewPath("first", 2, "third")
	o1, e1 := Object{}.SetByPath(p, String("value"))
	if assert.Nil(e1) {
		assert.Equal(`map[first:[<nil> <nil> map[third:"value"]]]`, o1.String())
	}

	o2 := Object(map[string]interface{}{
		"first": Array{},
	})
	o2, e2 := o2.SetByPath(p, "value")
	if assert.Nil(e2) {
		assert.Equal(`map[first:[<nil> <nil> map[third:"value"]]]`, o2.String())
	}

	o3 := Object{}
	_, e3 := o3.SetByPath(NewPath(), "")
	assert.NotNil(e3)
}

func TestPath_Append(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	p := NewPath()
	if assert.NotNil(p) {
		p.Append(1)
		assert.Equal([]interface{}{1}, p.paths)

		p.Append("s")
		assert.Equal([]interface{}{1, "s"}, p.paths)
	}
}

func TestPath_SetLast(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	p1 := NewPath()
	p1.SetLast(1)
	assert.Nil(p1.paths)

	p2 := NewPath(0)
	p2.SetLast(1)
	assert.Equal([]interface{}{1}, p2.paths)

	p3 := NewPath("")
	p3.SetLast("s")
	assert.Equal([]interface{}{"s"}, p3.paths)
}

func TestPath_RemoveLast(t *testing.T) {
	p := NewPath(0)
	p.RemoveLast()
	assert.Empty(t, p.paths)
}

func TestPath_String(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	assert.Equal("$[0].abc", NewPath(0, "abc").String())
	assert.Equal("$.abc[0]", NewPath("abc", 0).String())
	assert.Equal("$['@abc'][0]", NewPath("@abc", 0).String())
	assert.Equal("$['\\'@abc\\''][0]", NewPath("'@abc'", 0).String())
}
