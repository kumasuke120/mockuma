package mckmaps

import "github.com/kumasuke120/mockuma/internal/myjson"

var (
	fRemoveComment = commentFilter{}
)

type filterChain struct {
	filters []filter
	idx     int
	v       interface{}
}

func (c *filterChain) doFilter(v interface{}) error {
	c.v = v

	if c.idx < len(c.filters) {
		next := c.filters[c.idx]
		c.idx += 1
		return next.doFilter(v, c)
	}

	return nil
}

type filter interface {
	doFilter(v interface{}, chain *filterChain) error
}

type commentFilter struct {
}

func (f commentFilter) doFilter(v interface{}, chain *filterChain) error {
	f.removeComment(v)
	return chain.doFilter(v)
}

func (f commentFilter) removeComment(v interface{}) {
	switch v.(type) {
	case myjson.Object:
		delete(v.(myjson.Object), dComment)
	case myjson.Array:
		for _, _v := range v.(myjson.Array) {
			f.removeComment(_v)
		}
	}
}

func doFiltersOnV(v interface{}, filters ...filter) (interface{}, error) {
	chain := &filterChain{filters: filters}
	err := chain.doFilter(v)
	if err != nil {
		return nil, err
	}

	return chain.v, nil
}
