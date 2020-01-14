package main

import (
	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func SmallGearMesh() *model3d.Mesh {
	solid := SmallGearSolid()
	mesh := model3d.SolidToMesh(solid, 0.005, 0, -1, 5)
	return mesh
}

func SmallGearSolid() model3d.Solid {
	poleTop := model3d.Coord3D{Z: GearThickness + SpineThickness + SpineWasherSize + PoleExtraLength}
	return model3d.JoinedSolid{
		&toolbox3d.HelicalGear{
			P2: model3d.Coord3D{Z: GearThickness},
			Profile: toolbox3d.InvoluteGearProfileSizes(GearPressureAngle, GearModule,
				GearAddendum, GearDedendum, SmallGearTeeth),
			Angle: -GearHelicalAngle,
		},
		&model3d.CylinderSolid{
			P1:     model3d.Coord3D{Z: GearThickness},
			P2:     poleTop,
			Radius: PoleRadius,
		},
		&toolbox3d.ScrewSolid{
			P1:         poleTop,
			P2:         poleTop.Add(model3d.Coord3D{Z: BladeDepth}),
			Radius:     ScrewRadius - ScrewSlack,
			GrooveSize: ScrewGrooveSize,
		},
	}
}
