package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const MaxThickness = 0.2

func main() {
	solid := model3d.JoinedSolid{}
	for _, curve := range TalonCurves() {
		solid = append(solid, NewTalon(curve))
	}
	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.01, 16)
	mesh = mesh.FlipDelaunay()
	log.Println("Smoothing...")
	mesh = mesh.SmoothAreas(0.01, 16)
	log.Println("Saving...")
	mesh.SaveGroupedSTL("starfish.stl")
	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func TalonCurves() []model2d.BezierCurve {
	return []model2d.BezierCurve{
		// Top
		{
			model2d.XY(0, 0),
			model2d.XY(0.1, 0.5),
			model2d.XY(-0.2, 1.0),
		},
		// Right and left.
		{
			model2d.XY(0, 0),
			model2d.XY(0.3, 0.0),
			model2d.XY(0.5, 0.1),
			model2d.XY(1.0, -0.2),
		},
		{
			model2d.XY(0, 0),
			model2d.XY(-0.3, 0.0),
			model2d.XY(-0.5, -0.1),
			model2d.XY(-1.0, 0.2),
		},
		// Bottom right and left
		{
			model2d.XY(0, 0),
			model2d.XY(0.5, -0.4),
			model2d.XY(0.5, -1.0),
		},
		{
			model2d.XY(0, 0),
			model2d.XY(-0.5, -0.4),
			model2d.XY(-0.5, -1.0),
		},
	}
}

type Talon struct {
	Mesh      *model2d.Mesh
	SDF       model2d.PointSDF
	Collider  model2d.Collider
	Dists     map[model2d.Coord]float64
	TotalDist float64
}

func NewTalon(curve model2d.BezierCurve) *Talon {
	mesh := model2d.NewMesh()
	dists := map[model2d.Coord]float64{curve.Eval(0): 0.0}
	dist := 0.0
	for t := 0.0; t < 1.0; t += 0.01 {
		c1 := curve.Eval(t)
		c2 := curve.Eval(t + 0.01)
		seg := &model2d.Segment{c1, c2}
		mesh.Add(seg)
		dist += seg.Length()
		dists[c2] = dist
	}
	return &Talon{
		Mesh:      mesh,
		SDF:       model2d.MeshToSDF(mesh),
		Collider:  model2d.MeshToCollider(mesh),
		Dists:     dists,
		TotalDist: dist,
	}
}

func (t *Talon) Min() model3d.Coord3D {
	min := t.Collider.Min()
	return model3d.XYZ(min.X-MaxThickness, min.Y-MaxThickness, -MaxThickness)
}

func (t *Talon) Max() model3d.Coord3D {
	max := t.Collider.Max()
	return model3d.XYZ(max.X+MaxThickness, max.Y+MaxThickness, MaxThickness)
}

func (t *Talon) Contains(c model3d.Coord3D) bool {
	proj, dist := t.projDist(c.XY())
	thickness := MaxThickness * (1 - dist/t.TotalDist)
	projDist := proj.Dist(c.XY())
	return math.Abs(projDist)+math.Abs(c.Z) < thickness
}

func (t *Talon) projDist(c model2d.Coord) (model2d.Coord, float64) {
	proj, _ := t.SDF.PointSDF(c)

	if proj.Norm() < 1e-5 {
		return proj, math.Inf(1)
	}

	// Even though proj is on the mesh, a collision with (proj-c)
	// might not exist due to rounding errors, so we retry with tiny
	// random perturbations.
	for i := 0; i < 20; i++ {
		ray := &model2d.Ray{Origin: c, Direction: proj.Sub(c)}
		if i > 0 {
			ray.Direction = ray.Direction.Add(model2d.NewCoordRandUnit().Scale(1e-8))
		}
		rc, ok := t.Collider.FirstRayCollision(ray)
		if !ok {
			continue
		}
		seg := rc.Extra.(*model2d.Segment)
		dist := proj.Dist(seg[0]) + t.Dists[seg[0]]
		return proj, dist
	}
	return proj, math.Inf(1)
}
