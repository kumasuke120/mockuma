package myhttp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToHttpMethod(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	assert.Equal(Options, ToHttpMethod("Options"))
	assert.Equal(Get, ToHttpMethod("get"))
	assert.Equal(Head, ToHttpMethod("HEAD"))
	assert.Equal(Post, ToHttpMethod("POST"))
	assert.Equal(Put, ToHttpMethod("put"))
	assert.Equal(Delete, ToHttpMethod("dElEtE"))
	assert.Equal(Trace, ToHttpMethod("trAcE"))
	assert.Equal(Connect, ToHttpMethod("CONNECT"))
	assert.Equal(Patch, ToHttpMethod("PAtch"))
	assert.Equal(Any, ToHttpMethod("@any"))
}

func TestHttpMethod_Matches(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	assert.True(Get.Matches("GET"))
	assert.True(Any.Matches("POST"))
	assert.False(Post.Matches("GET"))
}

func TestHttpMethod_String(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	assert.Equal("GET", Get.String())
	assert.Equal("OPTIONS", Options.String())
}
