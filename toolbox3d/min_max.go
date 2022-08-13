package toolbox3d

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/numerical"
)

// LineSearch is an extension of numerical.LineSearch with
// extra helper methods for concrete applications.
type LineSearch numerical.LineSearch

func (l *LineSearch) Minimize(min, max float64, f func(float64) float64) (x, fVal float64) {
	return (*numerical.LineSearch)(l).Minimize(min, max, f)
}

func (l *LineSearch) Maximize(min, max float64, f func(float64) float64) (x, fVal float64) {
	return (*numerical.LineSearch)(l).Maximize(min, max, f)
}

// LineSearch3D is a 3-dimensional recursive line search
// with extra helper methods for concrete applications.
type LineSearch3D numerical.RecursiveLineSearch[numerical.Vec3]

func (l *LineSearch3D) Minimize(min, max model3d.Coord3D,
	f func(model3d.Coord3D) float64) (x model3d.Coord3D, fVal float64) {
	v, y := (*numerical.RecursiveLineSearch[numerical.Vec3])(l).Minimize(min.Array(), max.Array(),
		func(v numerical.Vec3) float64 {
			return f(model3d.NewCoord3DArray(v))
		})
	return model3d.NewCoord3DArray(v), y
}

func (l *LineSearch3D) Maximize(min, max model3d.Coord3D,
	f func(model3d.Coord3D) float64) (x model3d.Coord3D, fVal float64) {
	v, y := (*numerical.RecursiveLineSearch[numerical.Vec3])(l).Maximize(min.Array(), max.Array(),
		func(v numerical.Vec3) float64 {
			return f(model3d.NewCoord3DArray(v))
		})
	return model3d.NewCoord3DArray(v), y
}

// SolidBounds estimates a tighter bounding rectangle for
// the solid using line searches to find the boundaries.
// The resulting bounds will always be slightly too tight,
// and may need to be expanded depending on the use case.
func (l *LineSearch3D) SolidBounds(solid model3d.Solid) (min, max model3d.Coord3D) {
	rs := (*numerical.RecursiveLineSearch[numerical.Vec3])(l)
	initMin := solid.Min().Array()
	initMax := solid.Max().Array()
	minMax := [2]numerical.Vec3{initMin, initMax}
	for axis := 0; axis < 3; axis++ {
		permFn := func(v numerical.Vec3) numerical.Vec3 {
			// Whichever axis we are searching for should be the
			// final axis of the line search. Otherwise we will
			// not fully leverage interior line searches.
			if axis < 2 {
				v[axis], v[2] = v[2], v[axis]
			}
			return v
		}
		for isMax := 0; isMax < 2; isMax++ {
			coord, value := rs.Maximize(
				permFn(initMin),
				permFn(initMax),
				func(v numerical.Vec3) float64 {
					v = permFn(v)
					if !solid.Contains(model3d.NewCoord3DArray(v)) {
						return math.Inf(-1)
					}
					result := v[axis]
					if isMax == 1 {
						return result
					} else {
						return -result
					}
				},
			)
			// Only change the bounds if an interior point was actually
			// found within the solid.
			if !math.IsInf(value, -1) {
				minMax[isMax][axis] = permFn(coord)[axis]
			}
		}
	}
	return model3d.NewCoord3DArray(minMax[0]), model3d.NewCoord3DArray(minMax[1])
}

// CurveBounds approximates the bounding box of a
// parametric curve, such as a 2D bezier curve.
//
// The min and max arguments specify the minimum and
// maximum argument to pass to f, which is typically in
// the range [0, 1] for Bezier curves.
func (l *LineSearch) CurveBounds(min, max float64, f func(float64) model2d.Coord) (model2d.Coord,
	model2d.Coord) {
	minArr := [2]float64{}
	maxArr := [2]float64{}
	for i := 0; i < 2; i++ {
		_, minArr[i] = l.Minimize(min, max, func(t float64) float64 {
			return f(t).Array()[i]
		})
		_, maxArr[i] = l.Maximize(min, max, func(t float64) float64 {
			return f(t).Array()[i]
		})
	}
	return model2d.NewCoordArray(minArr), model2d.NewCoordArray(maxArr)
}

// GridSearch2D is an extension of numerical.GridSearch2D
// with extra helper methods for concrete applications.
type GridSearch2D numerical.GridSearch2D

func (g *GridSearch2D) Minimize(min, max model2d.Coord,
	f func(model2d.Coord) float64) (model2d.Coord, float64) {
	x, y := (*numerical.GridSearch2D)(g).Minimize(min.Array(), max.Array(),
		func(c numerical.Vec2) float64 {
			return f(model2d.NewCoordArray(c))
		})
	return model2d.NewCoordArray(x), y
}

func (g *GridSearch2D) Maximize(min, max model2d.Coord,
	f func(model2d.Coord) float64) (model2d.Coord, float64) {
	x, y := (*numerical.GridSearch2D)(g).Maximize(min.Array(), max.Array(),
		func(c numerical.Vec2) float64 {
			return f(model2d.NewCoordArray(c))
		})
	return model2d.NewCoordArray(x), y
}

// MaxSDF finds the point with maximal SDF and returns it,
// along with the SDF value.
func (g *GridSearch2D) MaxSDF(s model2d.SDF) (model2d.Coord, float64) {
	return g.Maximize(s.Min(), s.Max(), s.SDF)
}

// GridSearch3D is an extension of numerical.GridSearch3D
// with extra helper methods for concrete applications.
type GridSearch3D numerical.GridSearch3D

func (g *GridSearch3D) Minimize(min, max model3d.Coord3D,
	f func(model3d.Coord3D) float64) (model3d.Coord3D, float64) {
	x, y := (*numerical.GridSearch3D)(g).Minimize(min.Array(), max.Array(),
		func(c numerical.Vec3) float64 {
			return f(model3d.NewCoord3DArray(c))
		})
	return model3d.NewCoord3DArray(x), y
}

func (g *GridSearch3D) Maximize(min, max model3d.Coord3D,
	f func(model3d.Coord3D) float64) (model3d.Coord3D, float64) {
	x, y := (*numerical.GridSearch3D)(g).Maximize(min.Array(), max.Array(),
		func(c numerical.Vec3) float64 {
			return f(model3d.NewCoord3DArray(c))
		})
	return model3d.NewCoord3DArray(x), y
}

// MaxSDF finds the point with maximal SDF and returns it,
// along with the SDF value.
func (g *GridSearch3D) MaxSDF(s model3d.SDF) (model3d.Coord3D, float64) {
	return g.Maximize(s.Min(), s.Max(), s.SDF)
}
