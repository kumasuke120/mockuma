package mckmaps

import (
	"reflect"
	"testing"

	"github.com/kumasuke120/mockuma/internal/myjson"
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
	t1 := myjson.String("a = \"@{a}\", b1 = @{b1}, b2 = @{b2}, c = @{c}")
	r1, e1 := renderString(theContext, theJsonPath, t1, theVars)
	if e1 != nil {
		t.Error("r1: shouldn't error")
	} else {
		if reflect.TypeOf(r1) != reflect.TypeOf(myjson.String("")) {
			t.Error("r1: type should be myjson.String")
		}
		if string(r1.(myjson.String)) != "a = \"123\", b1 = 123, b2 = 123.123, c = true" {
			t.Error("r1: render failed")
		}
	}

	t2 := myjson.String("a = \"@{a:%8s}\", b1 = ORD@{b1:%010d}, b2 = @{b2:%.2f}, c = @{c}")
	r2, e2 := renderString(theContext, theJsonPath, t2, theVars)
	if e2 != nil {
		t.Error("r2: shouldn't error")
	} else {
		if reflect.TypeOf(r2) != reflect.TypeOf(myjson.String("")) {
			t.Error("r2: type should be myjson.String")
		}
		if string(r2.(myjson.String)) != "a = \"     123\", b1 = ORD0000000123, b2 = 123.12, c = true" {
			t.Error("r2: render failed")
		}
	}
}
