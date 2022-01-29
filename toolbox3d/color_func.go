package toolbox3d

import (
	"fmt"
	"math"
	"sync"

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

// Cached wraps c in another CoordColorFunc that caches
// colors for coordinates.
//
// The cached function is safe to call concurrently from
// multiple Goroutines at once.
func (c CoordColorFunc) Cached() CoordColorFunc {
	cache := &sync.Map{}
	return func(coord model3d.Coord3D) render3d.Color {
		value, ok := cache.Load(coord)
		if ok {
			return value.(render3d.Color)
		}
		actual := c(coord)
		cache.Store(coord, actual)
		return actual
	}
}

// Transform wraps c in another CoordColorFunc that applies
// the inverse of t to input points.
func (c CoordColorFunc) Transform(t model3d.Transform) CoordColorFunc {
	tInv := t.Inverse()
	return func(coord model3d.Coord3D) render3d.Color {
		return c(tInv.Apply(coord))
	}
}

// ConstantCoordColorFunc creates a CoordColorFunc that
// returns a constant value.
func ConstantCoordColorFunc(c render3d.Color) CoordColorFunc {
	return func(x model3d.Coord3D) render3d.Color {
		return c
	}
}

// JoinedCoordColorFunc creates a CoordColorFunc that
// evaluates separate CoordColorFunc for different objects,
// where the object with maximum SDF is chosen.
//
// Pass a sequence of object, color, object, color, ...
// where objects are *model3d.Mesh or model3d.SDF, and
// colors are render3d.Color or CoordColorFunc.
func JoinedCoordColorFunc(sdfsAndColors ...interface{}) CoordColorFunc {
	if len(sdfsAndColors)%2 != 0 {
		panic("must pass an even number of arguments")
	}
	sdfs := make([]model3d.SDF, 0, len(sdfsAndColors)/2)
	colorFns := make([]CoordColorFunc, 0, len(sdfsAndColors)/2)
	for i := 0; i < len(sdfsAndColors); i += 2 {
		switch obj := sdfsAndColors[i].(type) {
		case model3d.SDF:
			sdfs = append(sdfs, obj)
		case *model3d.Mesh:
			sdfs = append(sdfs, model3d.MeshToSDF(obj))
		default:
			panic(fmt.Sprintf("unknown type for object: %T", obj))
		}
		colorFns = append(colorFns, colorFuncFromObj(sdfsAndColors[i+1]))
	}
	return func(c model3d.Coord3D) render3d.Color {
		maxSDF := math.Inf(-1)
		var maxColorFn CoordColorFunc
		for i, sdf := range sdfs {
			value := sdf.SDF(c)
			if value > maxSDF {
				maxSDF = value
				maxColorFn = colorFns[i]
			}
		}
		return maxColorFn(c)
	}
}

// JoinedMeshCoordColorFunc combines CoordColorFuncs for
// different meshes, using the color function of the mesh
// closest to a given point.
//
// This behaves similarly to JoinedCoordColorFunc, but will
// choose the closest surface rather than the object with
// the overall greatest SDF. This should only affect points
// contained inside of the union of all of the objects.
func JoinedMeshCoordColorFunc(meshToColorFunc map[*model3d.Mesh]interface{}) CoordColorFunc {
	allMeshes := model3d.NewMesh()
	triToColorFunc := map[*model3d.Triangle]CoordColorFunc{}
	for mesh, colorObj := range meshToColorFunc {
		colorFunc := colorFuncFromObj(colorObj)
		mesh.Iterate(func(t *model3d.Triangle) {
			// Note: if a triangle is present in multiple meshes,
			// one mesh's color func will end up owning the triangle.
			triToColorFunc[t] = colorFunc
			allMeshes.Add(t)
		})
	}
	// The mesh may not have a well-defined sign, since the different
	// meshes may overlap and disobey the even-odd rule. The resulting
	// SDF should still produce correct closest points and faces.
	faceSDF := model3d.MeshToSDF(allMeshes)
	return func(c model3d.Coord3D) render3d.Color {
		face, _, _ := faceSDF.FaceSDF(c)
		return triToColorFunc[face](c)
	}
}

func colorFuncFromObj(obj interface{}) CoordColorFunc {
	switch colorFn := obj.(type) {
	case CoordColorFunc:
		return colorFn
	case func(c model3d.Coord3D) render3d.Color:
		return colorFn
	case render3d.Color:
		return ConstantCoordColorFunc(colorFn)
	default:
		panic(fmt.Sprintf("unknown type for color: %T", colorFn))
	}
}
