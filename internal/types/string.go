package types

import "fmt"

func ToString(value interface{}) string {
	switch value.(type) {
	case string:
		return value.(string)
	case fmt.Stringer:
		return value.(fmt.Stringer).String()
	default:
		return fmt.Sprintf("%v", value)
	}
}
