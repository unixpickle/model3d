package main

import "github.com/unixpickle/model3d/model3d"

var (
	LightDirection = model3d.XYZ(2, -3, 3).Normalize()
	LightCenter    = LightDirection.Normalize().Scale(50)
)

const (
	CameraY = -5
	CameraZ = 4

	WineGlassX = -5.0
	WineGlassY = 10.5

	PumpkinX = -2
	PumpkinY = 10

	RocksY = 15.0

	VaseX          = 5
	VaseY          = 11
	RoseX          = VaseX
	RoseY          = VaseY
	RoseZ          = 7
	RoseStemRadius = 0.15

	CurvyThingX = 1.8
	CurvyThingY = 9.0

	RoomRadius = 100.0

	LightRadius     = 5.0
	LightBrightness = 300.0
)
