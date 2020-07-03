package main

import (
	"archive/zip"
	"strconv"
	"strings"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func CreateCorgi(theta, x, y float64) render3d.Object {
	var model render3d.JoinedObject

	// Taken from all the parts together.
	min := model3d.XYZ(-0.36888350039080975, -0.3312187908572959, -0.5748622825455283)
	max := model3d.XYZ(1.7179133746091901, 0.33221870914270424, 1.0427549049544718)
	mid := min.Mid(max)

	r, err := zip.OpenReader("assets/corgi.zip")
	essentials.Must(err)
	defer r.Close()

	for _, f := range r.File {
		fr, err := f.Open()
		essentials.Must(err)
		defer fr.Close()
		triangles, err := model3d.ReadSTL(fr)
		essentials.Must(err)

		name := f.Name
		parts := strings.Split(name, ",")
		color := [3]float64{}
		for i, p := range parts {
			num, _ := strconv.Atoi(p)
			color[i] = float64(num) / 255.0
		}

		mesh := model3d.NewMeshTriangles(triangles)
		mesh = mesh.MapCoords(model3d.XYZ(-mid.X, -mid.Y, -min.Z).Add)
		mesh = mesh.MapCoords(model3d.NewMatrix3Rotation(model3d.Z(1), theta).MulColumn)
		mesh = mesh.Scale(0.8)
		mesh = mesh.MapCoords(model3d.Coord3D{X: x, Y: y}.Add)

		model = append(model, &render3d.ColliderObject{
			Collider: model3d.MeshToCollider(mesh),
			Material: &render3d.LambertMaterial{
				DiffuseColor: render3d.NewColorRGB(color[0], color[1], color[2]),
			},
		})
	}

	return model
}
