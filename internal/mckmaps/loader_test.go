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

	file, err := LoadFromFile(filepath.Join("testdata", "mappings-0.json"))
	assert.Nil(file)
	assert.NotNil(err)

	wd, err := os.Getwd()
	require.Nil(err)
	assert.True(strings.HasSuffix(wd, "testdata"))

	err = os.Chdir(oldWd)
	require.Nil(err)
}
