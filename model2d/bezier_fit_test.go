package model2d

import (
	"math"
	"testing"
)

func TestBezierFitterSingle(t *testing.T) {
	target := BezierCurve{XY(0, 0), XY(1, -1), XY(2, 1.5), XY(3, 0)}
	samples := make([]Coord, 50)
	for i := range samples {
		samples[i] = target.Eval(float64(i) / float64(len(samples)-1))
	}
	fitter := &BezierFitter{NumIters: 100, Momentum: 0.5}
	fit := fitter.FitCubic(samples, nil)
	for i, x := range target {
		a := fit[i]
		if x.Dist(a) > 1e-4 {
			t.Errorf("control point %d should be %v but got %v", i, x, a)
		}
	}
}

func TestBezierFitterFirstGuess(t *testing.T) {
	target := BezierCurve{XY(0, 0), XY(1, -1), XY(2, 1.5), XY(3, 0)}
	samples := make([]Coord, 50)
	for i := range samples {
		samples[i] = target.Eval(float64(i) / float64(len(samples)-1))
	}
	fitter := &BezierFitter{NumIters: 100, Momentum: 0.5}
	fit := fitter.FirstGuess(samples)
	for i, x := range target[1:3] {
		a := fit[i+1]
		if x.Dot(a)/math.Sqrt(x.Dot(x)*a.Dot(a)) < 0.95 {
			t.Errorf("control point %d should be %v but got %v", i, x, a)
		}
	}
}

func TestBezierFitterSingleConstrained(t *testing.T) {
	target := BezierCurve{XY(0, 0), XY(1, -1), XY(2, 1.5), XY(3, 0)}
	samples := make([]Coord, 50)
	for i := range samples {
		samples[i] = target.Eval(float64(i) / float64(len(samples)-1))
	}
	fitter := &BezierFitter{NumIters: 100, Momentum: 0.5}
	t1 := target[1].Sub(target[0])
	t2 := target[3].Sub(target[2])
	args := [][2]*Coord{{&t1, &t2}, {&t1, nil}, {nil, &t2}}
	for _, ts := range args {
		fit := fitter.FitCubicConstrained(samples, ts[0], ts[1], nil)
		for i, x := range target {
			a := fit[i]
			if x.Dist(a) > 1e-4 {
				t.Errorf("control point %d should be %v but got %v", i, x, a)
			}
		}
	}
}

func BenchmarkBezierFitterSingle(b *testing.B) {
	target := BezierCurve{XY(0, 0), XY(1, -1), XY(2, 1.5), XY(3, 0)}
	samples := make([]Coord, 50)
	for i := range samples {
		samples[i] = target.Eval(float64(i) / float64(len(samples)-1))
	}
	fitter := &BezierFitter{NumIters: 100, Momentum: 0.5}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fitter.FitCubic(samples, nil)
	}
}
