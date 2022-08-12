package toolbox3d

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

// Teardrop2D is a 2D solid in a "teardrop" shape, i.e. a
// circle with a tangent triangle pointing off in one
// direction.
//
// This can be used to cut circles out of shapes while
// avoiding support structures in FDM printing.
//
// The shape, rendered in ASCII, looks like so:
//
//	            --(&)--
//	       @@@@/       (@@@@
//	    @@@                 @@@
//	  /@@                     @@.
//	 @@                         @@
//	 @@                         @@
//	&@                           @
//	 @%                         @@
//	 @@                         @@
//	  @@#                     @@@
//	    @@@                 @@@
//	      &@@             @@*
//	         @@&       @@@
//	           @@@   @@#
//	              @@@
//	               ^
type Teardrop2D struct {
	// Center is the center of the circle.
	Center model2d.Coord

	// Radius is the radius of the circle.
	Radius float64

	// Direction is the direction in which the tip of the
	// tangent triangle is pointing.
	// If this is the zero vector, then a unit vector in
	// the Y direction is used.
	Direction model2d.Coord
}

func (t *Teardrop2D) Min() model2d.Coord {
	extraRadius := math.Sqrt2 * t.Radius
	return t.Center.Sub(model2d.XY(extraRadius, extraRadius))
}

func (t *Teardrop2D) Max() model2d.Coord {
	extraRadius := math.Sqrt2 * t.Radius
	return t.Center.Add(model2d.XY(extraRadius, extraRadius))
}

func (t *Teardrop2D) Contains(c model2d.Coord) bool {
	c2 := c.Sub(t.Center)
	mag := c2.Norm()
	if mag <= t.Radius {
		return true
	}

	// Rotate c2 so that the tip direction is Y.
	if (t.Direction != model2d.Coord{}) {
		axisY := t.Direction.Normalize()
		axisX := model2d.XY(axisY.Y, -axisY.X)
		c2 = model2d.XY(axisX.Dot(c2), axisY.Dot(c2))
	}

	// Under this Y value, we can't be in the tip.
	if c2.Y < t.Radius/math.Sqrt2 {
		return false
	}

	triangleVec := model2d.XY(math.Sqrt2/2, math.Sqrt2/2)
	for i := 0; i < 2; i++ {
		if c2.Dot(triangleVec) > t.Radius {
			return false
		}
		triangleVec.X *= -1
	}
	return true
}

// Teardrop3D creates a 3D teardrop by extending a profile
// Teardrop2D between p1 and p2.
//
// If possible, the point of the teardrop will be facing
// into the positive Z direction to avoid supports.
func Teardrop3D(p1, p2 model3d.Coord3D, radius float64) model3d.Solid {
	length := p1.Dist(p2)
	profileSolid := model3d.ProfileSolid(&Teardrop2D{
		Radius: radius,
	}, 0, length)
	zVec := p2.Sub(p1).Normalize()
	yVec := model3d.Z(1).ProjectOut(zVec)
	if yVec.Norm() < 1e-5 {
		// If the profile is extended into the Z direction,
		// then we default to use the Y axis for the tip.
		yVec = model3d.Y(1).ProjectOut(zVec).Normalize()
	} else {
		yVec = yVec.Normalize()
	}
	xVec := yVec.Cross(zVec)
	matrix := model3d.NewMatrix3Columns(xVec, yVec, zVec)
	xform := model3d.JoinedTransform{
		&model3d.Matrix3Transform{Matrix: matrix},
		&model3d.Translate{Offset: p1},
	}
	return model3d.TransformSolid(xform, profileSolid)
}
