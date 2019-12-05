package myjson

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//noinspection GoImportUsedAsName
func TestUnmarshal(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fb1, e1 := ioutil.ReadFile(filepath.Join("testdata", "json1.json"))
	require.Nil(e1)
	u1, e1 := Unmarshal(fb1)
	assert.Nil(e1)
	expected1 := Object(map[string]interface{}{
		"str":  String("hello"),
		"num":  Number(1.23),
		"bool": Boolean(false),
		"arr":  Array([]interface{}{Number(1), String("a"), nil}),
		"null": nil,
	})
	assert.Equal(expected1, u1)
}

func TestToMyJson(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	expected1 := Object(map[string]interface{}{
		"str":  String("hello"),
		"num":  Number(1.23),
		"bool": Boolean(false),
		"arr":  Array([]interface{}{Number(1), String("a"), nil}),
		"null": nil,
	})
	from1 := Object(map[string]interface{}{
		"str":  "hello",
		"num":  1.23,
		"bool": false,
		"arr":  []interface{}{Number(1), String("a"), nil},
		"null": nil,
	})
	assert.Equal(expected1, toMyJson(from1))

	expected2 := Object(map[string]interface{}{
		"str":  String("hello"),
		"num":  Number(1.23),
		"bool": Boolean(false),
		"arr":  Array([]interface{}{Number(1), String("a"), nil}),
	})
	from2 := map[string]interface{}{
		"str":  String("hello"),
		"num":  Number(1.23),
		"bool": Boolean(false),
		"arr":  Array([]interface{}{Number(1), String("a"), nil}),
	}
	assert.Equal(expected2, toMyJson(from2))
}
