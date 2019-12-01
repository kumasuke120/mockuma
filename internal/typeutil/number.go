package typeutil

import "math"

const epsilon64 float64 = 1e-8

func Float64AlmostEquals(a, b float64) bool {
	if math.Abs(a-b) < epsilon64 {
		return true
	} else {
		return false
	}
}
