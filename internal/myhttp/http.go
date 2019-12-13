package myhttp

import (
	"net/http"
	"strings"
)

type HttpMethod string

const (
	Any     = HttpMethod("*")
	Options = HttpMethod(http.MethodOptions)
	Get     = HttpMethod(http.MethodGet)
	Head    = HttpMethod(http.MethodHead)
	Post    = HttpMethod(http.MethodPost)
	Put     = HttpMethod(http.MethodPut)
	Delete  = HttpMethod(http.MethodDelete)
	Trace   = HttpMethod(http.MethodTrace)
	Connect = HttpMethod(http.MethodConnect)
	Patch   = HttpMethod(http.MethodPatch)
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
		case http.MethodOptions:
			return Options
		case http.MethodGet:
			return Get
		case http.MethodHead:
			return Head
		case http.MethodPost:
			return Post
		case http.MethodPut:
			return Put
		case http.MethodDelete:
			return Delete
		case http.MethodTrace:
			return Trace
		case http.MethodConnect:
			return Connect
		case http.MethodPatch:
			return Patch
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
