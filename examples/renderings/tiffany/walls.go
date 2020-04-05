package main

import (
	"math"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/render3d"
)

var TiffanyBlue = render3d.Color{X: 129.0 / 255.0, Y: 216.0 / 255.0, Z: 208.0 / 255.0}

type Walls struct {
	Base *render3d.ColliderObject
}

func NewWalls() *Walls {
	mesh := model3d.NewMeshRect(model3d.Coord3D{X: -6, Y: -15, Z: -5},
		model3d.Coord3D{X: 6, Y: 8, Z: 5})

	// Face normals inward.
	mesh = mesh.MapCoords(model3d.Coord3D{X: -1, Y: 1, Z: 1}.Mul)

	return &Walls{
		Base: &render3d.ColliderObject{
			Collider: model3d.MeshToCollider(mesh),
			Material: &render3d.LambertMaterial{
				DiffuseColor: render3d.Color{X: 1, Y: 1, Z: 1},
			},
		},
	}
}

func (w *Walls) Cast(r *model3d.Ray) (model3d.RayCollision, render3d.Material, bool) {
	collision, material, ok := w.Base.Cast(r)
	if !ok {
		return collision, material, ok
	}
	point := r.Origin.Add(r.Direction.Scale(collision.Scale))

	if math.Abs(point.X-w.Base.Collider.Max().X) < 1e-8 {
		scale := 1.0
		if math.Abs(math.Mod(point.Z+point.Y+100, 1.0)-0.5) < 0.25 {
			scale = 1.2
		}
		material = &render3d.LambertMaterial{
			DiffuseColor: TiffanyBlue.Scale(scale),
		}
	} else if math.Abs(point.X-w.Base.Collider.Min().X) < 1e-8 {
		scale := 1.2
		zMod1 := math.Mod(point.Z+100, 2) - 1
		yMod1 := math.Mod(point.Y+100, 4) - 2
		zMod2 := math.Mod(point.Z+101, 2) - 1
		yMod2 := math.Mod(point.Y+102, 4) - 2
		if math.Sqrt(zMod1*zMod1+yMod1*yMod1) < 0.4 || math.Sqrt(zMod2*zMod2+yMod2*yMod2) < 0.4 {
			scale = 1.0
		}
		material = &render3d.LambertMaterial{
			DiffuseColor: TiffanyBlue.Scale(scale),
		}
	}

	return collision, material, ok
}
