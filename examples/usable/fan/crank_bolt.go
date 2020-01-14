package main

import (
	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func CrankBoltMesh() *model3d.Mesh {
	solid := CrankBoltSolid()
	mesh := model3d.SolidToMesh(solid, 0.005, 0, -1, 5)
	return mesh
}

func CrankBoltSolid() model3d.Solid {
	return model3d.JoinedSolid{
		&model3d.CylinderSolid{
			P2:     model3d.Coord3D{Z: CrankBoltThickness},
			Radius: CrankBoltRadius,
		},
		&model3d.CylinderSolid{
			P1:     model3d.Coord3D{Z: CrankBoltThickness},
			P2:     model3d.Coord3D{Z: CrankBoltThickness + SpineThickness + PoleExtraLength},
			Radius: PoleRadius,
		},
		&toolbox3d.ScrewSolid{
			P1: model3d.Coord3D{Z: CrankBoltThickness + SpineThickness + PoleExtraLength},
			P2: model3d.Coord3D{Z: CrankBoltThickness + SpineThickness + PoleExtraLength +
				GearThickness},
			Radius:     ScrewRadius - ScrewSlack,
			GrooveSize: ScrewGrooveSize,
		},
	}
}
