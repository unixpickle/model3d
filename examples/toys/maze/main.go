package main

import (
	"log"
	"sort"
	"sync"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	TotalSize    = 5.0
	WallHeight   = 0.8
	BottomHeight = 0.1
	BallRadius   = 0.1
)

func main() {
	bmp := model2d.MustReadBitmap("maze.png", nil)
	mesh := bmp.FlipY().Mesh().SmoothSq(70).MapCoords(func(c model2d.Coord) model2d.Coord {
		return c.Scale(TotalSize / float64(bmp.Width))
	})
	collider := model2d.MeshToCollider(mesh)

	solid := &MazeSolid{Collider: collider}
	log.Println("Creating mesh...")
	ax := &toolbox3d.AxisSqueeze{
		Axis:  toolbox3d.AxisZ,
		Min:   BottomHeight + 0.05,
		Max:   WallHeight - 0.05,
		Ratio: 0.1,
	}
	mesh3d := model3d.MarchingCubesSearch(model3d.TransformSolid(ax, solid), 0.0125, 8)
	mesh3d = mesh3d.MapCoords(ax.Inverse().Apply)
	mesh3d = mesh3d.EliminateCoplanar(1e-8)
	log.Println("Saving mesh...")
	mesh3d.SaveGroupedSTL("maze.stl")
	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh3d, 3, 3, 300, nil)

	log.Println("Creating ball...")
	sphere := &model3d.Sphere{Radius: BallRadius}
	mesh3d = model3d.MarchingCubesSearch(sphere, 0.01, 8)
	mesh3d.SaveGroupedSTL("ball.stl")

}

type MazeSolid struct {
	Collider model2d.Collider
	YCache   sync.Map
}

func (m *MazeSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{}
}

func (m *MazeSolid) Max() model3d.Coord3D {
	max := m.Collider.Max()
	return model3d.XYZ(max.X, max.Y, WallHeight)
}

func (m *MazeSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(m, c) {
		return false
	}
	var intersections []float64
	if ints, ok := m.YCache.Load(c.Y); ok {
		intersections = ints.([]float64)
	} else {
		r := &model2d.Ray{
			Origin:    model2d.Coord{Y: c.Y},
			Direction: model2d.Coord{X: 1},
		}
		m.Collider.RayCollisions(r, func(rc model2d.RayCollision) {
			p := r.Origin.X + r.Direction.X*rc.Scale
			intersections = append(intersections, p)
		})
		sort.Float64s(intersections)
		m.YCache.Store(c.Y, intersections)
	}

	if len(intersections) == 0 {
		return false
	} else if c.X < intersections[0] || c.X > intersections[len(intersections)-1] {
		return false
	}
	for i := 0; i < len(intersections)-1; i += 2 {
		x := intersections[i]
		if x > c.X {
			break
		}
		if intersections[i+1] > c.X {
			return c.Z < WallHeight
		}
	}
	return c.Z < BottomHeight
}
