package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const Thickness = 0.1

func main() {
	polytope := model3d.NewConvexPolytopeRect(model3d.XYZ(-1, -1, -1), model3d.XYZ(1, 1, 1))
	for _, x := range []float64{-1, 1} {
		for _, y := range []float64{-1, 1} {
			for _, z := range []float64{-1, 1} {
				vec := model3d.XYZ(x, y, z)
				polytope = append(polytope, &model3d.LinearConstraint{
					Normal: vec,
					Max:    2,
				})
			}
		}
	}
	mesh := polytope.Mesh()

	// Remove top face.
	mesh.Iterate(func(t *model3d.Triangle) {
		if t.Normal().Z > 0.99 {
			mesh.Remove(t)
		}
	})

	// Add inner mesh.
	mesh.AddMesh(mesh.Scale(1 - Thickness))

	// Connect openings of inner and outer meshes.
	mesh.Iterate(func(t *model3d.Triangle) {
		for _, seg := range t.Segments() {
			if len(mesh.Find(seg[0], seg[1])) == 1 && seg[0].Z > 0.999 {
				// This segment needs a connection and is on
				// the larger model.
				inner0 := seg[0].Scale(1 - Thickness)
				inner1 := seg[1].Scale(1 - Thickness)
				mesh.AddQuad(seg[0], seg[1], inner1, inner0)
			}
		}
	})

	// We didn't pay attention to normals when we connected
	// the inner and outer meshes, so we must repair them.
	mesh, _ = mesh.RepairNormals(1e-5)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("cut_cube.stl")

	log.Println("Rendering...")
	SaveRendering(mesh)
}

func SaveRendering(mesh *model3d.Mesh) {
	obj := &render3d.ColliderObject{
		Collider: model3d.MeshToCollider(mesh),
		Material: &render3d.PhongMaterial{
			Alpha:         10.0,
			DiffuseColor:  render3d.NewColor(0.4),
			SpecularColor: render3d.NewColor(0.25),
		},
	}
	light := render3d.JoinAreaLights(
		render3d.NewSphereAreaLight(&model3d.Sphere{
			Center: model3d.XYZ(-5, -10, 10),
			Radius: 1.0,
		}, render3d.NewColor(120.0)),
		render3d.NewSphereAreaLight(&model3d.Sphere{
			Center: model3d.XYZ(5, -10, 10),
			Radius: 1.0,
		}, render3d.NewColor(120.0)),
	)
	camera := render3d.NewCameraAt(
		model3d.XYZ(2, -4, 4).Scale(0.8),
		model3d.Coord3D{},
		math.Pi/3.6,
	)
	renderer := &render3d.BidirPathTracer{
		Camera:         camera,
		Light:          light,
		MaxDepth:       10,
		MinDepth:       2,
		NumSamples:     1000,
		RouletteDelta:  0.2,
		Antialias:      1.0,
		PowerHeuristic: 2,
	}
	img := render3d.NewImage(400, 400)
	renderer.Render(img, render3d.JoinedObject{obj, light})
	img.Save("rendering.png")
}
