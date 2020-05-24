package main

import (
	"flag"
	"math"

	"github.com/unixpickle/model3d/model2d"
)

type Spec struct {
	DriveRadius    float64
	DrivenRadius   float64
	CenterDistance float64
	PinRadius      float64
	Slack          float64
}

func main() {
	var spec Spec

	flag.Float64Var(&spec.DriveRadius, "drive-radius", 1.5, "radius of driving disk")
	flag.Float64Var(&spec.DrivenRadius, "driven-radius", 2.0, "radius of driven disk")
	flag.Float64Var(&spec.CenterDistance, "center-distance", 2.5, "distance between disk centers")
	flag.Float64Var(&spec.PinRadius, "pin-radius", 0.1, "radius of driving pin")
	flag.Float64Var(&spec.Slack, "slack", 0.04, "extra space between parts")
	flag.Parse()

	driven := DrivenProfile(&spec)
	drive := DriveProfile(&spec, driven)

	RenderEngagedProfiles(&spec, driven, drive)
}

func RenderEngagedProfiles(spec *Spec, driven, drive model2d.Solid) {
	driveMesh := model2d.MarchingSquaresSearch(drive, 0.01, 8)
	drivenMesh := model2d.MarchingSquaresSearch(driven, 0.01, 8)

	mat := model2d.NewMatrix2Rotation(math.Pi / 4)
	drivenMesh = drivenMesh.MapCoords(func(c model2d.Coord) model2d.Coord {
		return mat.MulColumn(c).Add(model2d.Coord{X: spec.CenterDistance})
	})

	mesh := model2d.NewMesh()
	mesh.AddMesh(driveMesh)
	mesh.AddMesh(drivenMesh)
	mesh.Scale(200).SaveSVG("rendering_profiles.svg")
}
