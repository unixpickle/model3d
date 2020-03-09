package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	ShelfWidth      = 6.0
	ShelfHeight     = 2.0
	ShelfDepth      = 6.0
	ShelfSlack      = 0.05
	ShelfThickness  = 0.4
	ShelfHoleRadius = 0.1

	ShelfCount = 3

	ContainerThickness      = 0.2
	ContainerFootWidth      = 0.6
	ContainerFootHeight     = 0.2
	ContainerFootRampHeight = ContainerFootWidth / 2

	RidgeDepth = 0.2
)

func main() {
	container := CreateContainer()
	shelf := CreateShelf()

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

func CreateShelf() model3d.Solid {
	min := model3d.Coord3D{
		X: ShelfSlack,
		Y: 0,
		Z: ContainerThickness + ShelfSlack,
	}
	max := model3d.Coord3D{
		X: ShelfWidth - ShelfSlack,
		Y: ShelfDepth - ShelfSlack,
		Z: ContainerThickness + ShelfHeight - ShelfSlack,
	}

	result := model3d.JoinedSolid{
		// Bottom face.
		&model3d.RectSolid{
			MinVal: min,
			MaxVal: model3d.Coord3D{X: max.X, Y: max.Y,
				Z: ContainerThickness + ShelfThickness},
		},
	}

	// Side faces.
	for _, x := range []float64{min.X, max.X - ShelfThickness} {
		result = append(result, &model3d.RectSolid{
			MinVal: model3d.Coord3D{X: x, Y: min.Y, Z: min.Z},
			MaxVal: model3d.Coord3D{X: x + ShelfThickness, Y: max.Y, Z: max.Z},
		})
	}

	// Front/back faces.
	for _, y := range []float64{min.Y, max.Y - ShelfThickness} {
		result = append(result, &model3d.RectSolid{
			MinVal: model3d.Coord3D{X: min.X, Y: y, Z: min.Z},
			MaxVal: model3d.Coord3D{X: max.X, Y: y + ShelfThickness, Z: max.Z},
		})
	}

	mid := min.Mid(max)

	return &model3d.SubtractedSolid{
		Positive: result,
		Negative: model3d.JoinedSolid{
			&RidgeSolid{X1: min.X, X2: min.X + RidgeDepth, Z: mid.Z},
			&RidgeSolid{X1: max.X, X2: max.X - RidgeDepth, Z: mid.Z},
			&HoleCutout{
				X:      mid.X,
				Z:      mid.Z,
				Y1:     min.Y - 1e-5,
				Y2:     min.Y + ShelfThickness + 1e-5,
				Radius: ShelfHoleRadius,
			},
		},
	}
}

type HoleCutout struct {
	X float64
	Z float64

	Y1     float64
	Y2     float64
	Radius float64
}

func (h *HoleCutout) Min() model3d.Coord3D {
	return model3d.Coord3D{X: h.X - h.Radius, Y: h.Y1, Z: h.Z - h.Radius}
}

func (h *HoleCutout) Max() model3d.Coord3D {
	return model3d.Coord3D{X: h.X + h.Radius, Y: h.Y2, Z: h.Z + h.Radius*2}
}

func (h *HoleCutout) Contains(c model3d.Coord3D) bool {
	if !model3d.InSolidBounds(h, c) {
		return false
	}
	c2d := model3d.Coord2D{X: c.X - h.X, Y: c.Z - h.Z}
	if c2d.Norm() <= h.Radius {
		return true
	}
	if c2d.Y < 0 {
		return false
	}
	// Pointed tip to avoid support.
	vec := model3d.Coord2D{X: 1, Y: 1}.Normalize()
	vec1 := vec.Mul(model3d.Coord2D{X: -1, Y: 1})
	return c2d.Dot(vec) <= h.Radius && c2d.Dot(vec1) <= h.Radius
}

func CreateContainer() model3d.Solid {
	var solid model3d.JoinedSolid

	// Side walls.
	for _, x := range []float64{-ContainerThickness, ShelfWidth} {
		solid = append(solid, &model3d.RectSolid{
			MinVal: model3d.Coord3D{X: x, Z: -ContainerThickness},
			MaxVal: model3d.Coord3D{
				X: x + ContainerThickness,
				Y: ShelfDepth + ContainerThickness,
				Z: ShelfCount*ShelfHeight + ContainerThickness,
			},
		})
	}

	// Top/bottom walls.
	for _, z := range []float64{-ContainerThickness, ShelfHeight * ShelfCount} {
		solid = append(solid, &model3d.RectSolid{
			MinVal: model3d.Coord3D{X: -ContainerThickness, Z: z},
			MaxVal: model3d.Coord3D{
				X: ShelfWidth + ContainerThickness,
				Y: ShelfDepth + ContainerThickness,
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
	for i := 0; i < ShelfCount; i++ {
		for _, right := range []bool{false, true} {
			solid = append(solid, CreateRidge((float64(i)+0.5)*ShelfHeight, right))
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
		return &RidgeSolid{X1: ShelfWidth, X2: ShelfWidth - RidgeDepth, Z: z}
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
	return model3d.Coord3D{X: math.Max(r.X1, r.X2), Y: ShelfDepth + ContainerThickness,
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
