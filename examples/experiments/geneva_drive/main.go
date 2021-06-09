package main

import (
	"flag"
	"image"
	"image/color"
	"image/gif"
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
	CreateAnimation(&spec)
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
	drivenMesh = drivenMesh.Rotate(math.Pi / 4).Translate(model2d.X(spec.CenterDistance))

	mesh := model2d.NewMesh()
	mesh.AddMesh(driveMesh)
	mesh.AddMesh(drivenMesh)
	mesh.Scale(200).SaveSVG("rendering_profiles.svg")
}

func CreateRendering(spec *Spec) {
	driveTransform := model3d.JoinedTransform{
		model3d.Rotation(model3d.Z(1), math.Pi/2),
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

	mesh := LoadModel("board")
	mesh.AddMesh(LoadModel("drive").MapCoords(driveTransform.Apply))
	mesh.AddMesh(LoadModel("driven").MapCoords(drivenTransform.Apply))

	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func CreateAnimation(spec *Spec) {
	boardMesh := LoadModel("board")
	driveMesh := LoadModel("drive")
	drivenMesh := LoadModel("driven")

	cameraOrigin := model3d.Coord3D{
		X: boardMesh.Max().Mid(boardMesh.Min()).X,
		Y: boardMesh.Min().Y * 4,
		Z: boardMesh.Max().Y * 4,
	}
	cameraTarget := model3d.XYZ(cameraOrigin.X, 0, 0)
	renderer := render3d.RayCaster{
		Camera: render3d.NewCameraAt(cameraOrigin, cameraTarget, math.Pi/3.6),
		Lights: []*render3d.PointLight{
			{
				Origin: cameraTarget.Add(cameraOrigin.Sub(cameraTarget).Scale(10)),
				Color:  render3d.NewColor(1.0),
			},
		},
	}

	drivenCollider := model3d.MeshToCollider(drivenMesh)
	checkDrivenAngle := func(driveTransform model3d.Transform, angle float64) model3d.Transform {
		pinCenter := model3d.Coord3D{
			X: spec.DriveRadius() + spec.Slack,
			// We use a sphere to simulate collisions with the pin, so we want
			// it as high as possible so that it doesn't accidentally collide
			// with the bottom of the gear.
			// Ideally, we'd use a profile here instead, but alas.
			Z: spec.Thickness + spec.BottomThickness,
		}
		pinLoc := driveTransform.Apply(pinCenter)
		drivenTransform := model3d.JoinedTransform{
			model3d.Rotation(model3d.Z(1), angle),
			&model3d.Translate{
				Offset: model3d.Coord3D{
					X: spec.DriveRadius() + spec.CenterDistance,
					Z: spec.BoardThickness,
				},
			},
		}
		localLoc := drivenTransform.Inverse().Apply(pinLoc)
		if !drivenCollider.SphereCollision(localLoc, spec.PinRadius-spec.Slack) {
			return drivenTransform
		}
		return nil
	}

	// This is what the angle ends at the end, since
	// there's a bit of slack.
	drivenAngle := -1.5337695312500002
	adjustDriven := func(driveTransform model3d.Transform) model3d.Transform {
		for offset := 0.0; offset < 0.3; offset += 0.01 {
			if dt := checkDrivenAngle(driveTransform, drivenAngle-offset); dt != nil {
				if offset == 0 {
					return dt
				}
				minAngle := offset - 0.01
				maxAngle := offset
				for i := 0; i < 10; i++ {
					mid := (minAngle + maxAngle) / 2
					if checkDrivenAngle(driveTransform, drivenAngle-mid) != nil {
						maxAngle = mid
					} else {
						minAngle = mid
					}
				}
				drivenAngle -= maxAngle
				return checkDrivenAngle(driveTransform, drivenAngle)
			}
		}
		panic("no rotation found to avoid collisions")
	}

	var g gif.GIF
	for driveAngle := math.Pi / 2; driveAngle < math.Pi/2+math.Pi*2; driveAngle += 0.05 {
		log.Println("Rendering drive angle", driveAngle, "...")
		driveTransform := model3d.JoinedTransform{
			model3d.Rotation(model3d.Z(1), driveAngle),
			&model3d.Translate{
				Offset: model3d.Coord3D{X: spec.DriveRadius(), Z: spec.BoardThickness},
			},
		}
		drivenTransform := adjustDriven(driveTransform)

		mesh := model3d.NewMesh()
		mesh.AddMesh(boardMesh)
		mesh.AddMesh(driveMesh.MapCoords(driveTransform.Apply))
		mesh.AddMesh(drivenMesh.MapCoords(drivenTransform.Apply))

		img := render3d.NewImage(200, 200)
		renderer.Render(img, render3d.Objectify(mesh, nil))
		grayImg := img.Gray()

		var palette []color.Color
		for i := 0; i < 256; i++ {
			palette = append(palette, color.Gray{Y: uint8(i)})
		}
		outImg := image.NewPaletted(image.Rect(0, 0, img.Width, img.Height), palette)
		for y := 0; y < img.Height; y++ {
			for x := 0; x < img.Width; x++ {
				outImg.Set(x, y, grayImg.At(x, y))
			}
		}
		g.Image = append(g.Image, outImg)
		g.Delay = append(g.Delay, 3)
	}

	w, err := os.Create("output.gif")
	essentials.Must(err)
	defer w.Close()
	essentials.Must(gif.EncodeAll(w, &g))
}

func LoadModel(name string) *model3d.Mesh {
	path := filepath.Join("models", name+".stl")
	r, err := os.Open(path)
	essentials.Must(err)
	defer r.Close()
	tris, err := model3d.ReadSTL(r)
	essentials.Must(err)
	return model3d.NewMeshTriangles(tris)
}
