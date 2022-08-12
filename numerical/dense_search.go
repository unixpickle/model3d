package numerical

import (
	"math"
)

// A LineSearch implements a 1D line search for minimizing
// or maximizing 1D functions.
//
// The LineSearch is performed by first sampling coarsely
// along the function, and then iteratively sampling more
// densely around the current optimal solution.
type LineSearch struct {
	// Number of points to test along the input space.
	Stops int

	// Number of times to recursively search around the
	// best point on the line.
	Recursions int
}

// Minimize computes x and f(x) for x in [min, max] to
// minimize f(x).
func (l *LineSearch) Minimize(min, max float64, f func(float64) float64) (x, fVal float64) {
	x, fVal = l.Maximize(min, max, func(x float64) float64 {
		return -f(x)
	})
	return x, -fVal
}

// Maximize computes x and f(x) for x in [min, max] to
// maximize f(x).
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
func (g *GridSearch2D) Minimize(min, max Vec2,
	f func(Vec2) float64) (Vec2, float64) {
	c, v := g.Maximize(min, max, func(c Vec2) float64 {
		return -f(c)
	})
	return c, -v
}

// Maximize uses grid search to find the maximum value of
// f within the given bounds.
//
// Returns the point and its function value.
func (g *GridSearch2D) Maximize(min, max Vec2, f func(Vec2) float64) (Vec2, float64) {
	return g.maximize(min, max, f, g.Recursions)
}

func (g *GridSearch2D) maximize(min, max Vec2, f func(Vec2) float64, recursions int) (Vec2,
	float64) {
	if recursions < 0 {
		panic("number of recursions cannot be negative")
	}
	size := max.Sub(min)
	xStep := size[0] / float64(g.XStops)
	yStep := size[1] / float64(g.YStops)
	solution := Vec2{}
	value := math.Inf(-1)
	for xi := 0; xi < g.XStops; xi++ {
		x := float64(xi)*xStep + xStep/2 + min[0]
		for yi := 0; yi < g.YStops; yi++ {
			y := float64(yi)*yStep + yStep/2 + min[1]
			c := Vec2{x, y}
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
	newMin := min.Max(solution.Sub(Vec2{xStep, yStep}))
	newMax := max.Min(solution.Add(Vec2{xStep, yStep}))
	return g.maximize(newMin, newMax, f, recursions-1)
}

// A GridSearch3D is like GridSearch2D, but for searching
// over a 3D volume instead of a 2D area.
type GridSearch3D struct {
	// Number of points to test along x, y, and z axes.
	XStops int
	YStops int
	ZStops int

	// Number of times to recursively search around the
	// best point in the grid.
	Recursions int
}

// Minimize uses grid search to find the minimum value of
// f within the given bounds.
//
// Returns the point and its function value.
func (g *GridSearch3D) Minimize(min, max Vec3,
	f func(Vec3) float64) (Vec3, float64) {
	c, v := g.Maximize(min, max, func(c Vec3) float64 {
		return -f(c)
	})
	return c, -v
}

// Maximize uses grid search to find the maximum value of
// f within the given bounds.
//
// Returns the point and its function value.
func (g *GridSearch3D) Maximize(min, max Vec3, f func(Vec3) float64) (Vec3, float64) {
	return g.maximize(min, max, f, g.Recursions)
}

func (g *GridSearch3D) maximize(min, max Vec3, f func(Vec3) float64, recursions int) (Vec3,
	float64) {
	if recursions < 0 {
		panic("number of recursions cannot be negative")
	}
	size := max.Sub(min)
	xStep := size[0] / float64(g.XStops)
	yStep := size[1] / float64(g.YStops)
	zStep := size[2] / float64(g.ZStops)
	solution := Vec3{}
	value := math.Inf(-1)
	for xi := 0; xi < g.XStops; xi++ {
		x := float64(xi)*xStep + xStep/2 + min[0]
		for yi := 0; yi < g.YStops; yi++ {
			y := float64(yi)*yStep + yStep/2 + min[1]
			for zi := 0; zi < g.ZStops; zi++ {
				z := float64(zi)*zStep + zStep/2 + min[2]
				c := Vec3{x, y, z}
				v := f(c)
				if v > value {
					value = v
					solution = c
				}
			}
		}
	}
	if recursions == 0 {
		return solution, value
	}
	delta := Vec3{xStep, yStep, zStep}
	newMin := min.Max(solution.Sub(delta))
	newMax := max.Min(solution.Add(delta))
	return g.maximize(newMin, newMax, f, recursions-1)
}
