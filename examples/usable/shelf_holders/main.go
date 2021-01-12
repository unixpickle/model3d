package main

import (
	"fmt"

	"github.com/unixpickle/model3d/model3d"
)

const (
	SmallerHeight = 0.4
	LargerHeight  = 0.6
	LargerRadius  = 0.3
)

func main() {
	for holeSize := 0.04; holeSize < 0.2; holeSize += 0.02 {
		solid := model3d.StackSolids(
			&model3d.Cylinder{
				P2:     model3d.Z(LargerHeight),
				Radius: LargerRadius,
			},
			&model3d.Cylinder{
				P2:     model3d.Z(SmallerHeight),
				Radius: holeSize / 2,
			},
		)
		mesh := model3d.MarchingCubesSearch(solid, 0.01, 8)
		mesh.SaveGroupedSTL(fmt.Sprintf("peg_%0.2f.stl", holeSize))
	}
}
