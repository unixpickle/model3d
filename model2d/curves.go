package model2d

import (
	"math"
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
	curveIdx := int(t * float64(len(j)))
	if curveIdx == len(j) {
		curveIdx--
	} else if curveIdx < 0 {
		curveIdx = 0
	}
	subT := t*float64(len(j)) - float64(curveIdx)
	return j[curveIdx].Eval(subT)
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

// Length approximates the arclength of the curve within
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
