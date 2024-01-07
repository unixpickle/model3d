package main

import (
	"log"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

var Yellow = render3d.NewColorRGB(224.0/255, 209.0/255, 0)

func main() {
	body := InetBody(true)

	lid, lidCutout := LidAndCutout()

	jack := EthernetJackSolid()
	min, max := jack.Min(), jack.Max()
	min = min.AddScalar(0.01)
	max = max.AddScalar(-0.01)
	min.Y -= 0.1
	jackHole := model3d.NewRect(min, max)
	min.Y = max.Y - 0.1
	jackEnd := model3d.NewRect(min, max)

	usbPort := model3d.NewRect(
		model3d.XYZ(BodySideLength/2-0.3, -0.8, 0.2),
		model3d.XYZ(BodySideLength/2+0.1, -0.65, 0.7),
	)
	usbInner := model3d.NewRect(
		model3d.XYZ(BodySideLength/2-0.3, -0.8, 0.2),
		model3d.XYZ(BodySideLength/2-0.05, -0.75, 0.7),
	)

	fanHole := FanHole()

	powerSwitchHole := model3d.NewColliderSolidInset(
		model3d.NewRect(
			model3d.XYZ(BodySideLength/2-0.05, -0.1, 0.3),
			model3d.XYZ(BodySideLength/2+0.01, 0.1, 0.3001),
		),
		-0.075,
	)
	powerSwitch := model3d.JoinedSolid{
		&model3d.Cylinder{
			P1:     model3d.XYZ(BodySideLength/2-0.2, 0.0, 0.3),
			P2:     model3d.XYZ(BodySideLength/2, 0.0, 0.3),
			Radius: 0.08,
		},
		// Mask the rounded back of the powerSwitchHole.
		model3d.NewRect(
			model3d.XYZ(BodySideLength/2-0.2, -0.2, 0.2),
			model3d.XYZ(BodySideLength/2-0.05, 0.2, 0.4),
		),
	}

	body = model3d.JoinedSolid{
		&model3d.SubtractedSolid{
			Positive: body,
			Negative: model3d.JoinedSolid{
				jackHole,
				usbPort,
				fanHole,
				lidCutout,
				powerSwitchHole,
			},
		},
		usbInner,
		powerSwitch,
	}

	joined := model3d.JoinedSolid{
		body,
		jack,
		jackEnd,
	}

	log.Println("Creating meshes...")
	lidMesh := model3d.DualContour(lid, 0.01, true, false)
	lidMesh = lidMesh.EliminateCoplanar(1e-5)
	mesh, interior := model3d.DualContourInterior(joined, 0.01, true, false)
	colorFunc := toolbox3d.JoinedSolidCoordColorFunc(
		interior,
		body, Yellow,
		jackEnd, render3d.NewColor(0.5),
		usbInner, render3d.NewColor(0.5),
		jack, render3d.NewColor(0.9),
	)
	mesh = mesh.EliminateCoplanarFiltered(1e-5, colorFunc.ChangeFilterFunc(mesh, 0.05))
	mesh.SaveMaterialOBJ("body.zip", colorFunc.TriangleColor)
	lidMesh.SaveMaterialOBJ("lid.zip", toolbox3d.ConstantCoordColorFunc(Yellow).TriangleColor)
	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, colorFunc.RenderColor)
}
