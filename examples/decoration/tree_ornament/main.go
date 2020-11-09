package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
)

func main() {
	shape := Create2DShape(6)
	model2d.Rasterize("star.png", shape, 100)
}

func Create2DShape(sides int) model2d.Solid {
	var points []model2d.Coord
	for i := 0; i < sides; i++ {
		theta := float64(i) / float64(sides) * math.Pi * 2
		p := model2d.XY(math.Cos(theta), math.Sin(theta))
		points = append(points, p)
	}

	var segments []*model2d.Segment
	for i := 0; i < len(points); i++ {
		segments = append(segments, &model2d.Segment{points[i], points[(i+1)%len(points)]})
	}
	baseMesh := model2d.NewMeshSegments(segments)

	// Create points by extending sides until they intersect.
	meshes := []*model2d.Mesh{baseMesh}
	for i, first := range segments {
		next := segments[(i+2)%len(segments)]
		extended := ExtendSegment(first)
		ray := &model2d.Ray{Origin: next[1], Direction: next[0].Sub(next[1])}
		collision, ok := extended.FirstRayCollision(ray)
		if !ok {
			panic("no collision found")
		}
		tip := next[1].Add(next[0].Sub(next[1]).Scale(collision.Scale))
		meshes = append(meshes, model2d.NewMeshSegments([]*model2d.Segment{
			{points[(i+1)%len(points)], tip},
			{tip, points[(i+2)%len(points)]},
			{points[(i+1)%len(points)], points[(i+2)%len(points)]},
		}))
	}

	var solid model2d.JoinedSolid
	for _, mesh := range meshes {
		solid = append(solid, model2d.NewColliderSolid(model2d.MeshToCollider(mesh)))
	}

	return solid
}

func ExtendSegment(seg *model2d.Segment) *model2d.Segment {
	dir := seg[1].Sub(seg[0]).Scale(1000)
	return &model2d.Segment{
		seg[0].Sub(dir),
		seg[1].Add(dir),
	}
}
