package myhttp

import (
	"net/http"
	"strings"
)

type HTTPMethod string

const (
	Any     = HTTPMethod("*")
	Options = HTTPMethod(http.MethodOptions)
	Get     = HTTPMethod(http.MethodGet)
	Head    = HTTPMethod(http.MethodHead)
	Post    = HTTPMethod(http.MethodPost)
	Put     = HTTPMethod(http.MethodPut)
	Delete  = HTTPMethod(http.MethodDelete)
	Trace   = HTTPMethod(http.MethodTrace)
	Connect = HTTPMethod(http.MethodConnect)
	Patch   = HTTPMethod(http.MethodPatch)
)

type StatusCode int

const (
	Ok               = StatusCode(http.StatusOK)
	BadRequest       = StatusCode(http.StatusBadRequest)
	MethodNotAllowed = StatusCode(http.StatusMethodNotAllowed)
	NotFound         = StatusCode(http.StatusNotFound)
)

func ToHTTPMethod(method interface{}) HTTPMethod {
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

func (m HTTPMethod) Matches(s string) bool {
	return m == Any || m == ToHTTPMethod(s)
}

func (m HTTPMethod) String() string {
	return string(m)
}
