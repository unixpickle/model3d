package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func NewLightObject() render3d.Object {
	return &render3d.ColliderObject{
		Collider: &model3d.Sphere{
			Center: LightCenter,
			Radius: LightRadius,
		},
		Material: &render3d.LambertMaterial{
			EmissionColor: render3d.NewColor(LightBrightness),
		},
	}
}

type DomeObject struct {
	render3d.Object
}

func NewDomeObject() *DomeObject {
	skyColor := render3d.NewColorRGB(0.5, 0.8, 0.95).Scale(0.15).Add(render3d.NewColor(0.15))
	return &DomeObject{
		&render3d.ColliderObject{
			Collider: &model3d.Sphere{
				Radius: RoomRadius,
			},
			Material: &render3d.LambertMaterial{
				DiffuseColor: skyColor,
			},
		},
	}
}

func (d *DomeObject) Cast(r *model3d.Ray) (model3d.RayCollision, render3d.Material, bool) {
	rc, mat, ok := d.Object.Cast(r)
	rc.Normal = rc.Normal.Scale(-1)
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
	color := 0.6
	if int(math.Mod(p.X+300, 2)) == int(math.Mod(p.Y+301, 2)) {
		color = 0.1
	}
	mat = &render3d.LambertMaterial{
		DiffuseColor: render3d.NewColor(color),
	}
	return rc, mat, ok
}
