package types

func DoFiltersOnV(v interface{}, filters ...Filter) (interface{}, error) {
	chain := &FilterChain{filters: filters}
	err := chain.DoFilter(v)
	if err != nil {
		return nil, err
	}
	return chain.v, nil
}

type Filter interface {
	DoFilter(v interface{}, chain *FilterChain) error
}

type FilterChain struct {
	filters []Filter
	idx     int
	v       interface{}
}

func (c *FilterChain) DoFilter(v interface{}) error {
	c.v = v

	if c.idx < len(c.filters) {
		next := c.filters[c.idx]
		c.idx += 1
		return next.DoFilter(v, c)
	}

	return nil
}
