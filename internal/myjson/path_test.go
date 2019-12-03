package myjson

import (
	"testing"
)

func TestParsePath(t *testing.T) {
	p1Str := "$[0].name['@type']"
	p1, e1 := ParsePath(p1Str)
	if e1 != nil {
		t.Error("p1: shouldn't error")
	} else if p1.String() != p1Str {
		t.Error("p1: parse failed")
	}

	p2, e2 := ParsePath("$['first'][1].second")
	if e2 != nil {
		t.Error("p2: shouldn't error")
	} else if p2.String() != "$.first[1].second" {
		t.Error("p2: parse failed")
	}

	_, e3 := ParsePath("$$.first.second")
	if e3 == nil {
		t.Error("p3: should be error")
	}
}

func TestObject_SetByPath(t *testing.T) {
	o1Path := NewPath("first", 2, "third")
	o1, e1 := Object{}.SetByPath(o1Path, String("value"))
	if e1 != nil {
		t.Error("o1: shouldn't error")
	} else if o1.String() != `map[first:[<nil> <nil> map[third:"value"]]]` {
		t.Error("o1: set failed")
	}
}
