package main

import (
	"math"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/render3d"
)

type DomeObject struct {
	render3d.Object
}

func NewDomeObject() *DomeObject {
	return &DomeObject{
		&render3d.ColliderObject{
			Collider: &model3d.Sphere{
				Radius: RoomRadius,
			},
		},
	}
}

func (d *DomeObject) Cast(r *model3d.Ray) (model3d.RayCollision, render3d.Material, bool) {
	rc, _, ok := d.Object.Cast(r)
	rc.Normal = rc.Normal.Scale(-1)

	// Make the brightness non-uniform (simulating the sun)
	// so that shadows go in a certain direction and curves
	// are more visible in objects.
	brightest := LightDirection.Scale(-1)
	brightness := 0.2 + 9.0*math.Pow(math.Max(0, rc.Normal.Dot(brightest)), 10.0)
	mat := &render3d.LambertMaterial{
		DiffuseColor:  render3d.NewColorRGB(0.5, 0.8, 0.95).Scale(0.2),
		EmissionColor: render3d.NewColor(brightness),
	}

	return rc, mat, ok
}

type FloorObject struct {
	render3d.Object
}

func NewFloorObject() *FloorObject {
	return &FloorObject{
		Object: &render3d.ColliderObject{
			Collider: &model3d.Rect{
				MinVal: model3d.Coord3D{X: -RoomRadius, Y: -RoomRadius, Z: -0.01},
				MaxVal: model3d.Coord3D{X: RoomRadius, Y: RoomRadius, Z: 0},
			},
		},
	}
}

func (f *FloorObject) Cast(r *model3d.Ray) (model3d.RayCollision, render3d.Material, bool) {
	rc, mat, ok := f.Object.Cast(r)
	if !ok {
		return rc, mat, ok
	}
	p := r.Origin.Add(r.Direction.Scale(rc.Scale))
	color := 0.9
	if int(math.Mod(p.X+300, 2)) == int(math.Mod(p.Y+301, 2)) {
		color = 0.2
	}
	mat = &render3d.LambertMaterial{
		DiffuseColor: render3d.NewColor(color),
	}
	return rc, mat, ok
}
