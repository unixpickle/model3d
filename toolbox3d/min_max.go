package toolbox3d

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
)

// A LineSearch implements a 1D line search for minimizing
// or maximizing 1D functions.
type LineSearch struct {
	// Number of points to test along the input space.
	Stops int

	// Number of times to recursively search around the
	// best point on the line.
	Recursions int
}

func (l *LineSearch) Minimize(min, max float64, f func(float64) float64) (x, fVal float64) {
	x, fVal = l.Maximize(min, max, func(x float64) float64 {
		return -f(x)
	})
	return x, -fVal
}

func (l *LineSearch) Maximize(min, max float64, f func(float64) float64) (x, fVal float64) {
	return l.maximize(min, max, f, l.Recursions)
}

func (l *LineSearch) maximize(min, max float64, f func(float64) float64,
	recursions int) (float64, float64) {
	if recursions < 0 {
		panic("number of recursions cannot be negative")
	}
	size := max - min
	xStep := size / float64(l.Stops)
	solution := 0.0
	value := math.Inf(-1)
	for xi := 0; xi < l.Stops; xi++ {
		x := float64(xi)*xStep + xStep/2 + min
		v := f(x)
		if v > value {
			value = v
			solution = x
		}
	}
	if recursions == 0 {
		return solution, value
	}
	newMin := math.Max(min, solution-xStep)
	newMax := math.Min(max, solution+xStep)
	return l.maximize(newMin, newMax, f, recursions-1)
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

// A GridSearch2D implements 2D grid search for minimizing
// or maximizing 2D functions.
type GridSearch2D struct {
	// Number of points to test along x and y axes.
	XStops int
	YStops int

	// Number of times to recursively search around the
	// best point in the grid.
	Recursions int
}

// Minimize uses grid search to find the minimum value of
// f within the given bounds.
//
// Returns the point and its function value.
func (g *GridSearch2D) Minimize(min, max model2d.Coord,
	f func(model2d.Coord) float64) (model2d.Coord, float64) {
	c, v := g.Maximize(min, max, func(c model2d.Coord) float64 {
		return -f(c)
	})
	return c, -v
}

// Maximize uses grid search to find the maximum value of
// f within the given bounds.
//
// Returns the point and its function value.
func (g *GridSearch2D) Maximize(min, max model2d.Coord,
	f func(model2d.Coord) float64) (model2d.Coord, float64) {
	return g.maximize(min, max, f, g.Recursions)
}

func (g *GridSearch2D) maximize(min, max model2d.Coord,
	f func(model2d.Coord) float64, recursions int) (model2d.Coord, float64) {
	if recursions < 0 {
		panic("number of recursions cannot be negative")
	}
	size := max.Sub(min)
	xStep := size.X / float64(g.XStops)
	yStep := size.Y / float64(g.YStops)
	solution := model2d.Coord{}
	value := math.Inf(-1)
	for xi := 0; xi < g.XStops; xi++ {
		x := float64(xi)*xStep + xStep/2 + min.X
		for yi := 0; yi < g.YStops; yi++ {
			y := float64(yi)*yStep + yStep/2 + min.Y
			c := model2d.XY(x, y)
			v := f(c)
			if v > value {
				value = v
				solution = c
			}
		}
	}
	if recursions == 0 {
		return solution, value
	}
	newMin := min.Max(solution.Sub(model2d.XY(xStep, yStep)))
	newMax := max.Min(solution.Add(model2d.XY(xStep, yStep)))
	return g.maximize(newMin, newMax, f, recursions-1)
}

// MaxSDF finds the point with maximal SDF and returns it,
// along with the SDF value.
func (g *GridSearch2D) MaxSDF(s model2d.SDF) (model2d.Coord, float64) {
	return g.Maximize(s.Min(), s.Max(), s.SDF)
}
