package model2d

import "testing"

func TestBezierFitterSingle(t *testing.T) {
	target := BezierCurve{XY(0, 0), XY(1, -1), XY(2, 1.5), XY(3, 0)}
	samples := make([]Coord, 50)
	for i := range samples {
		samples[i] = target.Eval(float64(i) / float64(len(samples)-1))
	}
	fitter := &BezierFitter{NumIters: 400}
	fit := fitter.FitCubic(samples, nil)
	for i, x := range target {
		a := fit[i]
		if x.Dist(a) > 1e-4 {
			t.Errorf("control point %d should be %v but got %v", i, x, a)
		}
	}
}

func BenchmarkBezierFitterSingle(b *testing.B) {
	target := BezierCurve{XY(0, 0), XY(1, -1), XY(2, 1.5), XY(3, 0)}
	samples := make([]Coord, 50)
	for i := range samples {
		samples[i] = target.Eval(float64(i) / float64(len(samples)-1))
	}
	fitter := &BezierFitter{NumIters: 400}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fitter.FitCubic(samples, nil)
	}
}
