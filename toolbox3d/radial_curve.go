package toolbox3d

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
)

// A CurveFunc is a parametric curve defined over [0, 1].
// It returns a radius and point for each time value.
type CurveFunc = func(t float64) (model3d.Coord3D, float64)

// RadialCurve creates a solid around a parametric curve.
//
// The radius of the solid may vary across t, and extends
// in every normal direction to the curve.
func RadialCurve(steps int, closed bool, f CurveFunc) model3d.Solid {
	var result model3d.JoinedSolid
	var p0 model3d.Coord3D
	p1, r1 := f(0)
	for i := 0; i < steps; i++ {
		t2 := float64(i+1) / float64(steps)
		p2, r2 := f(t2)
		result = append(result, conicSection(p1, p2, r1, r2))
		if i > 0 {
			result = append(result, sphericalSlice(r1, p0, p1, p2))
		}
		p0 = p1
		p1, r1 = p2, r2
	}
	if closed {
		p1, r1 := f(0)
		p0, _ := f(1 - 1/float64(steps))
		p2, _ := f(1 / float64(steps))
		result = append(result, sphericalSlice(r1, p0, p1, p2))
	}
	return result.Optimize()
}

func conicSection(p1, p2 model3d.Coord3D, r1, r2 float64) model3d.Solid {
	if r1 < 0 || r2 < 0 {
		panic("negative radius not allowed")
	}
	boundCyl := &model3d.Cylinder{
		P1:     p1,
		P2:     p2,
		Radius: math.Max(r1, r2),
	}
	v := p2.Sub(p1).Normalize()
	size := p2.Dist(p1)
	return model3d.CheckedFuncSolid(
		boundCyl.Min().Sub(model3d.XYZ(0.1, 0.1, 0.1)),
		boundCyl.Max().Add(model3d.XYZ(0.1, 0.1, 0.1)),
		func(c model3d.Coord3D) bool {
			delta := c.Sub(p1)
			dot := v.Dot(delta)
			frac := dot / size
			if frac < 0 || frac > 1 {
				return false
			}
			dist := delta.Dist(v.Scale(dot))
			r := r2*frac + r1*(1-frac)
			return dist <= r
		},
	)
}

func sphericalSlice(radius float64, p1, p2, p3 model3d.Coord3D) model3d.Solid {
	center := p2
	normal1 := p1.Sub(p2)
	normal2 := p3.Sub(p2)

	sphere := &model3d.Sphere{Center: center, Radius: radius}
	constraint := model3d.ConvexPolytope{
		&model3d.LinearConstraint{
			Normal: normal1,
			Max:    normal1.Dot(center),
		},
		&model3d.LinearConstraint{
			Normal: normal2,
			Max:    normal2.Dot(center),
		},
	}
	return model3d.CheckedFuncSolid(
		sphere.Min(),
		sphere.Max(),
		func(c model3d.Coord3D) bool {
			return sphere.Contains(c) && constraint.Contains(c)
		},
	)
}
