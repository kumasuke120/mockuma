package myjson

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var a = Array([]interface{}{Number(1), String("2")})
var o = Object(map[string]interface{}{
	"k": String("v"),
	"k2": Object{
		"a": nil,
	},
	"k3": a,
	"k4": Number(1),
})

func TestObject_GetObject(t *testing.T) {
	r, e := o.GetObject("k2")
	if assert.Nil(t, e) {
		assert.Equal(t, o["k2"], r)
	}
}

func TestObject_GetArray(t *testing.T) {
	r, e := o.GetArray("k3")
	if assert.Nil(t, e) {
		assert.Equal(t, a, r)
	}
}

func TestObject_GetNumber(t *testing.T) {
	r, e := o.GetNumber("k4")
	if assert.Nil(t, e) {
		assert.Equal(t, Number(1), r)
	}
}

func TestObject_GetString(t *testing.T) {
	r, e := o.GetString("k")
	if assert.Nil(t, e) {
		assert.Equal(t, o["k"], r)
	}
}

func TestObject_Has(t *testing.T) {
	assert.True(t, o.Has("k"))
	assert.False(t, o.Has("n"))
}

func TestObject_Get(t *testing.T) {
	assert.Equal(t, String("v"), o.Get("k"))
	assert.Nil(t, o.Get("n"))
}

func TestObject_Set(t *testing.T) {
	o2 := o.Set("x", "v2")
	assert.Equal(t, String("v2"), o2.Get("x"))
	assert.Nil(t, o.Get("x"))
}

func TestArray_Has(t *testing.T) {
	assert.True(t, a.Has(1))
	assert.False(t, a.Has(2))
}

func TestArray_Get(t *testing.T) {
	assert.Equal(t, Number(1), a.Get(0))
}

func TestArray_Set(t *testing.T) {
	a2 := a.Set(2, Number(3))
	assert.Equal(t, Number(3), a2.Get(2))
	assert.False(t, a.Has(2))
}
