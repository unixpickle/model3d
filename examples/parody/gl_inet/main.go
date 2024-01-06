package main

import (
	"log"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func main() {
	body := InetBody()

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

	body = model3d.JoinedSolid{
		&model3d.SubtractedSolid{
			Positive: body,
			Negative: model3d.JoinedSolid{
				jackHole,
				usbPort,
				fanHole,
			},
		},
		usbInner,
	}

	joined := model3d.JoinedSolid{
		body,
		jack,
		jackEnd,
	}

	log.Println("Creating mesh...")
	mesh, interior := model3d.DualContourInterior(joined, 0.01, true, false)
	mesh.SaveGroupedSTL("out.stl")
	colorFunc := toolbox3d.JoinedSolidCoordColorFunc(
		interior,
		body, render3d.NewColorRGB(224.0/255, 209.0/255, 0),
		jackEnd, render3d.NewColor(0.5),
		usbInner, render3d.NewColor(0.5),
		jack, render3d.NewColor(0.9),
	)
	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, colorFunc.RenderColor)
}
