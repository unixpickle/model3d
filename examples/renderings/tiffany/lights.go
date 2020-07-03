package main

import (
	"math"
	"math/rand"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	CeilingLightBrightness = 150.0
	CeilingLightDepth      = 0.4
	CeilingLightRadius     = 0.4

	LampBrightness      = 5.0
	WallLightBrightness = 5.0
)

func CreateLamp() render3d.AreaLight {
	var lights []render3d.AreaLight
	for _, x := range []float64{-4, 4} {
		light := render3d.NewCylinderAreaLight(&model3d.Cylinder{
			P1:     model3d.XYZ(x, 6, 0),
			P2:     model3d.XYZ(x, 6, 2),
			Radius: 0.8,
		}, render3d.NewColor(LampBrightness))

		fullLamp := render3d.JoinedObject{
			light,

			// Base
			&render3d.ColliderObject{
				Collider: &model3d.Cylinder{
					P1:     model3d.XYZ(x, 6, -5),
					P2:     model3d.XYZ(x, 6, -4.9),
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
					P1:     model3d.XYZ(x, 6, -4.9),
					P2:     model3d.XYZ(x, 6, 0),
					Radius: 0.2,
				},
				Material: &render3d.PhongMaterial{
					Alpha:         10,
					SpecularColor: render3d.NewColor(0.5),
					DiffuseColor:  render3d.NewColor(0.5),
				},
			},
		}
		lights = append(lights, &lightedObject{
			AreaLight: light,
			Object:    fullLamp,
		})
	}
	return render3d.JoinAreaLights(lights...)
}

func CreateWallLights() render3d.AreaLight {
	mesh := model3d.NewMesh()
	mesh.AddMesh(model3d.NewMeshRect(
		model3d.XYZ(-6, -4, 3.5),
		model3d.XYZ(-5.6, 4, 4),
	))
	mesh.AddMesh(model3d.NewMeshRect(
		model3d.XYZ(5.6, -4, 3.5),
		model3d.XYZ(6, 4, 4),
	))
	return render3d.NewMeshAreaLight(mesh, render3d.NewColor(WallLightBrightness))
}

type CeilingLight struct {
	Cylinder *model3d.Cylinder
}

func CreateCeilingLights() []*CeilingLight {
	var lights []*CeilingLight
	for x := -3; x <= 3; x += 3 {
		for y := -5; y <= 4; y += 3 {
			lights = append(lights, &CeilingLight{
				Cylinder: &model3d.Cylinder{
					P1: model3d.XYZ(float64(x), float64(y), RoomHeight),
					P2: model3d.Coord3D{X: float64(x), Y: float64(y),
						Z: RoomHeight + CeilingLightDepth},
					Radius: CeilingLightRadius,
				},
			})
		}
	}
	return lights
}

func (c *CeilingLight) Min() model3d.Coord3D {
	return c.Cylinder.Min()
}

func (c *CeilingLight) Max() model3d.Coord3D {
	return c.Cylinder.Max()
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
		DiffuseColor: render3d.NewColor(0.025),
	}
	if coll.Normal.Z < -0.5 {
		mat.EmissionColor = render3d.NewColor(CeilingLightBrightness)
	}
	return coll, mat, found
}

func (c *CeilingLight) SampleLight(gen *rand.Rand) (p, n model3d.Coord3D, e render3d.Color) {
	theta := gen.Float64() * math.Pi * 2
	dir := model3d.Coord3D{X: math.Cos(theta), Y: math.Sin(theta)}
	coord := c.Cylinder.P2.Add(dir.Scale(c.Cylinder.Radius))
	return coord, model3d.Coord3D{Z: -1}, render3d.NewColor(CeilingLightBrightness)
}

func (c *CeilingLight) TotalEmission() float64 {
	return 3 * CeilingLightBrightness * math.Pi * math.Pow(c.Cylinder.Radius, 2)
}

func (c *CeilingLight) Cut(r *model3d.Ray, rc model3d.RayCollision) bool {
	p := r.Origin.Add(r.Direction.Scale(rc.Scale))
	return math.Abs(p.Z-c.Cylinder.P1.Z) < 1e-8 &&
		p.Dist(c.Cylinder.P1) < c.Cylinder.Radius
}

// lightedObject is an object with some part that emits
// light.
type lightedObject struct {
	render3d.AreaLight
	Object render3d.Object
}

func (l *lightedObject) Cast(r *model3d.Ray) (model3d.RayCollision, render3d.Material, bool) {
	return l.Object.Cast(r)
}
