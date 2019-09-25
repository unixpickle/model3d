package model3d

import "math"

// A Ray is a line originating at a point and extending
// infinitely in some direction.
//
// The direction should be a unit vector.
type Ray struct {
	Origin    Coord3D
	Direction Coord3D
}

// Collision computes where (and if) the ray intersects
// the triangle.
//
// If it returns true as the first value, then the ray or
// its reverse hits the triangle.
//
// The second return value is how much of the direction
// must be added to the origin to hit the plane spanned by
// the triangle.
// If it is negative, it means the triangle is behind the
// ray.
func (r *Ray) Collision(t *Triangle) (bool, float64) {
	matrix := Matrix3{
		t[1].X - t[0].X, t[2].X - t[0].X, r.Direction.X,
		t[1].Y - t[0].Y, t[2].Y - t[0].Y, r.Direction.Y,
		t[1].Z - t[0].Z, t[2].Z - t[0].Z, r.Direction.Z,
	}
	if math.Abs(matrix.Det()) < 1e-8 {
		return false, 0
	}
	result := matrix.Inverse().ApplyColumn(r.Origin.Add(t[0].Scale(-1)))
	return result.X >= 0 && result.Y >= 0 && result.X+result.Y <= 1, -result.Z
}
