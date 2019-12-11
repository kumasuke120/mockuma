package myhttp

import (
	"net/http"
	"strings"
)

type HttpMethod string

const (
	Any     = HttpMethod("*")
	Options = HttpMethod("OPTIONS")
	Get     = HttpMethod("GET")
	Head    = HttpMethod("HEAD")
	Post    = HttpMethod("POST")
	Put     = HttpMethod("PUT")
	Delete  = HttpMethod("DELETE")
	Trace   = HttpMethod("TRACE")
	Connect = HttpMethod("CONNECT")
)

type StatusCode int

const (
	Ok               = StatusCode(http.StatusOK)
	BadRequest       = StatusCode(http.StatusBadRequest)
	MethodNotAllowed = StatusCode(http.StatusMethodNotAllowed)
	NotFound         = StatusCode(http.StatusNotFound)
)

func ToHttpMethod(method interface{}) HttpMethod {
	switch method.(type) {
	case string:
		switch strings.ToUpper(method.(string)) {
		case "OPTIONS":
			return Options
		case "GET":
			return Get
		case "HEAD":
			return Head
		case "POST":
			return Post
		case "PUT":
			return Put
		case "DELETE":
			return Delete
		case "TRACE":
			return Trace
		case "CONNECT":
			return Connect
		}
	}

	return Any
}

func (m HttpMethod) Matches(s string) bool {
	return m == Any || m == ToHttpMethod(s)
}

func (m HttpMethod) String() string {
	return string(m)
}
