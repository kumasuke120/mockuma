package myjson

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//noinspection GoImportUsedAsName
func TestMarshal(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fb1, e1 := ioutil.ReadFile(filepath.Join("testdata", "json1-slim.json"))
	require.Nil(e1)
	j1 := Object(map[string]interface{}{
		"str":  String("hello"),
		"num":  Number(1.23),
		"bool": Boolean(false),
		"arr":  Array([]interface{}{Number(1), String("a"), nil}),
		"null": nil,
	})
	m1, e1 := Marshal(j1)
	assert.Nil(e1)
	assert.Equal(fb1, m1)
}
