package main

import (
	"flag"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/unixpickle/model3d/render3d"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model3d"

	"github.com/unixpickle/model3d/model2d"
)

type Spec struct {
	DrivenRadius    float64
	CenterDistance  float64
	PinRadius       float64
	Slack           float64
	SharpEdgeCutoff float64

	DrivenSupportRadius float64

	BottomThickness float64
	Thickness       float64
	BoardThickness  float64
	ScrewRadius     float64
	ScrewSlack      float64
	ScrewGroove     float64
	ScrewCapHeight  float64
	ScrewCapRadius  float64
}

func (s *Spec) DriveRadius() float64 {
	return math.Sqrt2 * s.CenterDistance / 2
}

func main() {
	var spec Spec

	flag.Float64Var(&spec.DrivenRadius, "driven-radius", 1.0, "radius of driven disk")
	flag.Float64Var(&spec.CenterDistance, "center-distance", 1.8, "distance between disk centers")
	flag.Float64Var(&spec.PinRadius, "pin-radius", 0.1, "radius of driving pin")
	flag.Float64Var(&spec.Slack, "slack", 0.015, "extra space between parts")
	flag.Float64Var(&spec.SharpEdgeCutoff, "sharp-edge-cutoff", 0.02,
		"driven gear cutoff to prevent sharp features")
	flag.Float64Var(&spec.DrivenSupportRadius, "driven-support-radius", 0.3,
		"radius of cylinder supporting the driven gear")

	flag.Float64Var(&spec.BottomThickness, "bottom-thickness", 0.2,
		"thickness of bottom half of gears")
	flag.Float64Var(&spec.Thickness, "thickness", 0.3, "thickness of engaged part of gears")
	flag.Float64Var(&spec.BoardThickness, "board-thickness", 0.4, "thickness of board")
	flag.Float64Var(&spec.ScrewRadius, "screw-radius", 0.2, "radius of screws")
	flag.Float64Var(&spec.ScrewSlack, "screw-slack", 0.02, "slack for screws in holes")
	flag.Float64Var(&spec.ScrewGroove, "screw-groove", 0.05, "groove size of screws")
	flag.Float64Var(&spec.ScrewCapHeight, "screw-cap-height", 0.3, "height of screw heads")
	flag.Float64Var(&spec.ScrewCapRadius, "screw-cap-radius", 0.3, "radius of screw heads")

	flag.Parse()

	log.Println("Creating profiles ...")
	driven := DrivenProfile(&spec)
	drive := DriveProfile(&spec, driven)
	RenderEngagedProfiles(&spec, driven, drive)

	if _, err := os.Stat("models"); os.IsNotExist(err) {
		essentials.Must(os.Mkdir("models", 0755))
	}

	CreateModel("drive", DriveBody(&spec, drive))
	CreateModel("driven", DrivenBody(&spec, driven))
	CreateModel("board", BoardSolid(&spec))
	CreateModel("screw", BoardScrewSolid(&spec))

	CreateRendering(&spec)
}

func CreateModel(name string, solid model3d.Solid) {
	outPath := filepath.Join("models", name+".stl")
	if _, err := os.Stat(outPath); err == nil {
		log.Println("Skipping existing model:", name)
		return
	}
	log.Println("Creating model:", name, "...")
	mesh := model3d.MarchingCubesSearch(solid, 0.01, 8)
	mesh = mesh.EliminateCoplanar(1e-5)
	essentials.Must(mesh.SaveGroupedSTL(outPath))
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

func CreateRendering(spec *Spec) {
	loadModel := func(name string) *model3d.Mesh {
		path := filepath.Join("models", name+".stl")
		r, err := os.Open(path)
		essentials.Must(err)
		defer r.Close()
		tris, err := model3d.ReadSTL(r)
		essentials.Must(err)
		return model3d.NewMeshTriangles(tris)
	}

	driveTransform := model3d.JoinedTransform{
		&model3d.Matrix3Transform{
			Matrix: model3d.NewMatrix3Rotation(model3d.Coord3D{Z: 1}, math.Pi/2),
		},
		&model3d.Translate{
			Offset: model3d.Coord3D{X: spec.DriveRadius(), Z: spec.BoardThickness},
		},
	}

	drivenTransform := &model3d.Translate{
		Offset: model3d.Coord3D{
			X: spec.DriveRadius() + spec.CenterDistance,
			Z: spec.BoardThickness,
		},
	}

	mesh := loadModel("board")
	mesh.AddMesh(loadModel("drive").MapCoords(driveTransform.Apply))
	mesh.AddMesh(loadModel("driven").MapCoords(drivenTransform.Apply))

	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}
