package numerical

import "math"

const DefaultGSSIters = 64

// GSS performs golden section search to find the minimum
// of a unimodal function, given correct bounds on the
// minimum.
//
// If iters is 0, DefaultGSSIters is used.
func GSS(min, max float64, iters int, f func(float64) float64) float64 {
	if iters == 0 {
		iters = DefaultGSSIters
	}
	phi := (math.Sqrt(5) + 1) / 2
	mid1 := max - (max-min)/phi
	mid2 := min + (max-min)/phi
	val1 := f(mid1)
	val2 := f(mid2)
	for i := 0; i < iters; i++ {
		if mid1 <= min || mid2 <= mid1 || max <= mid2 {
			// Numerical precision has been exhausted.
			break
		}
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
