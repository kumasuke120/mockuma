package typeutil

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloat64AlmostEquals(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	float1, err := strconv.ParseFloat("41.98", 64)
	assert.Nil(err)
	assert.True(Float64AlmostEquals(41.98, float1))

	assert.False(Float64AlmostEquals(1, 2))
}
