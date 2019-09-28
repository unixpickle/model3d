package main

import (
	"io/ioutil"

	"github.com/unixpickle/model3d"
)

func main() {
	m := model3d.NewMeshPolar(func(g model3d.GeoCoord) float64 {
		return 1.0
	}, 50)
	collider := model3d.MeshToCollider(m)
	solid := model3d.NewColliderSolid(collider)
	m1 := model3d.SolidToMesh(solid, 0.1, 2, 0.8, 2)
	ioutil.WriteFile("sphere.stl", m1.EncodeSTL(), 0755)
}
