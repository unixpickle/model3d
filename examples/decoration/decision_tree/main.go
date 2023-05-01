package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	model := TreeSolid(
		TreeSolid(
			TreeSolid(
				LeafSolid(),
				LeafSolid(),
			),
			TreeSolid(
				LeafSolid(),
				LeafSolid(),
			),
		),
		TreeSolid(
			TreeSolid(
				LeafSolid(),
				LeafSolid(),
			),
			TreeSolid(
				LeafSolid(),
				LeafSolid(),
			),
		),
	)
	log.Println("Creating mesh...")
	mesh := model3d.DualContour(model, 0.02, true, false)
	mesh = mesh.EliminateCoplanar(1e-5)
	log.Printf("Needs repair: %v; singular vertices: %d",
		mesh.NeedsRepair(), len(mesh.SingularVertices()))
	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
	log.Println("Saving...")
	mesh.SaveGroupedSTL("decision_tree.stl")
}

func TreeSolid(left, right model3d.Solid) model3d.Solid {
	leftMin, leftMax := left.Min(), left.Max()
	rightMin, rightMax := right.Min(), right.Max()
	maxWidth := math.Max(leftMax.X-leftMin.X, rightMax.X-rightMin.X) / 2
	span := maxWidth + 0.1
	return model3d.JoinedSolid{
		&model3d.Sphere{
			Center: model3d.Z(span * 1.5),
			Radius: 0.3,
		},
		&model3d.Cylinder{
			P1:     model3d.Z(span * 1.5),
			P2:     model3d.X(-span),
			Radius: 0.15,
		},
		&model3d.Cylinder{
			P1:     model3d.Z(span * 1.5),
			P2:     model3d.X(span),
			Radius: 0.15,
		},
		model3d.TranslateSolid(left, model3d.XZ(-span-(leftMin.X+leftMax.X)/2, -leftMax.Z+0.13)),
		model3d.TranslateSolid(right, model3d.XZ(span-(rightMin.X+rightMax.X)/2, -rightMax.Z+0.13)),
	}
}

func LeafSolid() model3d.Solid {
	return model3d.NewRect(model3d.XYZ(-0.3, -0.4, -0.2), model3d.XYZ(0.3, 0.4, 0.2))
}
