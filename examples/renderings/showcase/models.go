package main

import (
	"compress/gzip"
	"math"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func ReadCurvyThing() render3d.Object {
	mesh := ReadModel("models/curvy_thing.stl.gz")
	invMid := mesh.Min().Mid(mesh.Max()).Scale(-1)
	invMid.Z = -mesh.Min().Z
	mesh = mesh.Translate(invMid)
	mesh = mesh.Rotate(model3d.Z(1), -math.Pi/4)
	mesh = mesh.Translate(model3d.Coord3D{X: CurvyThingX, Y: CurvyThingY})

	return &render3d.ColliderObject{
		Collider: model3d.MeshToCollider(mesh),
		Material: &render3d.PhongMaterial{
			Alpha:         20.0,
			SpecularColor: render3d.NewColor(0.05),
			DiffuseColor:  render3d.NewColor(0.3),
		},
	}
}

func ReadRose() render3d.Object {
	mesh := ReadModel("models/rose.stl.gz")
	mesh = mesh.Translate(mesh.Min().Mid(mesh.Max()).Scale(-1))
	mesh = mesh.Rotate(model3d.X(1), math.Pi/4)
	mesh = mesh.Translate(model3d.XYZ(RoseX, RoseY, RoseZ))

	return render3d.JoinedObject{
		&render3d.ColliderObject{
			Collider: model3d.MeshToCollider(mesh),
			Material: &render3d.LambertMaterial{
				DiffuseColor: render3d.NewColorRGB(0.95, 0.2, 0.2).Scale(0.5),
			},
		},
		// Manually add a stem so the rose is not floating.
		&render3d.ColliderObject{
			Collider: &model3d.Cylinder{
				P1:     model3d.XYZ(VaseX, VaseY, 0),
				P2:     model3d.XYZ(VaseX, VaseY, RoseZ),
				Radius: RoseStemRadius,
			},
			Material: &render3d.LambertMaterial{
				DiffuseColor: render3d.NewColorRGB(0.1, 0.55, 0).Scale(0.5),
			},
		},
	}
}

func ReadVase() render3d.Object {
	mesh := ReadModel("models/vase.stl.gz")

	// The mesh is too big compared to the other objects.
	mesh = mesh.MapCoords(model3d.XYZ(0.8, 0.8, 0.8).Mul)

	min, max := mesh.Min(), mesh.Max()
	mesh = mesh.MapCoords(model3d.Coord3D{
		X: VaseX - (max.X+min.X)/2,
		Y: VaseY - (max.Y+min.Y)/2,
	}.Add)
	return &VaseObject{
		Object: &render3d.ColliderObject{
			Collider: model3d.MeshToCollider(mesh),
		},
		maxZ: max.Z,
	}
}

type VaseObject struct {
	render3d.Object
	maxZ float64
}

func (v *VaseObject) Cast(r *model3d.Ray) (model3d.RayCollision, render3d.Material, bool) {
	rc, _, ok := v.Object.Cast(r)
	if !ok {
		return rc, nil, ok
	}

	c := r.Origin.Add(r.Direction.Scale(rc.Scale))
	frac := c.Z / v.maxZ
	c1 := render3d.NewColorRGB(61.0/255.0, 222.0/255.0, 33.0/255.0)
	c2 := render3d.NewColorRGB(198.0/255.0, 52.0/255.0, 235.0/255.0)
	color := c1.Scale(frac).Add(c2.Scale(1 - frac))
	mat := &render3d.PhongMaterial{
		Alpha:         10.0,
		DiffuseColor:  color.Scale(0.4),
		SpecularColor: render3d.NewColor(0.1),
	}

	return rc, mat, ok
}

func ReadRocks() render3d.Object {
	mesh := ReadModel("models/rocks.stl.gz")
	min, max := mesh.Min(), mesh.Max()
	mesh = mesh.MapCoords(model3d.Coord3D{
		X: -(max.X + min.X) / 2,
		Y: RocksY - min.Y,
	}.Add)
	return &render3d.ColliderObject{
		Collider: model3d.MeshToCollider(mesh),
		Material: &render3d.LambertMaterial{
			DiffuseColor: render3d.NewColor(0.3),
		},
	}
}

func ReadPumpkin() render3d.Object {
	models := []string{
		"models/pumpkin_inside.stl.gz",
		"models/pumpkin_outside.stl.gz",
		"models/pumpkin_stem.stl.gz",
	}
	colors := []render3d.Color{
		render3d.NewColorRGB(255.0/255, 206.0/255, 107.0/255),
		render3d.NewColorRGB(214.0/255, 143.0/255, 0),
		render3d.NewColorRGB(79.0/255, 53.0/255, 0),
	}

	var parts render3d.JoinedObject
	for i, model := range models {
		color := colors[i]
		mesh := ReadModel(model)
		mesh = mesh.MapCoords(model3d.Coord3D{
			X: PumpkinX,
			Y: PumpkinY,
			Z: 1.1942578125000005,
		}.Add)
		parts = append(parts, &render3d.ColliderObject{
			Collider: model3d.MeshToCollider(mesh),
			Material: &render3d.LambertMaterial{
				DiffuseColor: color.Scale(0.5),
			},
		})
	}

	return parts
}

func ReadWineGlass() render3d.Object {
	mesh := ReadModel("models/wine_glass.stl.gz")
	min, max := mesh.Min(), mesh.Max()
	mesh = mesh.Translate(min.Scale(-1))
	mesh = mesh.MapCoords(model3d.Coord3D{
		X: WineGlassX - (max.X-min.X)/2,
		Y: WineGlassY - (max.Y-min.Y)/2,
		// Make sure rays can hit the ground under the model.
		Z: 1e-4,
	}.Add)

	return &render3d.ColliderObject{
		Collider: model3d.MeshToCollider(mesh),
		Material: &render3d.JoinedMaterial{
			Materials: []render3d.Material{
				&render3d.RefractMaterial{
					IndexOfRefraction: 1.3,
					RefractColor:      render3d.NewColor(0.95),
				},
				&render3d.PhongMaterial{
					Alpha:         100.0,
					SpecularColor: render3d.NewColor(0.05),
				},
			},
			Probs: []float64{0.8, 0.2},
		},
	}
}

func ReadModel(path string) *model3d.Mesh {
	r, err := os.Open(path)
	essentials.Must(err)
	defer r.Close()

	gr, err := gzip.NewReader(r)
	essentials.Must(err)
	defer gr.Close()

	tris, err := model3d.ReadSTL(gr)
	essentials.Must(err)

	return model3d.NewMeshTriangles(tris)
}
