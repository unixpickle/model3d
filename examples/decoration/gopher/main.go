package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func main() {
	base, baseColor := Base()
	body, bodyColor := Body()
	eyes, eyesColor := Eyes()
	ears, earsColor := Ears()
	nose, noseColor := Nose()
	teeth, teethColor := Teeth()
	hands, handsColor := Hands()
	feet, feetColor := Feet()
	solid := model3d.JoinedSolid{base, body, eyes, ears, nose, teeth, hands, feet}
	mesh, points := model3d.MarchingCubesInterior(solid, 0.01, 8)
	cf := toolbox3d.JoinedSolidCoordColorFunc(
		points,
		base, baseColor,
		body, bodyColor,
		eyes, eyesColor,
		ears, earsColor,
		nose, noseColor,
		teeth, teethColor,
		hands, handsColor,
		feet, feetColor,
	)
	render3d.SaveRendering("rendering.png", mesh, model3d.XYZ(1, 5, 3), 512, 512, cf.RenderColor)
	mesh.SaveMaterialOBJ("gopher.zip", cf.TriangleColor)
}
