package main

import (
	"fmt"
	"log"
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	Thickness   = 0.15
	PartSpacing = 0.01
	EtchInset   = 0.05

	AxleRadius   = 0.15
	HolderRadius = 0.3
	OrbitRadius  = 0.6
	CircleRadius = 0.4
	EtchRadius   = 0.4
)

func main() {
	log.Println("Creating axle mesh...")
	mesh := model3d.MarchingCubesSearch(CreateAxle(), 0.003, 8)
	log.Println("Creating body mesh...")
	mesh.AddMesh(model3d.MarchingCubesSearch(CreateBody(), 0.003, 8))
	log.Println("Post-processing mesh...")
	mesh = mesh.SmoothAreas(0.05, 10)
	mesh = mesh.EliminateCoplanar(1e-8)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("spinner.stl")

	log.Println("Saving rendering of mesh...")
	render3d.SaveRendering("rendering.png", mesh, model3d.Coord3D{Y: 2, Z: 2}, 400, 400, nil)
}

func CreateAxle() model3d.Solid {
	return &model3d.SubtractedSolid{
		Positive: model3d.JoinedSolid{
			&model3d.Cylinder{
				P1:     model3d.Coord3D{Z: -(Thickness + PartSpacing)},
				P2:     model3d.Coord3D{Z: Thickness + PartSpacing},
				Radius: AxleRadius - PartSpacing,
			},
			&model3d.Cylinder{
				P1:     model3d.Coord3D{Z: -(Thickness + PartSpacing)},
				P2:     model3d.Coord3D{Z: -(Thickness + PartSpacing) * 1.5},
				Radius: HolderRadius,
			},
			&model3d.Cylinder{
				P1:     model3d.Coord3D{Z: (Thickness + PartSpacing)},
				P2:     model3d.Coord3D{Z: (Thickness + PartSpacing) * 1.5},
				Radius: HolderRadius,
			},
		},
		Negative: model3d.JoinedSolid{
			&model3d.Sphere{
				Center: model3d.Coord3D{Z: -(Thickness+PartSpacing)*1.5 - 1.0},
				Radius: 1.05,
			},
			&model3d.Sphere{
				Center: model3d.Coord3D{Z: (Thickness+PartSpacing)*1.5 + 1.0},
				Radius: 1.05,
			},
		},
	}
}

func CreateBody() model3d.Solid {
	var circles model3d.JoinedSolid
	for i, theta := range []float64{0, math.Pi * 2 / 3, math.Pi * 4 / 3} {
		etching := NewEtchedImage(fmt.Sprintf("side_%d.png", i+1), theta)
		circles = append(circles, &model3d.SubtractedSolid{
			Positive: &model3d.Cylinder{
				P1: model3d.Coord3D{
					X: etching.X,
					Y: etching.Y,
					Z: -Thickness,
				},
				P2: model3d.Coord3D{
					X: etching.X,
					Y: etching.Y,
					Z: Thickness,
				},
				Radius: CircleRadius,
			},
			Negative: etching,
		})
	}
	return &model3d.SubtractedSolid{
		// Add a middle circle to hold things together.
		Positive: append(circles, &model3d.Cylinder{
			P1:     model3d.Coord3D{Z: -Thickness},
			P2:     model3d.Z(Thickness),
			Radius: CircleRadius,
		}),
		// Hole for axle.
		Negative: &model3d.Cylinder{
			P1:     model3d.Coord3D{Z: -(Thickness + 1e-5)},
			P2:     model3d.Coord3D{Z: Thickness + 1e-5},
			Radius: AxleRadius,
		},
	}
}

type EtchedImage struct {
	Solid  model2d.Solid
	X      float64
	Y      float64
	Radius float64
}

func NewEtchedImage(path string, theta float64) *EtchedImage {
	bmp := model2d.MustReadBitmap(path, nil)
	mesh := bmp.Mesh().Smooth(10)
	collider := model2d.MeshToCollider(mesh)
	solid := model2d.NewColliderSolid(collider)
	return &EtchedImage{
		Solid:  model2d.ScaleSolid(solid, 1/float64(bmp.Width)),
		X:      OrbitRadius * math.Sin(theta),
		Y:      OrbitRadius * math.Cos(theta),
		Radius: CircleRadius,
	}
}

func (e *EtchedImage) Min() model3d.Coord3D {
	return model3d.Coord3D{X: e.X - e.Radius, Y: e.Y - e.Radius, Z: -(Thickness + 1e-5)}
}

func (e *EtchedImage) Max() model3d.Coord3D {
	return model3d.Coord3D{X: e.X + e.Radius, Y: e.Y + e.Radius, Z: Thickness + 1e-5}
}

func (e *EtchedImage) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(e, c) {
		return false
	}
	if math.Abs(c.Z) > Thickness || math.Abs(c.Z) < Thickness-EtchInset {
		return false
	}

	axis1 := model3d.Coord2D{X: e.X, Y: e.Y}
	axis1 = axis1.Scale(1 / (axis1.Norm() * e.Radius))
	axis2 := model3d.Coord2D{X: -axis1.Y, Y: axis1.X}

	if c.Z < 0 {
		// Flip image on under-side of spinner.
		axis2 = axis2.Scale(-1)
	}

	p := model3d.Coord2D{X: c.X - e.X, Y: c.Y - e.Y}

	x := (axis2.Dot(p) + 1) / 2
	y := (axis1.Dot(p) + 1) / 2

	return e.Solid.Contains(model2d.Coord{X: x, Y: y})
}
