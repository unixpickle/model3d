package main

import (
	"github.com/unixpickle/model3d"
)

func main() {
	m := model3d.SolidToMesh(HersheyKissSolid{}, 0.01, 0, -1, 5)
	m.SaveGroupedSTL("kiss.stl")
}
