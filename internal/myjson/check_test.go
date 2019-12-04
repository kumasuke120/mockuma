package myjson

import "testing"

func TestIsAllNumber(t *testing.T) {
	v1 := []interface{}{Number(1), Number(2)}
	if !IsAllNumber(v1) {
		t.Error("v1: failed")
	}

	v2 := []interface{}{Number(1), Boolean(false)}
	if IsAllNumber(v2) {
		t.Error("v2: failed")
	}
}

func TestIsNumber(t *testing.T) {
	v1 := Number(1)
	if !IsNumber(v1) {
		t.Error("v1: failed")
	}

	v2 := String("")
	if IsNumber(v2) {
		t.Error("v2: failed")
	}
}
