package main

import (
	"log"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	Production = false
	AttachLamp = false
)

func main() {
	log.Println("Creating lamp...")
	light := NewLampLight()

	log.Println("Creating scene...")
	solid, colorFunc := CreateScene()

	log.Println("Creating final solid...")
	var hollowLight model3d.Solid
	if AttachLamp {
		hollowLight = light.Solid
	} else {
		hollowLight = &model3d.SubtractedSolid{
			Positive: light.Solid,
			Negative: model3d.JoinedSolid{
				model3d.NewColliderSolidInset(light.Object, 0.15),
				&model3d.Cylinder{P1: model3d.Z(2.3), P2: model3d.YZ(2, 2.3), Radius: 0.15},
			},
		}
	}

	solid = model3d.JoinedSolid{
		hollowLight,
		solid,
	}

	log.Println("Creating final mesh...")
	delta := 0.02
	if Production {
		delta = 0.01
	}
	fullMesh := model3d.MarchingCubesSearch(solid, delta, 8)

	log.Println("Recoloring scene...")
	colorFunc = light.Recolor(fullMesh, colorFunc)

	log.Println("Rendering...")
	obj := &diffuseObject{
		Object:    &render3d.ColliderObject{Collider: model3d.MeshToCollider(fullMesh)},
		ColorFunc: colorFunc.RenderColor,
	}
	render3d.SaveRendering("rendering.png", obj, model3d.XYZ(-1.0, -8.0, 3.0), 512, 512, nil)

	log.Println("Saving...")
	fullMesh.SaveQuantizedMaterialOBJ("lamp.zip", 32, colorFunc.Cached().TriangleColor)
}

func CreateScene() (model3d.Solid, toolbox3d.CoordColorFunc) {
	lampBase := model3d.JoinedSolid{
		model3d.NewRect(model3d.XYZ(-0.7, -0.2, 0.0), model3d.XYZ(0.7, 0.2, 1.7)),
		model3d.NewRect(model3d.XYZ(-0.7, -0.5, 0.0), model3d.XYZ(0.7, 0.5, 0.4)),
	}
	plant, plantColor := CreatePlant()
	base := model3d.NewRect(model3d.XYZ(-3.0, -1.5, -0.2), model3d.XYZ(1.0, 1.5, 0.001))
	if AttachLamp {
		lampBase[0].(*model3d.Rect).MaxVal.Z = 2.0
	}
	return model3d.JoinedSolid{lampBase, plant, base}, toolbox3d.JoinedCoordColorFunc(
		model3d.MeshToSDF(model3d.MarchingCubesSearch(lampBase, 0.02, 8)),
		render3d.NewColor(1.0),
		model3d.MeshToSDF(model3d.MarchingCubesSearch(plant, 0.02, 8)),
		plantColor,
		model3d.MeshToSDF(model3d.NewMeshRect(base.MinVal, base.MaxVal)),
		render3d.NewColor(1.0),
	)
}

type diffuseObject struct {
	render3d.Object
	ColorFunc render3d.ColorFunc
}

func (c *diffuseObject) Cast(r *model3d.Ray) (model3d.RayCollision, render3d.Material, bool) {
	rc, mat, ok := c.Object.Cast(r)
	if ok && c.ColorFunc != nil {
		p := r.Origin.Add(r.Direction.Scale(rc.Scale))
		color := c.ColorFunc(p, rc)
		mat = &render3d.LambertMaterial{
			DiffuseColor: color.Scale(0.9),
			AmbientColor: color.Scale(0.1),
		}
	}
	return rc, mat, ok
}
