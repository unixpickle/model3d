package numerical

import "math"

// GSS performs golden section search to find the minimum
// of a unimodal function, given correct bounds on the
// minimum.
func GSS(min, max float64, f func(float64) float64) float64 {
	phi := (math.Sqrt(5) + 1) / 2
	mid1 := max - (max-min)/phi
	mid2 := min + (max-min)/phi
	val1 := f(mid1)
	val2 := f(mid2)
	for i := 0; i < 32; i++ {
		if val2 > val1 {
			max = mid2
			mid2 = mid1
			val2 = val1
			mid1 = max - (max-min)/phi
			val1 = f(mid1)
		} else {
			min = mid1
			mid1 = mid2
			val1 = val2
			mid2 = min + (max-min)/phi
			val2 = f(mid2)
		}
	}

	if val1 < val2 {
		return mid1
	} else {
		return mid2
	}
}
