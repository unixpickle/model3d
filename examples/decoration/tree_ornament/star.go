package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
)

const (
	StarThickness   = 1.0
	StarPointRadius = 2.0
	StarRingRadius  = 0.2

	StarHolderRadius    = 0.4
	StarHolderLength    = 2.0
	StarHolderThickness = 0.05
	StarHolderOffset    = 0.4

	StarNumPoints = 6
)

func CreateStarSolid() model3d.Solid {
	baseMesh := CreateStarMesh()
	return model3d.JoinedSolid{
		model3d.NewColliderSolid(model3d.MeshToCollider(baseMesh)),
		CreateHolder(model3d.X(StarPointRadius - StarHolderOffset)),
		&model3d.Torus{
			Axis:        model3d.Z(1),
			Center:      model3d.X(-StarPointRadius),
			InnerRadius: 0.05,
			OuterRadius: StarRingRadius,
		},
	}
}

func CreateStarMesh() *model3d.Mesh {
	midPoint := model3d.Z(StarThickness / 2)

	mesh := model3d.NewMesh()
	for i := 0; i < StarNumPoints*2; i += 2 {
		theta0 := float64(i-1) / float64(StarNumPoints*2) * math.Pi * 2
		theta1 := float64(i) / float64(StarNumPoints*2) * math.Pi * 2
		theta2 := float64(i+1) / float64(StarNumPoints*2) * math.Pi * 2

		p1 := model3d.XY(math.Cos(theta0), math.Sin(theta0))
		p2 := model3d.XY(math.Cos(theta1), math.Sin(theta1)).Scale(StarPointRadius)
		p3 := model3d.XY(math.Cos(theta2), math.Sin(theta2))

		mesh.Add(&model3d.Triangle{p2, p1, midPoint})
		mesh.Add(&model3d.Triangle{p2, p3, midPoint})
	}
	mesh.AddMesh(mesh.MapCoords(model3d.XYZ(1, 1, -1).Mul))

	// We created the mesh in a lazy way, so we must
	// fix holes and normals.
	mesh = mesh.Repair(1e-5)
	mesh, _ = mesh.RepairNormals(1e-5)
	return mesh
}

func CreateHolder(tip model3d.Coord3D) model3d.Solid {
	conePoint := func(t, theta float64) model3d.Coord3D {
		r := StarHolderRadius * math.Sqrt(t)
		x := t*StarHolderLength + tip.X
		return model3d.XYZ(x, math.Cos(theta)*r, math.Sin(theta)*r)
	}

	surfaceMesh := model3d.NewMesh()
	dTheta := math.Pi * 2 / 100.0
	dT := 1.0 / 100.0
	for t := 0.0; t < 1.0; t += dT {
		for theta := 0.0; theta < math.Pi*2; theta += dTheta {
			p1 := conePoint(t, theta)
			p2 := conePoint(t, theta+dTheta)
			p3 := conePoint(t+dT, theta+dTheta)
			p4 := conePoint(t+dT, theta)
			surfaceMesh.AddQuad(p1, p2, p3, p4)
		}
	}

	return model3d.NewColliderSolidHollow(model3d.MeshToCollider(surfaceMesh), StarHolderThickness)
}
