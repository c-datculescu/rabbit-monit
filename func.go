package rabbitmonit

import (
	"math"
)

/*
Round performs rounding of a float64
*/
func Round(f float64) float64 {
	return math.Floor(f + .5)
}

/*
RoundPlus performs the decimal rounding to n specified integer places
*/
func RoundPlus(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return Round(f*shift) / shift
}
