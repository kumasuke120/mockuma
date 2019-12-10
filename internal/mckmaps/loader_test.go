package mckmaps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//noinspection GoImportUsedAsName
func TestLoadFromFile(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	oldWd, err := os.Getwd()
	require.Nil(err)

	file1, err1 := LoadFromFile(filepath.Join("testdata", "mappings-0.json"))
	assert.Nil(file1)
	assert.NotNil(err1)
	wd1, err1 := os.Getwd()
	require.Nil(err)
	assert.True(strings.HasSuffix(wd1, "testdata"))

	file2, err2 := LoadFromFile("")
	require.Nil(err2)
	assert.NotNil(file2)

	err = os.Chdir(oldWd)
	require.Nil(err)
}
