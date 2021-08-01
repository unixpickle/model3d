package model2d

import (
	"math"
	"math/rand"
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

	// AllowIntersections can be set to true to allow
	// Bezier curves to cross themselves.
	AllowIntersections bool
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
	seg := segs[rand.Intn(len(segs))]
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
	res := b.FitChain(points[:len(points)-1], true)
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
// The points should be ordered along the desired curve,
// and no two points should be exactly equal.
func (b *BezierFitter) FitChain(points []Coord, closed bool) []BezierCurve {
	if len(points) <= 4 {
		curve := b.FitCubic(points, nil)
		if closed {
			t2, t1 := bezierTangents(curve, -1)
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
				*constraint2, _ = bezierTangents(curves[0], -1)
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
		_, *constraint1 = bezierTangents(lastTry, -1)
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
// point from the first point, and t2 is the direction of
// the second control point from the last point.
func (b *BezierFitter) FitCubicConstrained(points []Coord, t1, t2 *Coord,
	start BezierCurve) BezierCurve {
	if len(points) == 0 {
		panic("at least one point is required")
	} else if len(points) == 1 {
		return BezierCurve{points[0], points[0], points[0], points[0]}
	}

	tc := newBezierTangentConstraints(t1, t2)

	if len(points) == 2 || start == nil {
		delta := 0.1 * points[0].Dist(points[len(points)-1])
		d1 := points[len(points)-1].Sub(points[0]).Normalize()
		d2 := d1.Scale(-1)
		if tc.F[0] {
			d1 = tc.V[0]
		}
		if tc.F[1] {
			d2 = tc.V[1]
		}
		start = BezierCurve{
			points[0],
			points[0].Add(d1.Scale(delta)),
			points[len(points)-1].Add(d2.Scale(delta)),
			points[len(points)-1],
		}
		if len(points) == 2 {
			return start
		}
	} else {
		start = tc.Constrain(start)
		if !b.AllowIntersections {
			// Even if the user-passed initial guess didn't self-intersect,
			// projecting it onto t1 and t2 could have caused intersections.
			start = bezierRemoveIntersections(start)
		}
	}

	var momentum numerical.Vec
	for i := 0; i < b.numIters(); i++ {
		var grad numerical.Vec
		if b.Momentum == 0 {
			grad = b.gradient(points, start, tc)
		} else {
			grad = b.gradient(points, start, tc)
			if momentum == nil {
				momentum = grad
			} else {
				grad = momentum.Scale(b.Momentum).Add(grad)
				momentum = grad
			}
		}
		constraint := newBezierStepConstraint(start, grad, tc, !b.AllowIntersections)
		newStart := b.lineSearch(points, start, grad.Norm(), constraint)
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
// along the curve from a start point to an end point, and
// no two points are exactly equal.
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
	n2 := interp(0.99).Sub(points[len(points)-1]).Normalize()

	// Fit third-way points.
	p1 := interp(1.0 / 3.0)
	p2 := interp(2.0 / 3.0)

	return b.FitCubicConstrained([]Coord{points[0], p1, p2, points[len(points)-1]}, &n1, &n2, nil)
}

// lineSearch finds the (locally optimal) best step size
// and returns the result of the step.
//
// If the optimal step is 0, nil is returned.
func (b *BezierFitter) lineSearch(points []Coord, curve BezierCurve, gradScale float64,
	constraint *bezierStepConstraint) BezierCurve {
	delta := b.delta(curve)
	minStep := delta / gradScale
	if b.MinStepScale != 0 {
		minStep *= b.MinStepScale
	} else {
		minStep *= DefaultBezierFitMinStepScale
	}

	if minStep > constraint.MaxStep {
		return nil
	}

	evalStep := func(s float64) float64 {
		return b.loss(points, constraint.Step(s))
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
	for i := 0; i < 64 && lastGuesses[1] < constraint.MaxStep; i++ {
		x := math.Min(lastGuesses[1]*step, constraint.MaxStep)
		y := evalStep(x)
		if y > lastValues[1] {
			gss := b.LineGSS
			if gss == 0 {
				gss = DefaultBezierFitLineGSS
			}
			s := numerical.GSS(lastGuesses[0], x, gss, evalStep)
			return constraint.Step(s)
		}
		lastGuesses[0], lastGuesses[1] = lastGuesses[1], x
		lastValues[0], lastValues[1] = lastValues[1], y
	}
	return constraint.Step(lastGuesses[1])
}

// gradient computes the gradient of the loss wrt the
// curve control points.
// The endpoint gradients are set to zero.
func (b *BezierFitter) gradient(points []Coord, curve BezierCurve,
	tc *bezierTangentConstraints) numerical.Vec {
	delta := b.delta(curve)

	grad := make(numerical.Vec, tc.Dim())
	fv := make(numerical.Vec, tc.Dim())
	for i := 0; i < tc.Dim(); i++ {
		fv[i] = delta
		loss1 := b.MSE(points, addBeziers(curve, tc.Expand(fv), 1.0, 1.0))
		loss2 := b.MSE(points, addBeziers(curve, tc.Expand(fv), 1.0, -1.0))
		grad[i] = (loss1 - loss2) / (2 * delta)
		fv[i] = 0
	}

	if b.L2Penalty != 0 {
		rel := curve[0].SquaredDist(curve[3])
		g1 := tc.LinGrad(BezierCurve{
			Coord{},
			curve[1].Sub(curve[0]).Scale(2 * rel * b.L2Penalty),
			curve[2].Sub(curve[3]).Scale(2 * rel * b.L2Penalty),
			Coord{},
		})
		grad = grad.Add(g1)
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

type bezierTangentConstraints struct {
	// Normalized tangent vectors.
	V [2]Coord

	// Flags set to true if tangent vector is constrained.
	F [2]bool
}

func newBezierTangentConstraints(t1, t2 *Coord) *bezierTangentConstraints {
	res := &bezierTangentConstraints{}
	for i, t := range []*Coord{t1, t2} {
		if t != nil {
			res.V[i] = (*t).Normalize()
			res.F[i] = true
		}
	}
	return res
}

// Dim returns the number of free variables.
func (b *bezierTangentConstraints) Dim() int {
	res := 2
	for _, f := range b.F {
		if !f {
			res++
		}
	}
	return res
}

// Expand fills the entries of a Bezier gradient with the
// free variables.
func (b *bezierTangentConstraints) Expand(free numerical.Vec) BezierCurve {
	res := make(BezierCurve, 4)
	for i, v := range b.V {
		if b.F[i] {
			res[i+1] = v.Scale(free[0])
			free = free[1:]
		} else {
			res[i+1] = XY(free[0], free[1])
			free = free[2:]
		}
	}
	if len(free) != 0 {
		panic("incorrect number of free variables passed")
	}
	return res
}

// Constrain projects a Bezier curve onto the space of
// allowed curves.
func (b *bezierTangentConstraints) Constrain(c BezierCurve) BezierCurve {
	if !b.F[0] && !b.F[1] {
		return c
	}

	c = append(BezierCurve{}, c...)
	if b.F[0] {
		v := b.V[0]
		c[1] = c[0].Add(v.Scale(v.Dot(c[1].Sub(c[0]))))
	}
	if b.F[1] {
		v := b.V[1]
		c[2] = c[3].Add(v.Scale(v.Dot(c[2].Sub(c[3]))))
	}
	return c
}

// LinGrad computes the gradient of a linear objective in
// terms of the free variables.
func (b *bezierTangentConstraints) LinGrad(g BezierCurve) numerical.Vec {
	var res numerical.Vec
	for i, c := range g[1:3] {
		if b.F[i] {
			res = append(res, b.V[i].Dot(c))
		} else {
			res = append(res, c.X, c.Y)
		}
	}
	return res
}

type bezierStepConstraint struct {
	MaxStep float64

	curve    BezierCurve
	grad     BezierCurve
	projStep float64
	projGrad BezierCurve
}

func newBezierStepConstraint(c BezierCurve, grad numerical.Vec, tc *bezierTangentConstraints,
	on bool) *bezierStepConstraint {
	if !on {
		g := tc.Expand(grad)
		return &bezierStepConstraint{
			MaxStep:  math.Inf(1),
			curve:    c,
			grad:     g,
			projStep: math.Inf(1),
			projGrad: g,
		}
	}

	v := c[3].Sub(c[0]).Normalize()
	separateDir := tc.LinGrad(BezierCurve{
		Coord{},
		v.Scale(-1),
		v.Scale(1),
		Coord{},
	})

	dot := separateDir.Dot(grad)

	if dot < 0 {
		// The negative gradient already pulls the points apart.
		g := tc.Expand(grad)
		return &bezierStepConstraint{
			MaxStep:  math.Inf(1),
			curve:    c,
			grad:     g,
			projStep: math.Inf(1),
			projGrad: g,
		}
	}

	curSeparation := c[2].Dot(v) - c[1].Dot(v)
	maxStep := math.Max(0, curSeparation/dot)

	projStep := grad.ProjectOut(separateDir)
	if projStep.Norm() < 1e-5*grad.Norm() {
		// For numerical reasons, we will ignore a small gradient
		// pointing orthogonal to the direction of separation.
		return &bezierStepConstraint{
			MaxStep:  maxStep,
			curve:    c,
			grad:     tc.Expand(grad),
			projStep: maxStep,
			projGrad: make(BezierCurve, 4),
		}
	}

	return &bezierStepConstraint{
		MaxStep:  math.Inf(1),
		curve:    c,
		grad:     tc.Expand(grad),
		projStep: maxStep,
		projGrad: tc.Expand(projStep),
	}
}

func (b *bezierStepConstraint) Step(size float64) BezierCurve {
	if size > b.MaxStep {
		size = b.MaxStep
	}
	if size <= b.projStep {
		return addBeziers(b.curve, b.grad, 1.0, -size)
	}
	return addBeziers(
		addBeziers(b.curve, b.grad, 1.0, -b.projStep),
		b.projGrad, 1.0, b.projStep-size,
	)
}

func bezierRemoveIntersections(c BezierCurve) BezierCurve {
	v := c[3].Sub(c[0]).Normalize()
	v1 := c[1].Sub(c[0])
	v2 := c[2].Sub(c[3])
	if v.Dot(c[1]) <= v.Dot(c[2]) {
		return c
	}
	// Solve for t in v*(c[0]+v1*t) = v*(c[3]+v2*t)
	// v*c[0]+v*v1*t = v*c[3]+v*v2*t
	// v*c[0]-v*c[3] = v*v2*t - v*v1*t
	// (v*c[0]-v*c[3]) / (v*v2 - v*v1) = t
	t := (v.Dot(c[0]) - v.Dot(c[3])) / (v.Dot(v2) - v.Dot(v1))
	return BezierCurve{
		c[0],
		c[0].Add(v1.Scale(t)),
		c[3].Add(v2.Scale(t)),
		c[3],
	}
}

func bezierTangents(b BezierCurve, scale float64) (n1, n2 Coord) {
	n1 = b[1].Sub(b[0])
	n2 = b[2].Sub(b[3])
	if n1.Norm() == 0 {
		n1 = b[2].Sub(b[0])
		if n1.Norm() == 0 {
			n1 = b[3].Sub(b[0])
		}
	}
	if n2.Norm() == 0 {
		n2 = b[1].Sub(b[3])
		if n2.Norm() == 0 {
			n2 = b[0].Sub(b[3])
		}
	}
	return n1.Normalize().Scale(scale), n2.Normalize().Scale(scale)
}

func addBeziers(c1, c2 BezierCurve, s1, s2 float64) BezierCurve {
	res := make(BezierCurve, len(c1))
	for i, x := range c1 {
		y := c2[i]
		res[i] = x.Scale(s1).Add(y.Scale(s2))
	}
	return res
}
