package mckmaps

import (
	"path/filepath"
	"testing"

	"github.com/kumasuke120/mockuma/internal/myjson"
	"github.com/stretchr/testify/assert"
)

func TestTemplateParser_parse(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

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
}
