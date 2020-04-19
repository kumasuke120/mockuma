package server

import (
	"testing"

	"github.com/kumasuke120/mockuma/internal/mckmaps"
	"github.com/stretchr/testify/assert"
)

func Test_newForwardPolicy(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	p := newForwardPolicy("/path")
	assert.Equal("/path", p.Forwards.Path)
	assert.Equal(mckmaps.CmdTypeForwards, p.CmdType)
}
