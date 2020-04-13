package main

import (
	"github.com/unixpickle/model3d"
)

func main() {
	m := model3d.NewMeshPolar(func(g model3d.GeoCoord) float64 {
		return 1.0
	}, 50)
	collider := model3d.MeshToCollider(m)
	solid := model3d.NewColliderSolid(collider)
	m1 := model3d.MarchingCubesSearch(solid, 0.025, 8)
	m1.SaveGroupedSTL("sphere.stl")
}
