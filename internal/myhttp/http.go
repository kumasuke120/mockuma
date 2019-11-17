package myhttp

import (
	"strings"
)

type HttpMethod int

const (
	Any = iota
	Options
	Get
	Head
	Post
	Put
	Delete
	Trace
	Connect
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
	case Options:
		return "OPTIONS"
	case Get:
		return "GET"
	case Head:
		return "HEAD"
	case Post:
		return "POST"
	case Put:
		return "PUT"
	case Delete:
		return "DELETE"
	case Trace:
		return "TRACE"
	case Connect:
		return "Connect"
	default:
		panic("Should't happen")
	}
}
