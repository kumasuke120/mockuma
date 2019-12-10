package mckmaps

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/kumasuke120/mockuma/internal/myjson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//noinspection GoImportUsedAsName
func TestCommentFilter(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fb0, err := ioutil.ReadFile(filepath.Join("testdata", "removeComment-0.json"))
	require.Nil(err)
	j0, err := myjson.Unmarshal(fb0)
	require.Nil(err)

	fbe, err := ioutil.ReadFile(filepath.Join("testdata", "removeComment-expected.json"))
	require.Nil(err)
	je, err := myjson.Unmarshal(fbe)
	require.Nil(err)

	ja, err := doFiltersOnV(j0, &commentFilter{})
	assert.Nil(err)
	assert.Equal(je, ja)
}

//noinspection GoImportUsedAsName
func TestLoadFileFilter(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	oldWd, err := os.Getwd()
	require.Nil(err)

	fb0, err := ioutil.ReadFile(filepath.Join("testdata", "loadFile-0.json"))
	require.Nil(err)
	j0, err := myjson.Unmarshal(fb0)
	require.Nil(err)

	fbe, err := ioutil.ReadFile(filepath.Join("testdata", "loadFile-expected.json"))
	require.Nil(err)
	je, err := myjson.Unmarshal(fbe)
	require.Nil(err)

	err = os.Chdir(filepath.Join(oldWd, "testdata"))
	require.Nil(err)

	ja, err := doFiltersOnV(j0, makeLoadFileFilter())
	if assert.Nil(err) {
		assert.Equal(je, ja)
	}

	err = os.Chdir(oldWd)
	require.Nil(err)
}
