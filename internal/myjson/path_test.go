package myjson

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

	_, e3 := ParsePath("$$.first.second")
	assert.NotNil(e3)
}

func TestObject_SetByPath(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	o1Path := NewPath("first", 2, "third")
	o1, e1 := Object{}.SetByPath(o1Path, String("value"))
	assert.Nil(e1)
	assert.Equal(`map[first:[<nil> <nil> map[third:"value"]]]`, o1.String())
}
