package mckmaps

import "github.com/kumasuke120/mockuma/internal/myjson"

const dComment = "@comment"

type commentProcessor struct {
	v interface{}
}

func (p commentProcessor) process() {
	removeComment(p.v)
}

func removeComment(v interface{}) {
	switch v.(type) {
	case myjson.Object:
		delete(v.(myjson.Object), dComment)
	case myjson.Array:
		for _, _v := range v.(myjson.Array) {
			removeComment(_v)
		}
	}
}
