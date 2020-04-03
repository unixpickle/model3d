package main

import (
	"fmt"
	"image/png"
	"math"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func main() {
	r, err := os.Open("../../decoration/globe/map.png")
	essentials.Must(err)
	defer r.Close()
	mapImage, err := png.Decode(r)
	essentials.Must(err)

	scene := render3d.JoinedObject{
		&Globe{
			Image: toolbox3d.NewEquirect(mapImage),
			Base: &render3d.Sphere{
				Center: model3d.Coord3D{},
				Radius: 1,
			},
		},
		CreateWalls(),
		CreateTable(),
		CreateLight(),
	}

	renderer := render3d.RecursiveRayTracer{
		Camera: render3d.NewCameraAt(model3d.Coord3D{Y: -13, Z: 2},
			model3d.Coord3D{Y: 0, Z: 2}, math.Pi/3.6),

		// Focus towards the area light.
		FocusPoints: []render3d.FocusPoint{
			&render3d.SphereFocusPoint{
				Center: model3d.Coord3D{Z: 5},
				Radius: 1,
			},
		},
		FocusPointProbs: []float64{0.5},

		MaxDepth:   5,
		NumSamples: 200,
		Antialias:  1.0,
		Cutoff:     1e-4,

		LogFunc: func(p, samples float64) {
			fmt.Printf("\rRendering %.1f%%...", p*100)
		},
	}

	fmt.Println("Ray variance:", renderer.RayVariance(scene, 200, 200, 5))

	img := render3d.NewImage(200, 200)
	renderer.Render(img, scene)
	fmt.Println()
	img.Save("output.png")
}

type Globe struct {
	Image *toolbox3d.Equirect
	Base  render3d.Object
}

func (g *Globe) Cast(r *model3d.Ray) (model3d.RayCollision, render3d.Material, bool) {
	collision, material, ok := g.Base.Cast(r)
	if !ok {
		return collision, material, ok
	}
	point := r.Origin.Add(r.Direction.Scale(collision.Scale))

	point = model3d.NewMatrix3Rotation(model3d.Coord3D{Z: 1}, -math.Pi/2).MulColumn(point)
	point.Y, point.Z = point.Z, -point.Y

	red, green, blue, _ := g.Image.At(point.Geo()).RGBA()
	material = &render3d.PhongMaterial{
		Alpha:         5,
		SpecularColor: render3d.Color{X: 0.1, Y: 0.1, Z: 0.1},
		DiffuseColor: render3d.Color{X: float64(red) / 0xffff, Y: float64(green) / 0xffff,
			Z: float64(blue) / 0xffff},
	}
	return collision, material, ok
}

func CreateWalls() render3d.Object {
	mesh := model3d.NewMeshRect(model3d.Coord3D{X: -6, Y: -15, Z: -5},
		model3d.Coord3D{X: 6, Y: 8, Z: 5})

	// Face normals inward.
	mesh = mesh.MapCoords(model3d.Coord3D{X: -1, Y: 1, Z: 1}.Mul)

	return &render3d.ColliderObject{
		Collider: model3d.MeshToCollider(mesh),
		Material: &render3d.LambertMaterial{
			DiffuseColor: render3d.Color{X: 1, Y: 1, Z: 1},
		},
	}
}

func CreateTable() render3d.Object {
	createPiece := func(min, max model3d.Coord3D) render3d.Object {
		return &render3d.ColliderObject{
			Collider: model3d.MeshToCollider(model3d.NewMeshRect(min, max)),
			Material: &render3d.LambertMaterial{
				DiffuseColor: render3d.Color{X: 0.75, Y: 0.25, Z: 0},
			},
		}
	}

	return render3d.JoinedObject{
		createPiece(model3d.Coord3D{X: -3, Y: -2, Z: -1.2},
			model3d.Coord3D{X: 3, Y: 2, Z: -1}),
		createPiece(model3d.Coord3D{X: -3, Y: -2, Z: -5},
			model3d.Coord3D{X: -2.8, Y: 2, Z: -1.2}),
		createPiece(model3d.Coord3D{X: 2.8, Y: -2, Z: -5},
			model3d.Coord3D{X: 3, Y: 2, Z: -1.2}),
	}
}

func CreateLight() render3d.Object {
	return &render3d.Sphere{
		Center: model3d.Coord3D{Z: 6},
		Radius: 1.3,
		Material: &render3d.LambertMaterial{
			EmissionColor: render3d.Color{X: 1, Y: 1, Z: 1}.Scale(200),
		},
	}
}
