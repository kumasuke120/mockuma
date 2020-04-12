package myhttp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToHTTPMethod(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	assert.Equal(MethodOptions, ToHTTPMethod("Options"))
	assert.Equal(MethodGet, ToHTTPMethod("get"))
	assert.Equal(MethodHead, ToHTTPMethod("HEAD"))
	assert.Equal(MethodPost, ToHTTPMethod("POST"))
	assert.Equal(MethodPut, ToHTTPMethod("put"))
	assert.Equal(MethodDelete, ToHTTPMethod("dElEtE"))
	assert.Equal(MethodTrace, ToHTTPMethod("trAcE"))
	assert.Equal(MethodConnect, ToHTTPMethod("CONNECT"))
	assert.Equal(MethodPatch, ToHTTPMethod("PAtch"))
	assert.Equal(MethodAny, ToHTTPMethod("*"))
	assert.Equal(HTTPMethod("RESET"), ToHTTPMethod("RESet"))
}

func TestHTTPMethod_Matches(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	assert.True(MethodGet.Matches(MethodGet))
	assert.True(MethodAny.Matches(MethodPost))
	assert.False(MethodPost.Matches(MethodGet))
}

func TestHTTPMethod_String(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	assert.Equal("GET", MethodGet.String())
	assert.Equal("OPTIONS", MethodOptions.String())
}
