package shorthand

import (
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

type C2 = model2d.Coord
type C3 = model3d.Coord3D

var Origin2 C2 = model2d.Origin
var Origin3 C3 = model3d.Origin

func XYZ(x, y, z float64) C3 {
	return model3d.XYZ(x, y, z)
}

func XY(x, y float64) C2 {
	return model2d.XY(x, y)
}

func Z(z float64) C3 {
	return model3d.Z(z)
}
