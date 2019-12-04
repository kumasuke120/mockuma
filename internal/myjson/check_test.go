package myjson

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAllNumber(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	v1 := []interface{}{Number(1), Number(2)}
	assert.True(IsAllNumber(v1))

	v2 := []interface{}{Number(1), Boolean(false)}
	assert.False(IsAllNumber(v2))
}

func TestIsNumber(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	v1 := Number(1)
	assert.True(IsNumber(v1))

	v2 := String("")
	assert.False(IsNumber(v2))
}
