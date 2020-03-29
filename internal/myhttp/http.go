package myhttp

import (
	"net/http"
	"strings"
)

type HTTPMethod string

const (
	MethodAny     = HTTPMethod("*")
	MethodOptions = HTTPMethod(http.MethodOptions)
	MethodGet     = HTTPMethod(http.MethodGet)
	MethodHead    = HTTPMethod(http.MethodHead)
	MethodPost    = HTTPMethod(http.MethodPost)
	MethodPut     = HTTPMethod(http.MethodPut)
	MethodDelete  = HTTPMethod(http.MethodDelete)
	MethodTrace   = HTTPMethod(http.MethodTrace)
	MethodConnect = HTTPMethod(http.MethodConnect)
	MethodPatch   = HTTPMethod(http.MethodPatch)
)

type StatusCode int

const (
	StatusOk = StatusCode(http.StatusOK)

	StatusFound = StatusCode(http.StatusFound)

	StatusBadRequest       = StatusCode(http.StatusBadRequest)
	StatusMethodNotAllowed = StatusCode(http.StatusMethodNotAllowed)
	StatusNotFound         = StatusCode(http.StatusNotFound)
)

func ToHTTPMethod(method string) HTTPMethod {
	upper := strings.ToUpper(method)
	switch upper {
	case "*":
		return MethodAny
	case http.MethodOptions:
		return MethodOptions
	case http.MethodGet:
		return MethodGet
	case http.MethodHead:
		return MethodHead
	case http.MethodPost:
		return MethodPost
	case http.MethodPut:
		return MethodPut
	case http.MethodDelete:
		return MethodDelete
	case http.MethodTrace:
		return MethodTrace
	case http.MethodConnect:
		return MethodConnect
	case http.MethodPatch:
		return MethodPatch
	default:
		return HTTPMethod(upper)
	}
}

func (m HTTPMethod) Matches(s string) bool {
	return m == MethodAny || m == ToHTTPMethod(s)
}

func (m HTTPMethod) String() string {
	return string(m)
}
