package model2d

import (
	"math"
)

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

// Evaluate a bezier curve without any explicit allocations.
// in time linear with the size of the curve.
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
	t := b.InverseX(x)
	if math.IsNaN(t) {
		return t
	}
	return b.Eval(t).Y
}

// InverseX gets the t value between 0 and 1 where the x
// value is equal to some x, assuming the curve is
// monotonic in x.
//
// If the t cannot be found, NaN is returned.
func (b BezierCurve) InverseX(x float64) float64 {
	lowT := 0.0
	highT := 1.0
	x0 := b.Eval(0).X
	x1 := b.Eval(1).X
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
		if b.Eval(t).X <= x {
			lowT = t
		} else {
			highT = t
		}
	}

	return (lowT + highT) / 2
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
