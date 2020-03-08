package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	DrawerWidth      = 6.0
	DrawerHeight     = 2.0
	DrawerDepth      = 6.0
	DrawerSlack      = 0.05
	DrawerThickness  = 0.4
	DrawerHoleRadius = 0.1

	DrawerCount = 3

	ContainerThickness      = 0.2
	ContainerFootWidth      = 0.6
	ContainerFootHeight     = 0.2
	ContainerFootRampHeight = ContainerFootWidth / 2

	RidgeDepth = 0.2
)

func main() {
	container := CreateContainer()
	shelf := CreateShelf(container)

	log.Println("Creating container mesh...")
	mesh := model3d.SolidToMesh(container, 0.02, 0, -1, 5)
	log.Println("Eliminating co-planar polygons...")
	mesh = mesh.EliminateCoplanar(1e-8)
	log.Println("Saving container mesh...")
	mesh.SaveGroupedSTL("container.stl")
	log.Println("Rendering container mesh...")
	model3d.SaveRandomGrid("container.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)

	log.Println("Creating shelf mesh...")
	mesh = model3d.SolidToMesh(shelf, 0.02, 0, -1, 5)
	log.Println("Eliminating co-planar polygons...")
	mesh = mesh.EliminateCoplanar(1e-8)
	log.Println("Saving shelf mesh...")
	mesh.SaveGroupedSTL("shelf.stl")
	log.Println("Rendering shelf mesh...")
	model3d.SaveRandomGrid("shelf.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)
}

func CreateShelf(container model3d.Solid) model3d.Solid {
	min := model3d.Coord3D{
		X: DrawerSlack,
		Y: 0,
		Z: ContainerThickness + DrawerSlack,
	}
	max := model3d.Coord3D{
		X: DrawerWidth - DrawerSlack,
		Y: DrawerDepth - DrawerSlack,
		Z: ContainerThickness + DrawerHeight - DrawerSlack,
	}

	result := model3d.JoinedSolid{
		// Bottom face.
		&model3d.RectSolid{
			MinVal: min,
			MaxVal: model3d.Coord3D{X: max.X, Y: max.Y,
				Z: ContainerThickness + DrawerThickness},
		},
	}

	// Side faces.
	for _, x := range []float64{min.X, max.X - DrawerThickness} {
		result = append(result, &model3d.RectSolid{
			MinVal: model3d.Coord3D{X: x, Y: min.Y, Z: min.Z},
			MaxVal: model3d.Coord3D{X: x + DrawerThickness, Y: max.Y, Z: max.Z},
		})
	}

	// Front/back faces.
	for _, y := range []float64{min.Y, max.Y - DrawerThickness} {
		result = append(result, &model3d.RectSolid{
			MinVal: model3d.Coord3D{X: min.X, Y: y, Z: min.Z},
			MaxVal: model3d.Coord3D{X: max.X, Y: y + DrawerThickness, Z: max.Z},
		})
	}

	holeStart := min.Mid(max)
	holeStart.Y = min.Y - 0.1

	return &model3d.SubtractedSolid{
		Positive: result,
		Negative: model3d.JoinedSolid{
			container,
			&model3d.CylinderSolid{
				P1:     holeStart,
				P2:     holeStart.Add(model3d.Coord3D{Y: DrawerThickness + 0.2}),
				Radius: DrawerHoleRadius,
			},
		},
	}
}

func CreateContainer() model3d.Solid {
	var solid model3d.JoinedSolid

	// Side walls.
	for _, x := range []float64{-ContainerThickness, DrawerWidth} {
		solid = append(solid, &model3d.RectSolid{
			MinVal: model3d.Coord3D{X: x, Z: -ContainerThickness},
			MaxVal: model3d.Coord3D{
				X: x + ContainerThickness,
				Y: DrawerDepth + ContainerThickness,
				Z: DrawerCount*DrawerHeight + ContainerThickness,
			},
		})
	}

	// Top/bottom walls.
	for _, z := range []float64{-ContainerThickness, DrawerHeight * DrawerCount} {
		solid = append(solid, &model3d.RectSolid{
			MinVal: model3d.Coord3D{X: -ContainerThickness, Z: z},
			MaxVal: model3d.Coord3D{
				X: DrawerWidth + ContainerThickness,
				Y: DrawerDepth + ContainerThickness,
				Z: z + ContainerThickness,
			},
		})
	}

	// Back wall.
	wallMin := solid.Min()
	wallMin.Y = solid.Max().Y - ContainerThickness
	solid = append(solid, &model3d.RectSolid{
		MinVal: wallMin,
		MaxVal: solid.Max(),
	})

	// Ridges for shelves.
	for i := 0; i < DrawerCount; i++ {
		for _, right := range []bool{false, true} {
			solid = append(solid, CreateRidge((float64(i)+0.5)*DrawerHeight, right))
		}
	}

	footXs := []float64{solid.Min().X + ContainerFootWidth/2, solid.Max().X - ContainerFootWidth/2}
	footYs := []float64{solid.Min().Y + ContainerFootWidth/2, solid.Max().Y - ContainerFootWidth/2}
	for _, x := range footXs {
		for _, y := range footYs {
			solid = append(solid, CreateFoot(x, y))
		}
	}

	return solid
}

func CreateRidge(z float64, onRight bool) model3d.Solid {
	if !onRight {
		return &RidgeSolid{X1: 0, X2: RidgeDepth, Z: z}
	} else {
		return &RidgeSolid{X1: DrawerWidth, X2: DrawerWidth - RidgeDepth, Z: z}
	}
}

type RidgeSolid struct {
	X1 float64
	X2 float64
	Z  float64
}

func (r *RidgeSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: math.Min(r.X1, r.X2), Y: 0, Z: r.Z - RidgeDepth}
}

func (r *RidgeSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: math.Max(r.X1, r.X2), Y: DrawerDepth + ContainerThickness,
		Z: r.Z + RidgeDepth}
}

func (r *RidgeSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InSolidBounds(r, c) {
		return false
	}
	return math.Abs(c.Z-r.Z) <= math.Abs(c.X-r.X2)
}

func CreateFoot(x, y float64) model3d.Solid {
	center := model3d.Coord3D{X: x, Y: y, Z: -ContainerThickness}
	halfSize := model3d.Coord3D{X: ContainerFootWidth / 2, Y: ContainerFootWidth / 2}
	return &toolbox3d.Ramp{
		Solid: &model3d.RectSolid{
			MinVal: center.Sub(halfSize).Sub(model3d.Coord3D{Z: ContainerFootHeight}),
			MaxVal: center.Add(halfSize),
		},
		P1: center.Sub(model3d.Coord3D{Z: ContainerFootRampHeight}),
		P2: center,
	}
}
