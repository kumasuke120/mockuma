package myjson

func IsAllNumber(v Array) bool {
	for _, _v := range v {
		if !IsNumber(_v) {
			return false
		}
	}

	return true
}

func IsNumber(v interface{}) bool {
	switch v.(type) {
	case Number:
		return true
	default:
		return false
	}
}
