package toolbox3d

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

// CoordColorFunc wraps a generic point-to-color function
// and provides methods for various other color-using APIs.
type CoordColorFunc func(c model3d.Coord3D) render3d.Color

// RenderColor is a render3d.ColorFunc wrapper for c.F.
func (c CoordColorFunc) RenderColor(coord model3d.Coord3D, rc model3d.RayCollision) render3d.Color {
	return c(coord)
}

// TriangeColor returns sRGB colors for triangles by
// averaging the sRGB values of each vertex.
func (c CoordColorFunc) TriangleColor(t *model3d.Triangle) [3]float64 {
	sum := [3]float64{}
	for _, coord := range t {
		vertexColor := c(coord)
		r, g, b := render3d.RGB(vertexColor)
		sum[0] += r / 3
		sum[1] += g / 3
		sum[2] += b / 3
	}
	return sum
}
