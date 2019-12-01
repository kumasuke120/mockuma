package myjson

import "testing"

func TestParsePath(t *testing.T) {
	p1Str := "$[0].name['@type']"
	p1, e1 := ParsePath(p1Str)
	if e1 != nil {
		t.Errorf("p1: shoudn't error")
	} else if p1.String() != p1Str {
		t.Errorf("p1: parse failed")
	}

	p2, e2 := ParsePath("$['first'][1].second")
	if e2 != nil {
		t.Errorf("p2: shoudn't error")
	} else if p2.String() != "$.first[1].second" {
		t.Errorf("p2: parse failed")
	}
}
