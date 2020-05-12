package main

import (
	"log"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

type ControlPoints struct {
	LeftArm  model3d.Coord3D
	RightArm model3d.Coord3D
	Head     model3d.Coord3D
	Foot     model3d.Coord3D
}

func main() {
	mesh := CreateMesh()
	points := FindControlPoints(mesh)

	SaveRendering("deform_0.png", mesh)

	a := model3d.NewARAP(mesh)

	log.Println("Creating first deformation...")
	deformed := a.Deform(model3d.ARAPConstraints{
		points.Foot:     points.Foot,
		points.LeftArm:  points.LeftArm.Add(model3d.Coord3D{Z: 0.2, X: 0.05}),
		points.RightArm: points.RightArm.Add(model3d.Coord3D{Z: 0.2, X: -0.05}),
	})
	SaveRendering("deform_1.png", deformed)

	log.Println("Creating second deformation...")
	deformed = a.Deform(model3d.ARAPConstraints{
		points.Foot:     points.Foot,
		points.RightArm: points.RightArm.Add(model3d.Coord3D{Z: -0.3, X: 0.1}),
	})
	SaveRendering("deform_2.png", deformed)
}

func SaveRendering(name string, mesh *model3d.Mesh) {
	render3d.SaveRendering(name, mesh, model3d.Coord3D{Z: 0.5, Y: -2}, 300, 300, nil)
}

func CreateMesh() *model3d.Mesh {
	solid := model3d.JoinedSolid{
		// Body.
		&RoundedTube{
			P2:     model3d.Coord3D{Z: 1.0},
			Radius: 0.1,
		},
		// Arms.
		&RoundedTube{
			P1:     model3d.Coord3D{Z: 0.6, X: -0.4},
			P2:     model3d.Coord3D{Z: 0.6, X: 0.4},
			Radius: 0.08,
		},
	}
	return model3d.MarchingCubesSearch(solid, 0.01, 8).FlipDelaunay()
}

func FindControlPoints(m *model3d.Mesh) *ControlPoints {
	min, max := m.Min(), m.Max()
	p := &ControlPoints{}
	for _, v := range m.VertexSlice() {
		if v.X == min.X {
			p.LeftArm = v
		} else if v.X == max.X {
			p.RightArm = v
		} else if v.Z == min.Z {
			p.Foot = v
		} else if v.Z == max.Z {
			p.Head = v
		}
	}
	return p
}

type RoundedTube struct {
	P1     model3d.Coord3D
	P2     model3d.Coord3D
	Radius float64
}

func (r *RoundedTube) Min() model3d.Coord3D {
	return r.boundingCylinder().Min()
}

func (r *RoundedTube) Max() model3d.Coord3D {
	return r.boundingCylinder().Max()
}

func (r *RoundedTube) Contains(c model3d.Coord3D) bool {
	seg := model3d.NewSegment(r.P1, r.P2)
	return seg.Dist(c) < r.Radius
}

func (r *RoundedTube) boundingCylinder() *model3d.Cylinder {
	axis := r.P2.Sub(r.P1).Normalize().Scale(r.Radius)
	return &model3d.Cylinder{
		P1:     r.P1.Sub(axis),
		P2:     r.P2.Add(axis),
		Radius: r.Radius,
	}
}
