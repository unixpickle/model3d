package model2d

import (
	"math"
	"sort"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/numerical"
)

const (
	DefaultBezierFitterNumIters  = 100
	DefaultBezierFitTolerance    = 1e-8
	DefaultBezierFitDelta        = 1e-5
	DefaultBezierFitMinStepScale = 1e-2
	DefaultBezierFitLineStep     = 2.0
	DefaultBezierFitLineGSS      = 8
)

// A BezierFitter fits Bezier curves to points.
type BezierFitter struct {
	// NumIters is the number of gradient steps to use to
	// fit each Bezier curve.
	// If 0, DefaultBezierFitterNumIters is used.
	NumIters int

	// Tolerance is the maximum mean-squared error for a
	// curve when fitting a chain of connected points.
	// It is relative to the area of the bounding box of
	// the points in the chain.
	//
	// If 0, DefaultBezierFitTolerance is used.
	Tolerance float64

	// Delta, if specified, controls the step size used
	// for finite differences, relative to the size of the
	// entire Bezier curve.
	// If 0, DefaultBezierFitDelta is used.
	Delta float64

	// L2Penalty, if specified, is a loss penalty imposed
	// on the squared distance between the control points
	// and their corresponding endpoints, scaled relative
	// to the distance between the endpoints.
	L2Penalty float64

	// MinStepScale, if specified, is a scalar multiplied
	// by the finite-differences delta to decide the first
	// (and smallest) step to try for line search.
	// If 0, DefaultBezierFitMinStepScale is used.
	MinStepScale float64

	// LineStep, if specified, is the rate of increase for
	// each step of line search. Larger values make line
	// search faster, but can miss local minima.
	// If 0, DefaultBezierFitLineStep is used.
	LineStep float64

	// LineGSS, if specified, is the number of steps used
	// for golden section search at the end of line
	// search. Higher values yield more precise steps.
	// If 0, DefaultBezierFitLineGSS is used.
	LineGSS int

	// Momentum, if specified, is the momentum coefficient.
	// If 0, regular gradient descent is used.
	Momentum float64
}

// Fit fits a collection of cubic Bezier curves to a
// manifold mesh.
func (b *BezierFitter) Fit(m *Mesh) []BezierCurve {
	var res []BezierCurve
	for _, hier := range MeshToHierarchy(m) {
		res = append(res, b.hierarchyToCurves(hier)...)
	}
	return res
}

func (b *BezierFitter) hierarchyToCurves(m *MeshHierarchy) []BezierCurve {
	segs := m.Mesh.SegmentSlice()
	if len(segs) == 0 {
		return nil
	}
	seg := segs[0]
	points := make([]Coord, 0, len(segs)+1)
	points = append(points, seg[0], seg[1])
	m.Mesh.Remove(seg)
	for i := 1; i < len(segs); i++ {
		next := m.Mesh.Find(seg[1])
		if len(next) != 1 {
			panic("mesh is non-manifold")
		}
		seg = next[0]
		m.Mesh.Remove(seg)
		points = append(points, seg[1])
	}
	res := b.FitChain(points, true)
	for _, child := range m.Children {
		res = append(res, b.hierarchyToCurves(child)...)
	}
	return res
}

// FitChain fits a sequence of points along some curve
// using one or more cubic Bezier curves.
//
// The closed argument indicates if the final point should
// be smoothly reconnected to the first point.
//
// The points should be ordered along the desired curve.
func (b *BezierFitter) FitChain(points []Coord, closed bool) []BezierCurve {
	if len(points) <= 4 {
		curve := b.FitCubic(points, nil)
		if closed {
			t1 := curve[3].Sub(curve[2])
			t2 := curve[0].Sub(curve[1])
			return []BezierCurve{
				curve,
				b.FitCubicConstrained([]Coord{points[len(points)-1], points[0]}, &t1, &t2, nil),
			}
		} else {
			return []BezierCurve{curve}
		}
	}

	tol := b.Tolerance
	if tol == 0 {
		tol = DefaultBezierFitTolerance
	}
	min, max := points[0], points[0]
	for _, p := range points[1:] {
		min = min.Min(p)
		max = max.Max(p)
	}
	tol *= (max.Y - min.Y) * (max.X - min.Y)

	if closed {
		points = append(append([]Coord{}, points...), points[0])
	}

	curves := []BezierCurve{}
	start := 0
	var constraint1 *Coord
	for start+1 < len(points) {
		var lastTry BezierCurve
		var lastSize int
		size := essentials.MinInt(4, len(points)-start)
		for {
			includeEnd := false
			if start+size >= len(points) {
				size = len(points) - start
				includeEnd = true
			}
			var constraint2 *Coord
			p := points[start : start+size]
			if includeEnd && closed {
				if len(curves) == 0 {
					// We will not attempt to fit the whole loop
					// with one curve.
					break
				}
				constraint2 = new(Coord)
				*constraint2 = curves[0][0].Sub(curves[0][1])
			}
			fit := b.FitCubicConstrained(p, constraint1, constraint2, b.FirstGuess(p))
			if lastTry == nil {
				lastTry = fit
				lastSize = size
			}
			if b.MSE(p, fit) > tol {
				break
			}
			lastTry = fit
			lastSize = size
			if includeEnd {
				break
			}
			size *= 2
		}
		curves = append(curves, lastTry)
		start += lastSize - 1
		constraint1 = new(Coord)
		*constraint1 = lastTry[3].Sub(lastTry[2])
	}
	return curves
}

// FitCubic finds the cubic Bezier curve of best fit for
// the points.
//
// The first and last points are used as start and end
// points, and all the other points may be in any order.
func (b *BezierFitter) FitCubic(points []Coord, start BezierCurve) BezierCurve {
	return b.FitCubicConstrained(points, nil, nil, start)
}

// FitCubicConstrained is like FitCubic, but constrains
// the tangent vectors at either the start or end point,
// or both.
//
// If non-nil, t1 is the direction for the first control
// point, and t2 is the direction for the final one.
func (b *BezierFitter) FitCubicConstrained(points []Coord, t1, t2 *Coord,
	start BezierCurve) BezierCurve {
	if len(points) == 0 {
		panic("at least one point is required")
	} else if len(points) == 1 {
		return BezierCurve{points[0], points[0], points[0], points[0]}
	} else if len(points) == 2 {
		res := BezierCurve{points[0], points[0], points[1], points[1]}
		delta := 0.1 * points[0].Dist(points[1])
		if t1 != nil {
			res[1] = points[0].Add((*t1).Normalize().Scale(delta))
		}
		if t2 != nil {
			res[2] = points[1].Add((*t2).Normalize().Scale(delta))
		}
		return res
	}
	if start == nil {
		dir := points[len(points)-1].Sub(points[0])
		start = BezierCurve{
			points[0],
			points[0].Add(dir.Scale(1.0 / 3.0)),
			points[0].Add(dir.Scale(2.0 / 3.0)),
			points[len(points)-1],
		}
	} else if t1 != nil || t2 != nil {
		// Make sure the solution has the correct normals by projecting
		// the first guess.
		start = append(BezierCurve{}, start...)
		if t1 != nil {
			d := (*t1).Normalize()
			start[1] = start[0].Add(d.Scale(d.Dot(start[1].Sub(start[0]))))
		}
		if t2 != nil {
			d := (*t2).Normalize()
			start[2] = start[3].Add(d.Scale(d.Dot(start[2].Sub(start[3]))))
		}
	}
	var momentum BezierCurve
	for i := 0; i < b.numIters(); i++ {
		var grad BezierCurve
		if b.Momentum == 0 {
			grad = b.gradient(points, t1, t2, start)
		} else {
			grad = b.gradient(points, t1, t2, start)
			if momentum == nil {
				momentum = grad
			} else {
				for i, x := range grad {
					momentum[i] = momentum[i].Scale(b.Momentum).Add(x)
				}
				grad = momentum
			}
		}
		newStart := b.lineSearch(points, start, grad)
		if newStart == nil {
			break
		}
		start = newStart
	}
	return start
}

// FirstGuess attempts to quickly approximate some subset
// of the specified points with a cubic Bezier curve,
// allowing for potentially faster convergence when
// fitting all of the points.
//
// This method assumes that all of the points are sorted
// along the curve from a start point to an end point.
// This is a stronger assumption than FitCubic() makes.
func (b *BezierFitter) FirstGuess(points []Coord) BezierCurve {
	if len(points) < 4 {
		return b.FitCubic(points, nil)
	}

	lengths := make([]float64, len(points))
	cur := 0.0
	last := points[0]
	for i, x := range points {
		if i > 0 {
			cur += x.Dist(last)
		}
		lengths[i] = cur
	}

	interp := func(t float64) Coord {
		idx := sort.SearchFloat64s(lengths, cur*t)
		if idx > len(points)-1 {
			return points[len(points)-1]
		} else if idx == 0 {
			idx = 1
		}
		t1 := lengths[idx-1] / cur
		t2 := lengths[idx] / cur
		frac2 := (t - t1) / (t2 - t1)
		return points[idx-1].Scale(1 - frac2).Add(points[idx].Scale(frac2))
	}

	// Approximate normals using a small fraction of the curve.
	n1 := interp(0.01).Sub(points[0]).Normalize()
	n2 := points[len(points)-1].Sub(interp(0.99)).Normalize()

	// Fit third-way points.
	p1 := interp(1.0 / 3.0)
	p2 := interp(2.0 / 3.0)

	return b.FitCubicConstrained([]Coord{points[0], p1, p2, points[len(points)-1]}, &n1, &n2, nil)
}

// lineSearch finds the (locally optimal) best step size
// and returns the result of the step.
//
// If the optimal step is 0, nil is returned.
func (b *BezierFitter) lineSearch(points []Coord, curve, grad BezierCurve) BezierCurve {
	delta := b.delta(curve)
	maxNorm := 0.0
	for _, g := range grad {
		maxNorm = math.Max(maxNorm, g.Norm())
	}
	minStep := delta / maxNorm
	if b.MinStepScale != 0 {
		minStep *= b.MinStepScale
	} else {
		minStep *= DefaultBezierFitMinStepScale
	}

	curveForStep := func(s float64) BezierCurve {
		c1 := append(BezierCurve{}, curve...)
		for i, x := range c1 {
			c1[i] = x.Add(grad[i].Scale(-s))
		}
		return c1
	}
	evalStep := func(s float64) float64 {
		return b.loss(points, curveForStep(s))
	}

	lastGuesses := [2]float64{0, minStep}
	lastValues := [2]float64{b.loss(points, curve), evalStep(minStep)}

	if lastValues[1] > lastValues[0] {
		return nil
	}
	step := b.LineStep
	if step == 0 {
		step = DefaultBezierFitLineStep
	}
	for i := 0; i < 64; i++ {
		x := lastGuesses[1] * step
		y := evalStep(x)
		if y > lastValues[1] {
			gss := b.LineGSS
			if gss == 0 {
				gss = DefaultBezierFitLineGSS
			}
			s := numerical.GSS(lastGuesses[0], x, gss, evalStep)
			return curveForStep(s)
		}
		lastGuesses[0], lastGuesses[1] = lastGuesses[1], x
		lastValues[0], lastValues[1] = lastValues[1], y
	}
	return curveForStep(lastGuesses[1])
}

// gradient computes the gradient of the loss wrt the
// curve control points.
// The endpoint gradients are set to zero.
func (b *BezierFitter) gradient(points []Coord, t1, t2 *Coord, curve BezierCurve) BezierCurve {
	delta := b.delta(curve)

	c1 := append(BezierCurve{}, curve...)
	grad := make(BezierCurve, len(curve))
	tangents := []*Coord{t1, t2}
	for i := 1; i < len(curve)-1; i++ {
		tangent := tangents[i-1]
		if tangent != nil {
			v := tangent.Normalize()
			c1[i] = curve[i].Add(v.Scale(delta))
			loss1 := b.MSE(points, c1)
			c1[i] = curve[i].Add(v.Scale(-delta))
			loss2 := b.MSE(points, c1)
			grad[i] = v.Scale((loss1 - loss2) / (2 * delta))
		} else {
			pArr := curve[i].Array()
			gradArr := [2]float64{}
			for axis := 0; axis < 2; axis++ {
				newPArr := pArr
				newPArr[axis] += delta
				c1[i] = NewCoordArray(newPArr)
				loss1 := b.MSE(points, c1)
				newPArr[axis] -= 2 * delta
				c1[i] = NewCoordArray(newPArr)
				loss2 := b.MSE(points, c1)
				gradArr[axis] = (loss1 - loss2) / (2 * delta)
			}
			c1[i] = curve[i]
			grad[i] = NewCoordArray(gradArr)
		}
	}

	if b.L2Penalty != 0 {
		rel := curve[0].SquaredDist(curve[3])
		grad[1] = grad[1].Add(curve[1].Sub(curve[0]).Scale(2 * rel * b.L2Penalty))
		grad[2] = grad[2].Add(curve[2].Sub(curve[3]).Scale(2 * rel * b.L2Penalty))
	}

	return grad
}

func (b *BezierFitter) loss(points []Coord, curve BezierCurve) float64 {
	res := b.MSE(points, curve)
	if b.L2Penalty != 0 {
		rel := curve[0].SquaredDist(curve[3])
		res += b.L2Penalty * rel * (curve[1].SquaredDist(curve[0]) + curve[2].SquaredDist(curve[3]))
	}
	return res
}

// MSE computes the MSE of a cubic Bezier fit.
func (b *BezierFitter) MSE(points []Coord, curve BezierCurve) float64 {
	if len(curve) != 4 {
		panic("curve must be a cubic Bezier curve (i.e. have four points)")
	}
	axisPolynomial := func(axis int) numerical.Polynomial {
		// Expand (1-t)^3*w + 3(1-t)^2t*x + 3(1-t)t^2y + t^3z =>
		// t^3 (-w) + 3 t^3 x - 3 t^3 y + t^3 z + 3 t^2 w - 6 t^2 x + 3 t^2 y - 3 t w + 3 t x + w
		//
		// t^3 coeff: (-3w + 3x - 3y + z)
		// t^2 coeff: (9w - 6x + 3y)
		// t coeff: -9w + 3x
		// bias: 3w
		w := curve[0].Array()[axis]
		x := curve[1].Array()[axis]
		y := curve[2].Array()[axis]
		z := curve[3].Array()[axis]
		return numerical.Polynomial{
			w, 3*x - 3*w, 3*y - 6*x + 3*w, z - 3*y + 3*x - w,
		}
	}
	total := 0.0
	for _, p := range points {
		// Project p onto the curve by finding the closest point.
		px := axisPolynomial(0)
		px[0] -= p.X
		py := axisPolynomial(1)
		py[0] -= p.Y
		lossPoly := px.Mul(px).Add(py.Mul(py))
		minLoss := math.Inf(1)
		for _, t := range append(lossPoly.Derivative().RealRoots(), 0.0, 1.0) {
			if t < 0 || t > 1 {
				// Roots may lie outside of the Bezier t bound.
				continue
			}
			loss := lossPoly.Eval(t)
			if loss < minLoss {
				minLoss = loss
			}
		}
		total += minLoss
	}
	return total / float64(len(points))
}

func (b *BezierFitter) delta(c BezierCurve) float64 {
	min, max := c[0], c[0]
	for _, x := range c {
		min = min.Min(x)
		max = max.Max(x)
	}
	if b.Delta != 0 {
		return max.Dist(min) * b.Delta
	} else {
		return max.Dist(min) * DefaultBezierFitDelta
	}
}

func (b *BezierFitter) numIters() int {
	if b.NumIters == 0 {
		return DefaultBezierFitterNumIters
	}
	return b.NumIters
}
