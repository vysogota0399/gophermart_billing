package utils

import "math"

func Delay(i uint8) float64 {
	return math.Exp(float64(i) / 2)
}
