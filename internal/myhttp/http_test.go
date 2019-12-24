package myhttp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToHTTPMethod(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	assert.Equal(Options, ToHTTPMethod("Options"))
	assert.Equal(Get, ToHTTPMethod("get"))
	assert.Equal(Head, ToHTTPMethod("HEAD"))
	assert.Equal(Post, ToHTTPMethod("POST"))
	assert.Equal(Put, ToHTTPMethod("put"))
	assert.Equal(Delete, ToHTTPMethod("dElEtE"))
	assert.Equal(Trace, ToHTTPMethod("trAcE"))
	assert.Equal(Connect, ToHTTPMethod("CONNECT"))
	assert.Equal(Patch, ToHTTPMethod("PAtch"))
	assert.Equal(Any, ToHTTPMethod("@any"))
}

func TestHTTPMethod_Matches(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	assert.True(Get.Matches("GET"))
	assert.True(Any.Matches("POST"))
	assert.False(Post.Matches("GET"))
}

func TestHTTPMethod_String(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	assert.Equal("GET", Get.String())
	assert.Equal("OPTIONS", Options.String())
}
