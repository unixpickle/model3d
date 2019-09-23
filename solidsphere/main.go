package main

import (
	"io/ioutil"

	"github.com/unixpickle/model3d"
)

func main() {
	solid := &model3d.SphereSolid{Radius: 0.5}
	scanner := model3d.NewRectScanner(solid, 0.1)
	for i := 0; i < 3; i++ {
		scanner.Subdivide()
	}
	mesh := scanner.Mesh()
	ioutil.WriteFile("output.stl", mesh.EncodeSTL(), 0755)
}
