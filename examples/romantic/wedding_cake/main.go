package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	solid, color := MarbleHexagon()
	mesh := model3d.MarchingCubesSearch(solid, 0.02, 8)
	render3d.SaveRotatingGIF("hexagon.gif", mesh, model3d.Z(1), model3d.X(1), 300, 50, 10.0, color.RenderColor)
}
