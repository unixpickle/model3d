package render3d

import (
	"math"

	"github.com/unixpickle/model3d"
)

// A Camera defines a viewer's position, orientation, and
// field of view for rendering.
//
// The right-hand rule is used to determine which way the
// camera is facing, such that if the viewing plane goes
// from top to bottom and left to right, then the rays of
// sight go away from the camera's origin.
// To reverse this, simply use a negative FieldOfView.
type Camera struct {
	// Origin is the location of the camera, from whence
	// lines if sight originate.
	Origin model3d.Coord3D

	// ScreenX is the (normalized) direction in 3D space
	// that is rendered along the x-axis in images.
	// In other words, it is parallel to the x-axis on the
	// viewing plane.
	ScreenX model3d.Coord3D

	// ScreenY is the (normalized) direction in 3D space
	// that is rendered along the y-axis in images.
	// See ScreenX.
	ScreenY model3d.Coord3D

	// FieldOfView is the angle spanning the viewing plane
	// from the camera's origin.
	//
	// This is measured in radians.
	FieldOfView float64
}

// Caster produces a function that converts image
// coordinates into directions for rays that emenate from
// the origin.
//
// Arguments to the resulting function are x and y values
// ranging from [0, imageWidth] and [0, imageHeight].
func (c *Camera) Caster(imageWidth, imageHeight float64) func(x, y float64) model3d.Coord3D {
	planeDistance := 1 / math.Tan(c.FieldOfView/2)

	x, y := c.ScreenX, c.ScreenY
	z := x.Cross(y).Normalize()
	if imageWidth > imageHeight {
		y = y.Scale(imageHeight / imageWidth)
	} else {
		x = x.Scale(imageWidth / imageHeight)
	}
	z = z.Scale(planeDistance)

	cx, cy := imageWidth/2, imageHeight/2
	return func(imgX, imgY float64) model3d.Coord3D {
		outX := x.Scale((imgX - cx) / cx)
		outY := y.Scale((imgY - cy) / cy)
		return outX.Add(outY).Add(z)
	}
}
