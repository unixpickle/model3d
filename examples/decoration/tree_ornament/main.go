package main

import (
	"math"

	"github.com/unixpickle/model3d/render3d"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

const PointScale = 4.0

func main() {
	shape := Create2DShape(6)
	maxZ := shape.SDF(model2d.XY(0, 0))
	shape3D := &Shape3D{Star2D: shape, MaxZ: maxZ}
	mesh := model3d.MarchingCubesSearch(shape3D, 0.03, 8)
	mesh.SaveGroupedSTL("star.stl")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func Create2DShape(sides int) model2d.SDF {
	var points []model2d.Coord
	for i := 0; i < sides*2; i++ {
		theta := float64(i) / float64(sides*2) * math.Pi * 2
		p := model2d.XY(math.Cos(theta), math.Sin(theta))
		if i%2 == 0 {
			p = p.Scale(PointScale)
		}
		points = append(points, p)
	}

	var segments []*model2d.Segment
	for i := 0; i < len(points); i++ {
		segments = append(segments, &model2d.Segment{points[i], points[(i+1)%len(points)]})
	}
	baseMesh := model2d.NewMeshSegments(segments)
	return model2d.MeshToSDF(baseMesh)
}

type Shape3D struct {
	Star2D model2d.SDF
	MaxZ   float64
}

func (s *Shape3D) Min() model3d.Coord3D {
	min := s.Star2D.Min()
	return model3d.XYZ(min.X, min.Y, -s.MaxZ)
}

func (s *Shape3D) Max() model3d.Coord3D {
	max := s.Star2D.Max()
	return model3d.XYZ(max.X, max.Y, s.MaxZ)
}

func (s *Shape3D) Contains(c model3d.Coord3D) bool {
	return s.Star2D.SDF(c.XY()) > math.Abs(c.Z)
}
