package myjson

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToObject(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	_, e1 := ToObject(1)
	assert.NotNil(e1)
	assert.NotEmpty(e1.Error())

	o2, e2 := ToObject(Object{})
	if assert.Nil(e2) {
		assert.Equal(Object{}, o2)
	}
}

func TestToArray(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	_, e1 := ToArray(1)
	assert.NotNil(e1)
	assert.NotEmpty(e1.Error())

	o2, e2 := ToArray(Array{})
	if assert.Nil(e2) {
		assert.Equal(Array{}, o2)
	}
}

func TestToNumber(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	_, e1 := ToNumber(Object{})
	assert.NotNil(e1)

	n2, e2 := ToNumber(Number(1))
	if assert.Nil(e2) {
		assert.Equal(Number(1), n2)
	}

	n3, e3 := ToNumber(String("1"))
	if assert.Nil(e3) {
		assert.Equal(Number(1), n3)
	}

	_, e4 := toNumber(String("abc"), "n4")
	assert.NotNil(e4)
	assert.Contains(e4.Error(), "'n4'")
}

func TestToString(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	_, e1 := ToString(nil)
	assert.NotNil(e1)

	s2, e2 := ToString(String("s"))
	if assert.Nil(e2) {
		assert.Equal(String("s"), s2)
	}

	s3, e3 := ToString(1)
	if assert.Nil(e3) {
		assert.Equal(String("1"), s3)
	}
}

func TestToBoolean(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	_, e1 := ToBoolean(nil)
	assert.NotNil(e1)

	b2, e2 := toBoolean(Boolean(true), "b")
	if assert.Nil(e2) {
		assert.Equal(Boolean(true), b2)
	}
}
