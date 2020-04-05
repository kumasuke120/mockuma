package mckmaps

import (
	"path/filepath"
	"testing"

	"github.com/kumasuke120/mockuma/internal/myjson"
	"github.com/stretchr/testify/assert"
)

func TestVarsJSONParser_parse(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	t0 := &varsJSONParser{Parser: Parser{filename: filepath.Join("testdata", "vars", "vars-0.json")}}
	_, e0 := t0.parse()
	assert.NotNil(e0)

	t1 := &varsJSONParser{Parser: Parser{filename: filepath.Join("testdata", "vars", "vars-1.json")}}
	p1, e1 := t1.parse()
	if assert.Nil(e1) {
		expected1 := []*vars{
			{
				table: map[string]interface{}{
					"v": myjson.String("1"),
				},
			},
			{
				table: map[string]interface{}{
					"v": myjson.String("2"),
					"a": myjson.String("b"),
				},
			},
		}
		assert.Equal(expected1, p1)
	}
}

func TestVarsCSVParser_parse(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	expected := []*vars{
		{
			table: map[string]interface{}{
				"a": myjson.Number(1),
				"b": myjson.Boolean(true),
				"c": myjson.String("one"),
			},
		},
		{
			table: map[string]interface{}{
				"a": myjson.Number(0),
				"b": myjson.Boolean(false),
				"c": myjson.String("zero"),
			},
		},
	}

	t0 := &varsCSVParser{Parser: Parser{filename: filepath.Join("testdata", "vars", "vars-0.csv")}}
	p0, e0 := t0.parse()
	if assert.Nil(e0) {
		assert.Equal(expected, p0)
	}

	t1 := &varsCSVParser{Parser: Parser{filename: filepath.Join("testdata", "vars", "vars-1.csv")}}
	p1, e1 := t1.parse()
	if assert.Nil(e1) {
		assert.Equal(expected, p1)
	}
}
