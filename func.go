package rabbitmonit

import (
	"math"
	"strconv"
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

/*
reduceInt reduces an integer to kilo, mega, giga corresponding sizes
*/
func reduceInt(in int) string {
	scale := 1000
	sizes := []string{
		"",
		"K",
		"M",
		"G",
		"T",
		"P",
		"E",
	}

	element := math.Floor(math.Log(float64(in)) / math.Log(float64(scale)))
	if in == 0 {
		return "0"
	}

	// get the final result
	result := float64(in) / math.Pow(float64(scale), element)

	// convert the result into the final return
	return strconv.FormatFloat(result, 'f', 1, 64) + sizes[int(element)]
}
