package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToString(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	assert.Equal("a", ToString("a"))

	var sb strings.Builder
	sb.WriteString("abc")
	assert.Equal("abc", ToString(&sb))

	assert.Equal("1", ToString(1))
}
