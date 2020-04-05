package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//noinspection GoImportUsedAsName
func TestUnzip(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fn0 := filepath.Join("testdata", "template.zip")
	dir, err := unzip(fn0)
	if assert.Nil(err) {
		var files0 []string
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			require.Nil(err)

			rel, err := filepath.Rel(dir, path)
			require.Nil(err)
			files0 = append(files0, rel)
			return nil
		})
		if assert.Nil(err) {
			assert.Len(files0, 6)
			assert.Contains(files0, ".")
			assert.Contains(files0, "main.json")
			assert.Contains(files0, "t")
			assert.Contains(files0, filepath.Join("t", "a.template.json"))
			assert.Contains(files0, filepath.Join("t", "b.template.json"))
			assert.Contains(files0, "test.mappings.json")
		}
	}

	require.Nil(os.RemoveAll(dir))
}
