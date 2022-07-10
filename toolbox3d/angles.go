package toolbox3d

import "math"

// AngleDist computes the (shortest) absolute distance in
// radians between two angles. In the most basic case,
// this is equivalent to ||theta2 - theta1||. However, it
// handles the fact that angles can be offset by multiples
// of 2*pi.
func AngleDist(theta1, theta2 float64) float64 {
	theta1 = CanonicalAngle(theta1)
	theta2 = CanonicalAngle(theta2)
	diff := math.Abs(theta1 - theta2)
	return math.Min(diff, 2*math.Pi-diff)
}

// CanonicalAngle converts an angle to the range [0, 2*pi].
func CanonicalAngle(theta float64) float64 {
	if theta < 0 {
		return CanonicalAngle(2*math.Pi - theta)
	}
	return math.Mod(theta, 2*math.Pi)
}
