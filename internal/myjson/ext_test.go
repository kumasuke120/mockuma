package myjson

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtJsonMatcher_Matches(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	jm := MakeExtJsonMatcher(Object(map[string]interface{}{
		"a": Array([]interface{}{
			nil,
			Number(1),
			Number(2),
		}),
		"b": nil,
		"c": Array([]interface{}{
			String("s"),
			String("1"),
			String("true"),
		}),
		"d": Array([]interface{}{
			Boolean(true),
			Boolean(false),
		}),
		"e": Array([]interface{}{
			ExtRegexp(regexp.MustCompile("^ab\\w$")),
			ExtRegexp(regexp.MustCompile("^\\d{3}$")),
			ExtRegexp(regexp.MustCompile("^true$")),
			ExtRegexp(regexp.MustCompile("^test")),
		}),
	}))

	o1 := Object(map[string]interface{}{
		"a": Array([]interface{}{
			Number(0),
			Number(1),
			String("2"),
		}),
		"b": nil,
		"c": Array([]interface{}{
			String("s"),
			Number(1),
			Boolean(true),
		}),
		"d": Array([]interface{}{
			Boolean(true),
			String("false"),
		}),
		"e": Array([]interface{}{
			String("abc"),
			Number(120),
			Boolean(true),
			ExtRegexp(regexp.MustCompile("^test")),
		}),
	})
	assert.True(jm.Matches(o1))

	o2 := Object(map[string]interface{}{
		"a": Array([]interface{}{
			Number(0),
			Boolean(true),
			String("2"),
		}),
		"b": nil,
		"c": Array([]interface{}{
			String("s"),
			Number(1),
			Boolean(true),
		}),
		"d": Array([]interface{}{
			Boolean(true),
			String("false"),
		}),
		"e": Array([]interface{}{
			String("abc"),
			Number(120),
			Boolean(true),
			ExtRegexp(regexp.MustCompile("^test")),
		}),
	})
	assert.False(jm.Matches(o2))

	o3 := Object(map[string]interface{}{
		"a": Object{},
		"b": nil,
		"c": Array([]interface{}{
			String("s"),
			Number(1),
			Boolean(true),
		}),
		"d": Array([]interface{}{
			Boolean(true),
			String("false"),
		}),
		"e": Array([]interface{}{
			String("abc"),
			Number(120),
			Boolean(true),
			ExtRegexp(regexp.MustCompile("^test")),
		}),
	})
	assert.False(jm.Matches(o3))

	o4 := Object(map[string]interface{}{
		"a": Array([]interface{}{
			Number(0),
			Number(1),
			String("2"),
		}),
		"b": nil,
		"c": Array([]interface{}{
			Object{},
			Number(1),
			Boolean(true),
		}),
		"d": Array([]interface{}{
			Boolean(true),
			String("false"),
		}),
		"e": Array([]interface{}{
			String("abc"),
			Number(120),
			Boolean(true),
			ExtRegexp(regexp.MustCompile("^test")),
		}),
	})
	assert.False(jm.Matches(o4))

	o5 := Object(map[string]interface{}{
		"a": Array([]interface{}{
			Number(0),
			Number(1),
			String("2"),
		}),
		"b": nil,
		"c": Array([]interface{}{
			String("s"),
			Number(1),
			Boolean(true),
		}),
		"d": Array([]interface{}{
			Boolean(true),
			Number(0),
		}),
		"e": Array([]interface{}{
			String("abc"),
			Number(120),
			Boolean(true),
			ExtRegexp(regexp.MustCompile("^test")),
		}),
	})
	assert.False(jm.Matches(o5))

	o6 := Object(map[string]interface{}{
		"a": Array([]interface{}{
			Number(0),
			Number(1),
			String("2"),
		}),
		"b": nil,
		"c": Array([]interface{}{
			String("s"),
			Number(1),
			Boolean(true),
		}),
		"d": Array([]interface{}{
			Boolean(true),
			String("false"),
		}),
		"e": Array([]interface{}{
			String("abc"),
			Number(120),
			Boolean(true),
			ExtRegexp(regexp.MustCompile("^test$")),
		}),
	})
	assert.False(jm.Matches(o6))
}
