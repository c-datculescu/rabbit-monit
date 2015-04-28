package rabbitmonit

import (
	"math"
)

func round(f float64) float64 {
	return math.Floor(f + .5)
}

func roundPlus(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return round(f*shift) / shift
}
