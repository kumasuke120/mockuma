package mckmaps

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/kumasuke120/mockuma/internal/myjson"
	"github.com/kumasuke120/mockuma/internal/myos"
	"github.com/kumasuke120/mockuma/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//noinspection GoImportUsedAsName
func TestCommentFilter(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fb0, err := ioutil.ReadFile(filepath.Join("testdata", "preprocessor", "removeComment-0.json"))
	require.Nil(err)
	j0, err := myjson.Unmarshal(fb0)
	require.Nil(err)

	fbe, err := ioutil.ReadFile(filepath.Join("testdata", "preprocessor", "removeComment-expected.json"))
	require.Nil(err)
	je, err := myjson.Unmarshal(fbe)
	require.Nil(err)

	ja, err := types.DoFiltersOnV(j0, &dCommentProcessor{})
	assert.Nil(err)
	assert.Equal(je, ja)
}

//noinspection GoImportUsedAsName
func TestLoadFileFilter(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	require.Nil(myos.InitWd())
	oldWd := myos.GetWd()

	fb0, err := ioutil.ReadFile(filepath.Join("testdata", "preprocessor", "loadFile-0.json"))
	require.Nil(err)
	j0, err := myjson.Unmarshal(fb0)
	require.Nil(err)

	fbe, err := ioutil.ReadFile(filepath.Join("testdata", "preprocessor", "loadFile-expected.json"))
	require.Nil(err)
	je, err := myjson.Unmarshal(fbe)
	require.Nil(err)

	err = myos.Chdir(filepath.Join(oldWd, "testdata", "preprocessor"))
	require.Nil(err)

	ja, err := types.DoFiltersOnV(j0, makeDFileProcessor())
	if assert.Nil(err) {
		assert.Equal(je, ja)
	}

	require.Nil(myos.Chdir(oldWd))
}

//noinspection GoImportUsedAsName
func TestParseRegexp(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fb0, err := ioutil.ReadFile(filepath.Join("testdata", "preprocessor", "parseRegexp.json"))
	require.Nil(err)
	j0, err := myjson.Unmarshal(fb0)
	require.Nil(err)

	ja0, e0 := types.DoFiltersOnV(j0, &dJSONProcessor{}, makeDRegexpProcessor())
	if assert.Nil(e0) && assert.IsType(myjson.Array{}, ja0) {
		_ja0 := ja0.(myjson.Array)
		assert.IsType(myjson.ExtRegexp(nil), _ja0[0])
		o1, e1 := myjson.ToObject(_ja0.Get(1))
		if assert.Nil(e1) {
			assert.IsType(myjson.ExtRegexp(nil), o1.Get("r"))
		}
		o2, e2 := myjson.ToObject(o1.Get("j").(myjson.ExtJSONMatcher).Unwrap())
		if assert.Nil(e2) {
			assert.IsType(myjson.ExtRegexp(nil), o2.Get("r"))
		}
	}
}

//noinspection GoImportUsedAsName
func TestRenderTemplate(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	require.Nil(myos.InitWd())
	oldWd := myos.GetWd()

	fb0, err := ioutil.ReadFile(filepath.Join("testdata", "preprocessor", "renderTemplate.json"))
	require.Nil(err)
	j0, err := myjson.Unmarshal(fb0)
	require.Nil(err)

	fbe, err := ioutil.ReadFile(filepath.Join("testdata", "preprocessor", "renderTemplate-expected.json"))
	require.Nil(err)
	je, err := myjson.Unmarshal(fbe)
	require.Nil(err)

	err = myos.Chdir(filepath.Join(oldWd, "testdata", "preprocessor"))
	require.Nil(err)

	ja, err := types.DoFiltersOnV(j0, makeDTemplateProcessor())
	if assert.Nil(err) {
		assert.Equal(je, ja)
	}

	require.Nil(myos.Chdir(oldWd))
}

//noinspection GoImportUsedAsName
func TestToJSONMatcher(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fb0, err := ioutil.ReadFile(filepath.Join("testdata", "preprocessor", "toJSONMatcher.json"))
	require.Nil(err)
	j0, err := myjson.Unmarshal(fb0)
	require.Nil(err)

	expected := myjson.MakeExtJSONMatcher(myjson.Object{
		"a": myjson.MakeExtJSONMatcher(myjson.Object{
			"v": myjson.Number(1),
		}),
		"b1": myjson.Array{
			nil,
			myjson.Number(1),
			myjson.Number(2),
		},
		"b2": myjson.Array{
			myjson.MakeExtJSONMatcher(myjson.Object{
				"v": myjson.String("1"),
			}),
			myjson.MakeExtJSONMatcher(myjson.Object{
				"v": myjson.String("2"),
			}),
		},
		"b3": myjson.MakeExtJSONMatcher(myjson.Array{
			myjson.Number(1),
			myjson.Number(2),
		}),
		"$c": myjson.String("3"),
	})

	ja, err := types.DoFiltersOnV(j0, &dJSONProcessor{})
	if assert.Nil(err) {
		assert.Equal(expected, ja)
	}
}
