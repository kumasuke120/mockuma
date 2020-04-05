package myjson

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	fb1 := []byte(`{"arr":[1,"a",null],"bool":false,"null":null,"num":1.23,"str":"hello"}`)
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
