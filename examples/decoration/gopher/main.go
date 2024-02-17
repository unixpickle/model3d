package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func main() {
	body, bodyColor := Body()
	eyes, eyesColor := Eyes()
	ears, earsColor := Ears()
	nose, noseColor := Nose()
	teeth, teethColor := Teeth()
	solid := model3d.JoinedSolid{body, eyes, ears, nose, teeth}
	mesh, points := model3d.MarchingCubesInterior(solid, 0.02, 8)
	cf := toolbox3d.JoinedSolidCoordColorFunc(
		points,
		body, bodyColor,
		eyes, eyesColor,
		ears, earsColor,
		nose, noseColor,
		teeth, teethColor,
	)
	render3d.SaveRotatingGIF("rendering.gif", mesh, model3d.Z(1), model3d.Y(-1), 300, 20, 5.0, cf.RenderColor)
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, cf.RenderColor)
}
