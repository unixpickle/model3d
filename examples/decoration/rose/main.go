package main

import (
	"log"
	"math"
	"math/rand"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const NumStops = 100

func main() {
	rings := []*RingFunction{
		RandomRingFunction(math.Pi / 3),
		RandomRingFunction(math.Pi / 4),
		RandomRingFunction(math.Pi / 6),
		RandomRingFunction(math.Pi / 8),
	}
	m := model3d.NewMesh()
	for _, ring := range rings {
		ring.Add(m)
	}
	collider := model3d.MeshToCollider(m)
	solid := model3d.NewColliderSolidHollow(collider, 0.1)
	m1 := model3d.MarchingCubesSearch(solid, 0.01, 8).Blur(-1)
	m1.SaveGroupedSTL("rose.stl")

	log.Println("Generating rendering...")
	render3d.SaveRendering("rendering.png", m1, model3d.Coord3D{Y: -1, Z: 2}, 500, 500, nil)
}

// RingFunction is one conic surface originating from the
// stem of the rose.
type RingFunction struct {
	Angle        float64
	LengthCoeffs []float64
}

func RandomRingFunction(angle float64) *RingFunction {
	res := &RingFunction{
		Angle:        angle,
		LengthCoeffs: make([]float64, 10),
	}
	for i := range res.LengthCoeffs {
		res.LengthCoeffs[i] = 0.1 * rand.NormFloat64() * math.Exp(-math.Pow((float64(i)-6.5), 2)/2)
	}
	return res
}

func (r *RingFunction) Add(m *model3d.Mesh) {
	for i := 0; i < NumStops; i++ {
		r.AddSegment(m, 2*math.Pi*float64(i)/NumStops, 2*math.Pi/NumStops)
	}
}

func (r *RingFunction) AddSegment(m *model3d.Mesh, theta, deltaTheta float64) {
	length1 := r.Length(theta)
	length2 := r.Length(theta + deltaTheta)
	for i := 0; i < NumStops; i++ {
		p1 := r.Point(theta, length1*float64(i)/NumStops)
		p2 := r.Point(theta+deltaTheta, length2*float64(i)/NumStops)
		p3 := r.Point(theta+deltaTheta, length2*float64(i+1)/NumStops)
		p4 := r.Point(theta, length1*float64(i+1)/NumStops)
		m.Add(&model3d.Triangle{p1, p2, p3})
		m.Add(&model3d.Triangle{p1, p3, p4})
	}
}

func (r *RingFunction) Point(theta, distance float64) model3d.Coord3D {
	vec := model3d.Coord3D{
		X: math.Cos(theta),
		Y: math.Sin(theta),
		Z: math.Tan(r.Angle),
	}
	return vec.Scale(distance / vec.Norm())
}

func (r *RingFunction) Length(theta float64) float64 {
	length := 1.0
	for i, x := range r.LengthCoeffs {
		if i%2 == 0 {
			length += x * math.Sin(theta*float64(i/2+1))
		} else {
			length += x * math.Cos(theta*float64(i/2+1))
		}
	}
	return length
}
