package myjson

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var o = Object(map[string]interface{}{
	"k": String("v"),
})

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
