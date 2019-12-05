package mckmaps

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/kumasuke120/mockuma/internal/myjson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//noinspection GoImportUsedAsName
func TestMappingsParser_parse(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fb1, e1 := ioutil.ReadFile(filepath.Join("testdata", "mappings-1.json"))
	require.Nil(e1)
	j1, e1 := myjson.Unmarshal(fb1)
	assert.Nil(e1)
	m1 := &mappingsParser{json: j1}
	p1, e1 := m1.parse()
	assert.Nil(e1)
	fmt.Println(p1)
}
