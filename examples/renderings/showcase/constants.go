package main

import "github.com/unixpickle/model3d"

var LightDirection = model3d.Coord3D{X: 2, Y: -3, Z: 3}.Normalize()

const (
	CameraY = -5
	CameraZ = 4

	WineGlassX = -6.5
	WineGlassY = 9.0

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
)
