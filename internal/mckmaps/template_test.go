package mckmaps

import (
	"path/filepath"
	"testing"

	"github.com/kumasuke120/mockuma/internal/myjson"
	"github.com/kumasuke120/mockuma/internal/myos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//noinspection GoImportUsedAsName
func TestTemplateParser_parse(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp0 := filepath.Join("testdata", "template-0.json")
	t0 := &templateParser{Parser: Parser{filename: fp0}}
	_, e0 := t0.parse()
	assert.NotNil(e0)

	fp1 := filepath.Join("testdata", "template-1.json")
	t1 := &templateParser{Parser: Parser{filename: fp1}}
	p1, e1 := t1.parse()
	if assert.Nil(e1) {
		expected1 := &template{
			content: myjson.Object{
				"v":   myjson.String("@{v}"),
				"v-v": myjson.String("@{v}-@{v}"),
			},
			filename: fp1,
		}
		assert.Equal(expected1, p1)
	}

	fp2 := filepath.Join("testdata", "template-2.json")
	t2 := &templateParser{Parser: Parser{filename: fp2}}
	p2, e2 := t2.parse()
	if assert.Nil(e2) {
		expected2 := &template{
			content: myjson.Array{
				myjson.Object{
					"v": myjson.String("@{v}"),
				},
				myjson.Object{
					"v-v": myjson.String("@{v}-@{v}"),
				},
			},
			filename: fp2,
		}
		assert.Equal(expected2, p2)
	}

	require.Nil(myos.InitWd())
	oldWd := myos.GetWd()
	require.Nil(myos.Chdir(filepath.Join(oldWd, "testdata")))

	fp3 := "template-3.json"
	t3 := &templateParser{Parser: Parser{filename: fp3}}
	p3, e3 := t3.parse()
	if assert.Nil(e3) {
		expected3 := &template{
			content: myjson.Array{
				myjson.Object{
					"v": myjson.String("@{v}"),
				},
				myjson.Object{
					"v-v": myjson.String("@{v}-@{v}"),
				},
			},
			filename: fp3,
		}
		assert.Equal(expected3, p3)
	}

	fp4 := "template-cyclic-a.json"
	t4 := &templateParser{Parser: Parser{filename: fp4}}
	_, e4 := t4.parse()
	assert.NotNil(e4)
	assert.Contains(e4.Error(), "cyclic")

	require.Nil(myos.Chdir(oldWd))
}
