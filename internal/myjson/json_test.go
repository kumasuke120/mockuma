package myjson

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var o = Object(map[string]interface{}{
	"k": String("v"),
})
var a = Array([]interface{}{Number(1), String("2")})

func TestObject_Has(t *testing.T) {
	assert.True(t, o.Has("k"))
	assert.False(t, o.Has("n"))
}

func TestObject_Get(t *testing.T) {
	assert.Equal(t, String("v"), o.Get("k"))
	assert.Nil(t, o.Get("n"))
}

func TestObject_Set(t *testing.T) {
	o2 := o.Set("k2", "v2")
	assert.Equal(t, String("v2"), o2.Get("k2"))
	assert.Nil(t, o.Get("k2"))
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