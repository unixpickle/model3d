package model2d

import (
	"math"
	"sort"
	"sync"

	"github.com/unixpickle/model3d/numerical"
)

// DefaultBezierMaxSplits determines the maximum number of
// subdivisions when computing Bezier arc lengths.
const DefaultBezierMaxSplits = 16

var binomialCoeffs = [][]float64{
	{1, 1},
	{1, 2, 1},
	{1, 3, 3, 1},
	{1, 4, 6, 4, 1},
	{1, 5, 10, 10, 5, 1},
	{1, 6, 15, 20, 15, 6, 1},
	{1, 7, 21, 35, 35, 21, 7, 1},
	{1, 8, 28, 56, 70, 56, 28, 8, 1},
	{1, 9, 36, 84, 126, 126, 84, 36, 9, 1},
	{1, 10, 45, 120, 210, 252, 210, 120, 45, 10, 1},
	{1, 11, 55, 165, 330, 462, 462, 330, 165, 55, 11, 1},
	{1, 12, 66, 220, 495, 792, 924, 792, 495, 220, 66, 12, 1},
	{1, 13, 78, 286, 715, 1287, 1716, 1716, 1287, 715, 286, 78, 13, 1},
	{1, 14, 91, 364, 1001, 2002, 3003, 3432, 3003, 2002, 1001, 364, 91, 14, 1},
}

// A Curve is a parametric curve that returns points for
// values of t in the range [0, 1].
type Curve interface {
	Eval(t float64) Coord
}

// An ArcLenCurve is a curve with an ArcLen() method to get the
// approximate length of the curve.
type ArcLenCurve interface {
	Curve
	ArcLen() float64
}

// CurveEvalX finds the y value that occurs at the given x
// value, assuming that the curve is monotonic in x.
//
// If the y value cannot be found, NaN is returned.
func CurveEvalX(c Curve, x float64) float64 {
	t := CurveInverseX(c, x)
	if math.IsNaN(t) {
		return t
	}
	return c.Eval(t).Y
}

// CurveInverseX gets the t value between 0 and 1 where
// the x value is equal to some x, assuming the curve is
// monotonic in x.
//
// If the t cannot be found, NaN is returned.
func CurveInverseX(c Curve, x float64) float64 {
	return bisectionSearch(x, func(t float64) float64 {
		return c.Eval(t).X
	})
}

// CurveTranspose generates a Curve where x and y are
// swapped from the original c.
func CurveTranspose(c Curve) Curve {
	return &transposedCurve{C: c}
}

// CurveMesh creates a mesh with n evenly-spaced segments
// along the curve.
func CurveMesh(c Curve, n int) *Mesh {
	m := NewMesh()
	c1 := c.Eval(0.0)
	for i := 0; i < n; i++ {
		t2 := float64(i+1) / float64(n)
		c2 := c.Eval(t2)
		m.Add(&Segment{c1, c2})
		c1 = c2
	}
	return m
}

// A JoinedCurve combines Curves into a single curve.
// Each curve should end where the next curve begins.
type JoinedCurve []Curve

// SmoothBezier creates a joined cubic bezier curve where
// control points are reflected around end-points.
// The first four points define the first bezier curve.
// After that, each group of two points defines a control
// point and an endpoint.
func SmoothBezier(start1, c1, c2, end1 Coord, ctrlEnds ...Coord) JoinedCurve {
	if len(ctrlEnds)%2 == 1 {
		panic("must be an even number of extra points")
	}
	res := JoinedCurve{
		BezierCurve{start1, c1, c2, end1},
	}
	lastCtrl := c2
	lastEnd := end1
	for i := 0; i < len(ctrlEnds); i += 2 {
		nextCtrl := ctrlEnds[i]
		nextEnd := ctrlEnds[i+1]
		res = append(res, BezierCurve{
			lastEnd,
			lastEnd.Add(lastEnd.Sub(lastCtrl)),
			nextCtrl,
			nextEnd,
		})
		lastCtrl, lastEnd = nextCtrl, nextEnd
	}
	return res
}

// Eval evaluates the joint curve.
//
// Each sub-curve consumes an equal fraction of t.
// For t outside of [0, 1], the first or last curve is
// used.
func (j JoinedCurve) Eval(t float64) Coord {
	if t == 1 {
		return j[len(j)-1].Eval(1)
	}
	curveIdx := int(t * float64(len(j)))
	if curveIdx == len(j) {
		curveIdx--
	} else if curveIdx < 0 {
		curveIdx = 0
	}
	subT := t*float64(len(j)) - float64(curveIdx)
	return j[curveIdx].Eval(subT)
}

// A JoinedArcLenCurve combines Curves into a single curve.
// Each curve should end where the next curve begins.
//
// Unlike JoinedCurve, this does not evenly divide the range of
// t between sub-curves, but rather allocates a sub-range of t
// proportional to the arc length of each curve.
type JoinedArcLenCurve struct {
	cumuLengths []float64
	totalLength float64
	curves      []ArcLenCurve
}

// NewJoinedArcLenCurve joins the curves.
func NewJoinedArcLenCurve[T ArcLenCurve](curves []T) *JoinedArcLenCurve {
	lengths := make([]float64, len(curves))
	cs := make([]ArcLenCurve, len(curves))
	total := 0.0
	for i, c := range curves {
		total += c.ArcLen()
		lengths[i] = total
		cs[i] = c
	}
	return &JoinedArcLenCurve{
		cumuLengths: lengths,
		totalLength: total,
		curves:      cs,
	}
}

// Eval evaluates the joint curve.
func (e *JoinedArcLenCurve) Eval(t float64) Coord {
	if t == 1 {
		// Make sure to return the endpoint exactly without rounding issues,
		// to ensure meshes created from the curve are manifold.
		return e.curves[len(e.curves)-1].Eval(1)
	} else if t == 0 {
		return e.curves[0].Eval(0)
	}
	lenValue := t * e.totalLength
	found := sort.SearchFloat64s(e.cumuLengths, lenValue)
	if found == len(e.curves) {
		found = len(e.curves) - 1
	}
	endLen := e.cumuLengths[found]
	startLen := 0.0
	if found > 0 {
		startLen = e.cumuLengths[found-1]
	}
	curveLen := endLen - startLen
	frac := math.Max(0, math.Min(1, (lenValue-startLen)/curveLen))
	return e.curves[found].Eval(frac)
}

// ArcLen returns the total arc length of all subcurves.
func (e *JoinedArcLenCurve) ArcLen() float64 {
	return e.totalLength
}

// BezierCurve implements an arbitrarily high-dimensional
// Bezier curve.
type BezierCurve []Coord

// Eval evaluates the curve at time t, where 0 <= t <= 1.
func (b BezierCurve) Eval(t float64) Coord {
	if len(b) < 2 {
		panic("need at least two points")
	} else if len(b) == 2 {
		return b[0].Scale(1 - t).Add(b[1].Scale(t))
	} else if len(b) == 3 {
		t2 := t * t
		invT := 1 - t
		invT2 := invT * invT
		return b[0].Scale(invT2).Add(b[1].Scale(2 * invT * t)).Add(b[2].Scale(t2))
	} else if len(b) == 4 {
		t2 := t * t
		t3 := t2 * t
		invT := 1 - t
		invT2 := invT * invT
		invT3 := invT2 * invT
		res := b[0].Scale(invT3)
		res = res.Add(b[1].Scale(3 * invT2 * t))
		res = res.Add(b[2].Scale(3 * invT * t2))
		res = res.Add(b[3].Scale(t3))
		return res
	} else if len(b)-2 < len(binomialCoeffs) {
		sum, _ := recursiveBezierFast(b, 0, t, 1)
		return sum
	}
	return b[:len(b)-1].Eval(t).Scale(1 - t).Add(b[1:].Eval(t).Scale(t))
}

// recursiveBezierFast evaluates a bezier curve without
// any explicit allocations in time linear with the size
// of the curve.
//
// Hack to use the stack to store invTProd in the opposite
// order as tProd.
func recursiveBezierFast(b BezierCurve, i int, t, tProd float64) (sum Coord, invTProd float64) {
	if i == len(b) {
		return Coord{}, 1
	}
	sum, invTProd = recursiveBezierFast(b, i+1, t, tProd*t)
	sum = sum.Add(b[i].Scale(binomialCoeffs[len(b)-2][i] * invTProd * tProd))
	invTProd *= (1 - t)
	return
}

// EvalX finds the y value that occurs at the given x
// value, assuming that the curve is monotonic in x.
//
// If the y value cannot be found, NaN is returned.
func (b BezierCurve) EvalX(x float64) float64 {
	return CurveEvalX(b, x)
}

// CachedEvalX returns a function like EvalX that is
// cached between calls in a concurrency-safe manner.
func (b BezierCurve) CachedEvalX(x float64) func(float64) float64 {
	return CacheScalarFunc(b.EvalX)
}

// InverseX gets the t value between 0 and 1 where the x
// value is equal to some x, assuming the curve is
// monotonic in x.
//
// If the t cannot be found, NaN is returned.
func (b BezierCurve) InverseX(x float64) float64 {
	return CurveInverseX(b, x)
}

// Transpose generates a BezierCurve where x and y are
// swapped.
func (b BezierCurve) Transpose() BezierCurve {
	var res BezierCurve
	for _, c := range b {
		res = append(res, Coord{X: c.Y, Y: c.X})
	}
	return res
}

// Split creates two Bezier curves from b, where the first
// curve represents b in the range [0, t] and the second
// in the range [t, 1].
func (b BezierCurve) Split(t float64) (BezierCurve, BezierCurve) {
	c1 := make(BezierCurve, len(b))
	c2 := make(BezierCurve, len(b))

	for axis := 0; axis < 2; axis++ {
		// https://en.wikipedia.org/wiki/De_Casteljau%27s_algorithm
		n := len(b) - 1
		firstRow := make([]float64, n+1)
		for i, c := range b {
			firstRow[i] = c.Array()[axis]
		}
		betas := [][]float64{firstRow}
		for j := 1; j <= n; j++ {
			prev := betas[j-1]
			row := make([]float64, n-j+1)
			for i := range row {
				row[i] = prev[i]*(1-t) + prev[i+1]*t
			}
			betas = append(betas, row)
		}
		for i, row := range betas {
			arr := c1[i].Array()
			arr[axis] = row[0]
			c1[i] = NewCoordArray(arr)
			arr = c2[i].Array()
			arr[axis] = betas[n-i][i]
			c2[i] = NewCoordArray(arr)
		}
	}

	return c1, c2
}

// Polynomials converts the X and Y coordinates of the
// curve into polynomials of t.
func (b BezierCurve) Polynomials() [2]numerical.Polynomial {
	if len(b) == 0 {
		return [2]numerical.Polynomial{nil, nil}
	} else if len(b) == 1 {
		return [2]numerical.Polynomial{{b[0].X}, {b[0].Y}}
	}
	p1 := b[:len(b)-1].Polynomials()
	p2 := b[1:].Polynomials()

	// Polynomials representing (1-t) and t
	t1 := numerical.Polynomial{1, -1}
	t2 := numerical.Polynomial{0, 1}

	return [2]numerical.Polynomial{
		p1[0].Mul(t1).Add(p2[0].Mul(t2)),
		p1[1].Mul(t1).Add(p2[1].Mul(t2)),
	}
}

// ArcLen computes the arc length with a default tolerance.
// For more control, see Length().
func (b BezierCurve) ArcLen() float64 {
	return b.Length(1e-8, 0)
}

// Length approximates the arc length of the curve within
// the given margin of error.
//
// If maxSplits is specified, it determines the maximum
// number of sub-divisions to perform. Otherwise,
// DefaultBezierMaxSplits is used.
func (b BezierCurve) Length(tol float64, maxSplits int) float64 {
	if maxSplits == 0 {
		maxSplits = DefaultBezierMaxSplits
	}
	if len(b) == 4 {
		return b.cubicLength(tol, maxSplits)
	} else {
		return b.length(tol, maxSplits)
	}
}

func (b BezierCurve) length(tol float64, maxSplits int) float64 {
	lowerBound := b[0].Dist(b[len(b)-1])
	upperBound := 0.0
	for i, c := range b[1:] {
		upperBound += c.Dist(b[i])
	}
	// Simplest version of adaptive subdivision.
	// See "Adaptive subdivision and the length and energy of Bézier curves"
	// (https://www.sciencedirect.com/science/article/pii/0925772195000542).
	if maxSplits == 0 || upperBound-lowerBound < tol {
		n := len(b) - 1
		return (2*lowerBound + float64(n-1)*upperBound) / float64(n+1)
	}
	b1, b2 := b.Split(0.5)
	return b1.length(tol/2, maxSplits-1) + b2.length(tol/2, maxSplits-1)
}

func (b BezierCurve) cubicLength(tol float64, maxSplits int) float64 {
	// Algorithm ported from:
	// https://github.com/linebender/kurbo/blob/347dbc7cdfb864ee3b5a6832b8748daf98722ade/src/cubicbez.rs#L213

	if b[0].SquaredDist(b[1])+b[2].SquaredDist(b[3]) <= 0.5*tol*tol {
		// Nearly degenerate Bezier curve.
		return b[0].Dist(b[3])
	}

	ps := b.Polynomials()

	cubicErrnorm := func() float64 {
		d1 := ps[0].Derivative().Derivative()
		d2 := ps[1].Derivative().Derivative()
		d := BezierCurve{XY(d1.Eval(0), d2.Eval(0)), XY(d1.Eval(1), d2.Eval(1))}
		dd := d[1].Sub(d[0])
		return d[0].Dot(d[0]) + d[0].Dot(dd) + (1.0/3.0)*dd.Dot(dd)
	}
	errEstimate := func() float64 {
		lc := b[0].Dist(b[len(b)-1])
		lp := 0.0
		for i, c := range b[1:] {
			lp += c.Dist(b[i])
		}
		return 2.56e-8 * math.Pow(cubicErrnorm()/(lc*lc), 8) * lp
	}

	if maxSplits == 0 || errEstimate() < tol {
		// 9th-order Gauss-Legendre quadrature.
		deriv1 := ps[0].Derivative()
		deriv2 := ps[1].Derivative()
		normCurve := func(x float64) float64 {
			return XY(deriv1.Eval(x), deriv2.Eval(x)).Norm()
		}
		// Gauss-Legendre coefficients.
		weightsAndXs := [9][2]float64{
			{0.3302393550012598, 0.0000000000000000},
			{0.1806481606948574, -0.8360311073266358},
			{0.1806481606948574, 0.8360311073266358},
			{0.0812743883615744, -0.9681602395076261},
			{0.0812743883615744, 0.9681602395076261},
			{0.3123470770400029, -0.3242534234038089},
			{0.3123470770400029, 0.3242534234038089},
			{0.2606106964029354, -0.6133714327005904},
			{0.2606106964029354, 0.6133714327005904},
		}
		sum := 0.0
		for i := range weightsAndXs {
			wi, xi := weightsAndXs[i][0], weightsAndXs[i][1]
			sum += wi * normCurve(0.5*(xi+1))
		}
		return sum / 2
	}

	b1, b2 := b.Split(0.5)
	return b1.cubicLength(tol/2, maxSplits-1) + b2.cubicLength(tol/2, maxSplits-1)
}

// An ArcCurve implements an elliptical arc in the style of SVG.
type ArcCurve struct {
	radii    Coord
	start    Coord
	end      Coord
	rotation float64
	largeArc bool
	sweep    bool

	cosRotation, sinRotation float64
	center                   Coord
	startTheta               float64
	endTheta                 float64
}

// NewArcCurve creates a curve and precomputes its geometry.
func NewArcCurve(radii, start, end Coord, rotation float64, largeArc, sweep bool) *ArcCurve {
	result := &ArcCurve{
		radii:    radii.Abs(),
		start:    start,
		end:      end,
		rotation: rotation,
		largeArc: largeArc,
		sweep:    sweep,

		cosRotation: math.Cos(rotation),
		sinRotation: math.Sin(rotation),
	}
	result.center, result.startTheta, result.endTheta = result.calculate()
	return result
}

// det computes a 2D determinant based on two rows.
func det(a, b Coord) float64 { return a.X*b.Y - a.Y*b.X }

// angleBetween returns a signed angle from u to v in (-pi, pi]
func angleBetween(u, v Coord) float64 {
	return math.Atan2(det(u, v), u.Dot(v))
}

func (a *ArcCurve) calculate() (center Coord, startTheta, endTheta float64) {
	// Implements SVG arc conversion (endpoint -> center parameterization).
	// Rotation is assumed to be in radians.
	x1, y1 := a.start.X, a.start.Y
	x2, y2 := a.end.X, a.end.Y

	rx := a.radii.X
	ry := a.radii.Y

	// Degenerate: treat as point/line; Eval() will handle.
	if (x1 == x2 && y1 == y2) || rx == 0 || ry == 0 {
		return XY((x1+x2)*0.5, (y1+y2)*0.5), 0, 0
	}

	cphi := a.cosRotation
	sphi := a.sinRotation

	// Step 1: compute (x1', y1')
	dx := (x1 - x2) * 0.5
	dy := (y1 - y2) * 0.5
	x1p := cphi*dx + sphi*dy
	y1p := -sphi*dx + cphi*dy

	// Step 2: correct radii if too small
	lam := x1p*x1p/(rx*rx) + y1p*y1p/(ry*ry)
	if lam > 1 {
		s := math.Sqrt(lam)
		rx *= s
		ry *= s
	}

	// Step 3: compute center in the prime coordinates (cx', cy')
	// sign = +1 or -1 based on flags
	sign := 1.0
	if a.largeArc == a.sweep {
		sign = -1.0
	}

	rx2 := rx * rx
	ry2 := ry * ry
	x1p2 := x1p * x1p
	y1p2 := y1p * y1p

	num := rx2*ry2 - rx2*y1p2 - ry2*x1p2
	den := rx2*y1p2 + ry2*x1p2

	// Numeric guard: if slightly negative due to floating error, clamp to 0.
	if den == 0 {
		// Shouldn’t happen unless start/end are pathological; fall back.
		return XY((x1+x2)*0.5, (y1+y2)*0.5), 0, 0
	}
	factor := num / den
	if factor < 0 {
		factor = 0
	}
	coef := sign * math.Sqrt(factor)

	cxp := coef * (rx * y1p / ry)
	cyp := coef * (-ry * x1p / rx)

	// Step 4: transform center back to original coordinates
	mid := XY((x1+x2)*0.5, (y1+y2)*0.5)
	center = XY(
		cphi*cxp-sphi*cyp+mid.X,
		sphi*cxp+cphi*cyp+mid.Y,
	)

	// Step 5: compute angles
	// v1 = ((x1' - cx')/rx, (y1' - cy')/ry)
	// v2 = ((x2' - cx')/rx, (y2' - cy')/ry) but x2'=-x1', y2'=-y1'
	v1 := XY((x1p-cxp)/rx, (y1p-cyp)/ry)
	v2 := XY((-x1p-cxp)/rx, (-y1p-cyp)/ry)

	startTheta = angleBetween(Coord{1, 0}, v1)
	delta := angleBetween(v1, v2)

	// Step 6: adjust delta based on sweep flag to get the correct arc
	if !a.sweep && delta > 0 {
		delta -= 2 * math.Pi
	} else if a.sweep && delta < 0 {
		delta += 2 * math.Pi
	}

	endTheta = startTheta + delta
	return center, startTheta, endTheta
}

func (a *ArcCurve) Eval(t float64) Coord {
	// t in [0,1], evaluate from start (t=0) to end (t=1)
	if t <= 0 {
		return a.start
	}
	if t >= 1 {
		return a.end
	}

	rx := a.radii.X
	ry := a.radii.Y

	// Degenerate behaviors (SVG-like):
	// - If radii are zero, it's a straight line.
	// - If start==end, it's a single point.
	if rx == 0 || ry == 0 || (a.start.X == a.end.X && a.start.Y == a.end.Y) {
		return a.start.Add(a.end.Sub(a.start).Scale(t))
	}

	center, th0, th1 := a.calculate()
	dth := th1 - th0
	theta := th0 + dth*t

	cphi := a.cosRotation
	sphi := a.sinRotation

	// SVG center parameterization:
	// x = cx + cosφ*rx*cosθ - sinφ*ry*sinθ
	// y = cy + sinφ*rx*cosθ + cosφ*ry*sinθ
	ct := math.Cos(theta)
	st := math.Sin(theta)

	return Coord{
		X: center.X + cphi*rx*ct - sphi*ry*st,
		Y: center.Y + sphi*rx*ct + cphi*ry*st,
	}
}

// ArcLen approximates the length of the arc, returning an
// exact value when the radii match.
func (a *ArcCurve) ArcLen() float64 {
	rx := a.radii.X
	ry := a.radii.Y

	// Degenerate cases match Eval()
	if rx == 0 || ry == 0 || (a.start.X == a.end.X && a.start.Y == a.end.Y) {
		return a.end.Sub(a.start).Norm()
	}

	th0 := a.startTheta
	th1 := a.endTheta
	dth := th1 - th0

	if dth == 0 {
		return 0
	}

	// Integrand for arc length
	f := func(theta float64) float64 {
		s := math.Sin(theta)
		c := math.Cos(theta)
		return math.Sqrt(rx*rx*s*s + ry*ry*c*c)
	}

	// 5-point Gauss-Legendre quadrature on [-1,1]
	nodes := []float64{
		0.0,
		-0.5384693101056831,
		0.5384693101056831,
		-0.9061798459386640,
		0.9061798459386640,
	}

	weights := []float64{
		0.5688888888888889,
		0.4786286704993665,
		0.4786286704993665,
		0.2369268850561891,
		0.2369268850561891,
	}

	// Map from [-1,1] to [th0, th1]
	mid := 0.5 * (th0 + th1)
	half := 0.5 * (th1 - th0)

	sum := 0.0
	for i := range nodes {
		theta := mid + half*nodes[i]
		sum += weights[i] * f(theta)
	}

	return half * sum
}

// FuncCurve is a Curve defined as a single function.
type FuncCurve func(t float64) Coord

// Eval returns the result of f(t).
func (f FuncCurve) Eval(t float64) Coord {
	return f(t)
}

// A SegmentCurve is a curve comprised of connected linear
// segments.
type SegmentCurve struct {
	segments    []*Segment
	lengths     []float64
	totalLength float64
}

// NewSegmentCurve creates a SegmentCurve from a sequence
// of consecutive segments along the curve.
func NewSegmentCurve(segments []*Segment) *SegmentCurve {
	var lengths []float64
	var totalLength float64
	for _, x := range segments {
		lengths = append(lengths, totalLength)
		totalLength += x.Length()
	}
	return &SegmentCurve{
		segments:    segments,
		lengths:     lengths,
		totalLength: totalLength,
	}
}

// NewSegmentCurveMesh turns a mesh, which must contain a single
// unclosed continuous path, into a curve.
//
// The mesh should have exactly two vertices which are
// connected to one segment, and the rest should connect to
// two segments.
// The second vertex of each segment should be connected to
// the first vertex of the next segment, except for the
// final segment in the path.
func NewSegmentCurveMesh(m *Mesh) *SegmentCurve {
	var firstSeg *Segment
	m.Iterate(func(s *Segment) {
		if count := len(m.Find(s[0])); count == 1 {
			if firstSeg != nil {
				panic("multiple first-vertices found in mesh")
			}
			firstSeg = s
		} else if count != 2 {
			panic("mesh has loops")
		}
	})

	var segments []*Segment
	removed := m.Copy()
	removed.Remove(firstSeg)
	segments = append(segments, firstSeg)
	nextPoint := firstSeg[1]
	for removed.NumSegments() > 0 {
		segs := removed.Find(nextPoint)
		if len(segs) != 1 {
			panic("mesh is not a valid single closed path")
		}
		segments = append(segments, segs[0])
		nextPoint = segs[0][1]
		removed.Remove(segs[0])
	}
	return NewSegmentCurve(segments)
}

// Eval returns the piecewise-linear interpolated point.
func (s *SegmentCurve) Eval(t float64) Coord {
	l := t * s.totalLength
	idx := sort.SearchFloat64s(s.lengths, l)
	if idx == len(s.segments) {
		idx -= 1
	}
	seg := s.segments[idx]
	offset := l - s.lengths[idx]
	offsetFrac := offset / seg.Length()
	return seg[0].Add(seg[1].Sub(seg[0]).Scale(offsetFrac))
}

// ArcLen returns the total segment length.
func (s *SegmentCurve) ArcLen() float64 {
	return s.totalLength
}

// CacheScalarFunc creates a scalar function that is
// equivalent to a deterministic function f, but caches
// results across calls in a concurrency-safe manner.
func CacheScalarFunc(f func(float64) float64) func(float64) float64 {
	cache := sync.Map{}
	return func(x float64) float64 {
		value, ok := cache.Load(x)
		if ok {
			return value.(float64)
		} else {
			y := f(x)
			cache.Store(x, y)
			return y
		}
	}
}

func bisectionSearch(x float64, f func(float64) float64) float64 {
	lowT := 0.0
	highT := 1.0
	x0 := f(0)
	x1 := f(1)
	if x0 == x {
		return 0
	} else if x1 == x {
		return 1
	}
	eval0 := x0 <= x
	eval1 := x1 <= x
	if eval0 == eval1 {
		return math.NaN()
	} else if eval1 {
		highT, lowT = lowT, highT
	}

	for i := 0; i < 63; i++ {
		t := (lowT + highT) / 2
		if f(t) <= x {
			lowT = t
		} else {
			highT = t
		}
	}

	return (lowT + highT) / 2
}

type transposedCurve struct {
	C Curve
}

func (t *transposedCurve) Eval(tVal float64) Coord {
	c := t.C.Eval(tVal)
	return XY(c.Y, c.X)
}
