package main

import (
	"math"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	CeilingLightBrightness = 150.0
	CeilingLightDepth      = 0.4
	CeilingLightRadius     = 0.4

	LampBrightness      = 5.0
	WallLightBrightness = 5.0
)

func CreateLamp() render3d.Object {
	var res render3d.JoinedObject
	for _, x := range []float64{-4, 4} {
		res = append(res,
			// Base
			&render3d.ColliderObject{
				Collider: &model3d.Cylinder{
					P1:     model3d.Coord3D{X: x, Y: 6, Z: -5},
					P2:     model3d.Coord3D{X: x, Y: 6, Z: -4.9},
					Radius: 0.5,
				},
				Material: &render3d.PhongMaterial{
					Alpha:         10,
					SpecularColor: render3d.NewColor(0.5),
					DiffuseColor:  render3d.NewColor(0.5),
				},
			},
			// Pole
			&render3d.ColliderObject{
				Collider: &model3d.Cylinder{
					P1:     model3d.Coord3D{X: x, Y: 6, Z: -4.9},
					P2:     model3d.Coord3D{X: x, Y: 6, Z: 0},
					Radius: 0.2,
				},
				Material: &render3d.PhongMaterial{
					Alpha:         10,
					SpecularColor: render3d.NewColor(0.5),
					DiffuseColor:  render3d.NewColor(0.5),
				},
			},
			// Shade
			&render3d.ColliderObject{
				Collider: &model3d.Cylinder{
					P1:     model3d.Coord3D{X: x, Y: 6, Z: 0},
					P2:     model3d.Coord3D{X: x, Y: 6, Z: 2},
					Radius: 0.8,
				},
				Material: &render3d.LambertMaterial{
					EmissionColor: render3d.Color{X: 1, Y: 1, Z: 1}.Scale(LampBrightness),
				},
			},
		)
	}
	return res
}

func CreateWallLights() render3d.Object {
	mesh := model3d.NewMesh()
	mesh.AddMesh(model3d.NewMeshRect(
		model3d.Coord3D{X: -6, Y: -4, Z: 3.5},
		model3d.Coord3D{X: -5.6, Y: 4, Z: 4},
	))
	mesh.AddMesh(model3d.NewMeshRect(
		model3d.Coord3D{X: 5.6, Y: -4, Z: 3.5},
		model3d.Coord3D{X: 6, Y: 4, Z: 4},
	))
	return &render3d.ColliderObject{
		Collider: model3d.MeshToCollider(mesh),
		Material: &render3d.LambertMaterial{
			EmissionColor: render3d.Color{X: 1, Y: 1, Z: 1}.Scale(WallLightBrightness),
		},
	}
}

type CeilingLight struct {
	Cylinder *model3d.Cylinder
}

func (c *CeilingLight) Cast(r *model3d.Ray) (model3d.RayCollision, render3d.Material, bool) {
	var found bool
	var coll model3d.RayCollision
	c.Cylinder.RayCollisions(r, func(rc model3d.RayCollision) {
		p := r.Origin.Add(r.Direction.Scale(rc.Scale))
		if math.Abs(p.Z-c.Cylinder.P1.Z) < 1e-8 {
			return
		}
		if !found || rc.Scale < coll.Scale {
			found = true
			coll = rc
		}
	})
	if !found {
		return model3d.RayCollision{}, nil, false
	}
	coll.Normal = coll.Normal.Scale(-1)
	mat := &render3d.LambertMaterial{
		DiffuseColor: render3d.Color{X: 1, Y: 1, Z: 1}.Scale(0.05),
	}
	if coll.Normal.Z < -0.5 {
		mat.EmissionColor = render3d.Color{X: 1, Y: 1, Z: 1}.Scale(CeilingLightBrightness)
	}
	return coll, mat, found
}

func (c *CeilingLight) Cut(r *model3d.Ray, rc model3d.RayCollision) bool {
	p := r.Origin.Add(r.Direction.Scale(rc.Scale))
	return math.Abs(p.Z-c.Cylinder.P1.Z) < 1e-8 &&
		p.Dist(c.Cylinder.P1) < c.Cylinder.Radius
}
