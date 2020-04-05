package loader

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/kumasuke120/mockuma/internal/myos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//noinspection GoImportUsedAsName
func TestLoader_Load(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	require.Nil(myos.InitWd())
	oldWd := myos.GetWd()

	file1, err1 := New(filepath.Join("testdata", "mappings-0.json")).Load()
	assert.Nil(file1)
	assert.NotNil(err1)
	wd1 := myos.GetWd()
	assert.True(strings.HasSuffix(wd1, "testdata"))

	file2, err2 := New("").Load()
	require.Nil(err2)
	assert.NotNil(file2)

	require.Nil(myos.Chdir(oldWd))
}
