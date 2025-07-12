package model2d

import (
	"math"
	"testing"
)

func TestBezierFitterSingle(t *testing.T) {
	target, samples, fitter := testBezierFitProblem()
	fit := fitter.FitCubic(samples, nil)
	for i, x := range target {
		a := fit[i]
		if x.Dist(a) > 2e-4 {
			t.Errorf("control point %d should be %v but got %v", i, x, a)
		}
	}
}

func TestBezierFitterNaNCase(t *testing.T) {
	// Fitting these points historically gave a NaN because
	// the gradient was zero.
	points := []Coord{
		XY(9.5720846748406, 18.68020595991276),
		XY(9.7874167060906, 18.68167080366276),
		XY(9.5720846748406, 18.92141689741276),
		XY(9.5696432685906, 18.68167080366276),
	}
	fitter := &BezierFitter{
		NumIters:     200,
		Delta:        1e-7,
		LineStep:     1.1,
		PerimPenalty: 1e-4,
		Momentum:     0.5,
		NoFirstGuess: true,
	}
	fit := fitter.FitCubic(points, nil)
	for _, c := range fit {
		if math.IsNaN(c.Sum()) {
			t.Fatalf("got NaNs in fit: %v", fit)
		}
	}
}

func TestBezierFitterFirstGuess(t *testing.T) {
	target, samples, fitter := testBezierFitProblem()
	fit := fitter.FirstGuess(samples)
	for i, x := range target[1:3] {
		a := fit[i+1]
		if x.Dot(a)/math.Sqrt(x.Dot(x)*a.Dot(a)) < 0.95 {
			t.Errorf("control point %d should be %v but got %v", i, x, a)
		}
	}
}

func TestBezierFitterSingleConstrained(t *testing.T) {
	target, samples, fitter := testBezierFitProblem()
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
	_, samples, fitter := testBezierFitProblem()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fitter.FitCubic(samples, nil)
	}
}

func BenchmarkBezierFitterSingleConstrained(b *testing.B) {
	target, samples, fitter := testBezierFitProblem()
	t1 := target[1].Sub(target[0])
	t2 := target[3].Sub(target[2])
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fitter.FitCubicConstrained(samples, &t1, &t2, nil)
	}
}

func BenchmarkBezierFitterFirstGuess(b *testing.B) {
	_, samples, fitter := testBezierFitProblem()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fitter.FirstGuess(samples)
	}
}

func testBezierFitProblem() (BezierCurve, []Coord, *BezierFitter) {
	target := BezierCurve{XY(0, 0), XY(1, -1), XY(2, 1.5), XY(3, 0)}
	samples := make([]Coord, 50)
	for i := range samples {
		samples[i] = target.Eval(float64(i) / float64(len(samples)-1))
	}
	fitter := &BezierFitter{NumIters: 100, Momentum: 0.5}
	return target, samples, fitter
}
