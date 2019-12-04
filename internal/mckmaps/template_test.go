package mckmaps

import (
	"testing"

	"github.com/kumasuke120/mockuma/internal/myjson"
	"github.com/stretchr/testify/assert"
)

var theContext = &renderContext{filename: "test"}
var theJsonPath = myjson.NewPath()
var theVars = &vars{table: map[string]interface{}{
	"a":  myjson.String("123"),
	"b1": myjson.Number(123),
	"b2": myjson.Number(123.123),
	"c":  myjson.Boolean(true),
}}

func TestRenderString(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	t1 := myjson.String("a = \"@{a}\", b1 = @{b1}, b2 = @{b2}, c = @{c}")
	r1, e1 := renderString(theContext, theJsonPath, t1, theVars)
	assert.Nil(e1)
	assert.IsType(myjson.String(""), r1)
	assert.Equal("a = \"123\", b1 = 123, b2 = 123.123, c = true", string(r1.(myjson.String)))

	t2 := myjson.String("a = \"@{a:%8s}\", b1 = ORD@{b1:%010d}, b2 = @{b2:%.2f}, c = @{c}")
	r2, e2 := renderString(theContext, theJsonPath, t2, theVars)
	assert.Nil(e2)
	assert.IsType(myjson.String(""), r2)
	assert.Equal("a = \"     123\", b1 = ORD0000000123, b2 = 123.12, c = true", string(r2.(myjson.String)))
}
