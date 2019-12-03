package myhttp

import "testing"

func TestToHttpMethod(t *testing.T) {
	if ToHttpMethod("get") != Get {
		t.Error("get: failed")
	}
	if ToHttpMethod("GET") != Get {
		t.Error("GET: failed")
	}
	if ToHttpMethod("Put") != Put {
		t.Error("Put: failed")
	}
}

func TestHttpMethod_Matches(t *testing.T) {
	if !Get.Matches("GET") {
		t.Error("Get-GET: failed")
	}
	if Post.Matches("GET") {
		t.Error("Post-GET: failed")
	}
	if !Any.Matches("OPTIONS") {
		t.Error("Any-OPTIONS: failed")
	}
}
