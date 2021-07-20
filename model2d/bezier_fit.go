package model2d

import (
	"math"

	"github.com/unixpickle/model3d/numerical"
)

const (
	DefaultBezierFitterNumIters = 10
)

// A BezierFitter fits Bezier curves to points.
type BezierFitter struct {
	// NumIters is the number of gradient steps to use to
	// fit each Bezier curve.
	// If 0, DefaultBezierFitterNumIters is used.
	NumIters int
}

// Fit finds the cubic Bezier curve of best fit for the
// points.
func (b *BezierFitter) FitCubic(points []Coord, start BezierCurve) BezierCurve {
	if len(points) == 0 {
		panic("at least one point is required")
	} else if len(points) == 1 {
		return BezierCurve{points[0], points[0], points[0], points[0]}
	} else if len(points) == 2 {
		return BezierCurve{points[0], points[0], points[1], points[1]}
	}
	if start == nil {
		start = BezierCurve{points[0], points[0], points[len(points)-1], points[len(points)-1]}
	}
	for i := 0; i < b.numIters(); i++ {
		grad := b.gradient(points, start)
		start = b.lineSearch(points, start, grad)
	}
	return start
}

// lineSearch finds the (locally optimal) best step size
// and returns the result of the step.
func (b *BezierFitter) lineSearch(points []Coord, curve, grad BezierCurve) BezierCurve {
	delta := b.delta(curve)
	maxNorm := 0.0
	for _, g := range grad {
		maxNorm = math.Max(maxNorm, g.Norm())
	}
	minStep := delta / maxNorm

	bestCurve := curve
	bestLoss := b.loss(points, curve)
	evalStep := func(s float64) float64 {
		c1 := append(BezierCurve{}, curve...)
		for i, x := range c1 {
			c1[i] = x.Add(grad[i].Scale(-s))
		}
		loss := b.loss(points, c1)
		if loss < bestLoss {
			bestLoss = loss
			bestCurve = curve
		}
		return loss
	}

	guess := minStep
	for i := 0; i < 32; i++ {
		loss := evalStep(guess)
		if loss > bestLoss || math.IsNaN(loss) || math.IsInf(loss, 0) {
			break
		}
		guess *= 2.0
	}

	return bestCurve
}

// gradient computes the gradient of the loss wrt the
// curve control points.
// The endpoint gradients are set to zero.
func (b *BezierFitter) gradient(points []Coord, curve BezierCurve) BezierCurve {
	delta := b.delta(curve)

	c1 := append(BezierCurve{}, curve...)
	grad := make(BezierCurve, len(curve))
	for i := 1; i < len(curve)-1; i++ {
		pArr := c1[i].Array()
		gradArr := [2]float64{}
		for axis := 0; axis < 2; axis++ {
			newPArr := pArr
			newPArr[axis] += delta
			c1[i] = NewCoordArray(newPArr)
			loss1 := b.loss(points, c1)
			newPArr[axis] -= 2 * delta
			c1[i] = NewCoordArray(newPArr)
			loss2 := b.loss(points, c1)
			gradArr[axis] = (loss1 - loss2) / (2 * delta)
		}
		c1[i] = curve[i]
		grad[i] = NewCoordArray(gradArr)
	}

	return grad
}

// loss computes the MSE of a Bezier fit.
func (b *BezierFitter) loss(points []Coord, curve BezierCurve) float64 {
	if len(curve) != 4 {
		panic("curve must be a cubic Bezier curve (i.e. have four points)")
	}
	axisPolynomial := func(axis int) numerical.Polynomial {
		// Expand 3(1-t)^3*w + 3(1-t)^2t*x + 3(1-t)t^2y + t^3z =>
		// -3 t^3 w + 3 t^3 x - 3 t^3 y + t^3 z + 9 t^2 w - 6 t^2 x + 3 t^2 y - 9 t w + 3 t x + 3 w
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
			3 * w, -9*w + 3*x, 9*w - 6*x + 3*y, -3*w + 3*x - 3*y + z,
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
		for _, t := range append(lossPoly.RealRoots(), 0.0, 1.0) {
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
	return max.Dist(min) * 1e-5
}

func (b *BezierFitter) numIters() int {
	if b.NumIters == 0 {
		return DefaultBezierFitterNumIters
	}
	return b.NumIters
}
