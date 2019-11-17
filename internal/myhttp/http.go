package myhttp

import (
	"strings"
)

type HttpMethod int

const (
	Any = iota
	Get
	Post
	Put
	Delete
)

type StatusCode int

const (
	Ok         = StatusCode(200)
	BadRequest = StatusCode(400)
	NotFound   = StatusCode(404)
)

const (
	HeaderServer      = "Server"
	HeaderContentType = "Content-Type"
)

const ContentTypeJson = "application/json; charset=utf8"

func ToHttpMethod(method interface{}) HttpMethod {
	switch method.(type) {
	case string:
		switch strings.ToUpper(method.(string)) {
		case "GET":
			return Get
		case "POST":
			return Post
		case "PUT":
			return Put
		case "DELETE":
			return Delete
		default:
			return Any
		}
	default:
		return Any
	}
}

func (m HttpMethod) Matches(s string) bool {
	return m == Any || m == ToHttpMethod(s)
}

func (m HttpMethod) String() string {
	switch m {
	case Any:
		return "Any"
	case Get:
		return "GET"
	case Post:
		return "POST"
	case Put:
		return "PUT"
	case Delete:
		return "DELETE"
	default:
		panic("Should't happen")
	}
}
