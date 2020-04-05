package types

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var order []string

type mapPutter struct {
	key   string
	value interface{}
}

func (p *mapPutter) DoFilter(v interface{}, chain *FilterChain) error {
	order = append(order, "p")

	if m, ok := v.(map[string]interface{}); ok {
		m[p.key] = p.value
		return chain.DoFilter(v)
	} else {
		return errors.New("incorrect type")
	}
}

type mapValueGetter struct {
	key string
}

func (g *mapValueGetter) DoFilter(v interface{}, chain *FilterChain) error {
	order = append(order, "g")

	if m, ok := v.(map[string]interface{}); ok {
		return chain.DoFilter(m[g.key])
	} else {
		return errors.New("incorrect type")
	}
}

func TestDoFiltersOnV(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	order = nil
	v0 := make(map[string]interface{})
	r0, e0 := DoFiltersOnV(v0,
		&mapPutter{key: "key", value: "value"},
		&mapValueGetter{key: "key"})
	if assert.Nil(e0) {
		assert.Equal("value", r0)
		assert.Equal([]string{"p", "g"}, order)
	}

	order = nil
	v1 := make(map[string]interface{})
	_, e1 := DoFiltersOnV(v1,
		&mapValueGetter{key: "key"},
		&mapPutter{key: "key", value: "value"})
	if assert.NotNil(e1) {
		assert.Equal("incorrect type", e1.Error())
		assert.Equal([]string{"g", "p"}, order)
	}
}
