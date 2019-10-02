package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
)

const (
	AxleRadius  = 0.15
	Thickness   = 0.15
	PartSpacing = 0.01
	EtchInset   = 0.05
)

func main() {
	images := [3]image.Image{}
	for i := range images {
		f, err := os.Open(fmt.Sprintf("side_%d.png", i+1))
		essentials.Must(err)
		images[i], _, err = image.Decode(f)
		f.Close()
		essentials.Must(err)
	}

	axle := &model3d.SubtractedSolid{
		Positive: model3d.JoinedSolid{
			&model3d.CylinderSolid{
				P1:     model3d.Coord3D{Z: -Thickness},
				P2:     model3d.Coord3D{Z: Thickness},
				Radius: AxleRadius,
			},
			&model3d.CylinderSolid{
				P1:     model3d.Coord3D{Z: -Thickness},
				P2:     model3d.Coord3D{Z: -Thickness * 1.5},
				Radius: 0.3,
			},
			&model3d.CylinderSolid{
				P1:     model3d.Coord3D{Z: Thickness},
				P2:     model3d.Coord3D{Z: Thickness * 1.5},
				Radius: 0.3,
			},
		},
		Negative: model3d.JoinedSolid{
			&model3d.SphereSolid{
				Center: model3d.Coord3D{Z: -Thickness*1.5 - 1.0},
				Radius: 1.05,
			},
			&model3d.SphereSolid{
				Center: model3d.Coord3D{Z: Thickness*1.5 + 1.0},
				Radius: 1.05,
			},
		},
	}

	mesh := model3d.SolidToMesh(axle, 0.05, 4, 1.0, 3)

	body := &model3d.SubtractedSolid{
		Positive: model3d.JoinedSolid{
			&model3d.CylinderSolid{
				P1:     model3d.Coord3D{Z: -Thickness + PartSpacing},
				P2:     model3d.Coord3D{Z: Thickness - PartSpacing},
				Radius: 0.4,
			},
			&model3d.CylinderSolid{
				P1:     model3d.Coord3D{Y: -0.6, Z: -Thickness + PartSpacing},
				P2:     model3d.Coord3D{Y: -0.6, Z: Thickness - PartSpacing},
				Radius: 0.4,
			},
			&model3d.CylinderSolid{
				P1: model3d.Coord3D{
					X: -0.6 * math.Sin(math.Pi*2/3),
					Y: -0.6 * math.Cos(math.Pi*2/3),
					Z: -Thickness + PartSpacing,
				},
				P2: model3d.Coord3D{
					X: -0.6 * math.Sin(math.Pi*2/3),
					Y: -0.6 * math.Cos(math.Pi*2/3),
					Z: Thickness - PartSpacing,
				},
				Radius: 0.4,
			},
			&model3d.CylinderSolid{
				P1: model3d.Coord3D{
					X: -0.6 * math.Sin(math.Pi*4/3),
					Y: -0.6 * math.Cos(math.Pi*4/3),
					Z: -Thickness + PartSpacing,
				},
				P2: model3d.Coord3D{
					X: -0.6 * math.Sin(math.Pi*4/3),
					Y: -0.6 * math.Cos(math.Pi*4/3),
					Z: Thickness - PartSpacing,
				},
				Radius: 0.4,
			},
		},
		Negative: model3d.JoinedSolid{
			&model3d.CylinderSolid{
				P1:     model3d.Coord3D{Z: -Thickness + PartSpacing},
				P2:     model3d.Coord3D{Z: Thickness - PartSpacing},
				Radius: AxleRadius + PartSpacing,
			},
			&EtchedImage{
				Image:  images[0],
				X:      0,
				Y:      -0.6,
				Radius: 0.4,
			},
			&EtchedImage{
				Image:  images[1],
				X:      -0.6 * math.Sin(math.Pi*2/3),
				Y:      -0.6 * math.Cos(math.Pi*2/3),
				Radius: 0.4,
			},
			&EtchedImage{
				Image:  images[2],
				X:      -0.6 * math.Sin(math.Pi*4/3),
				Y:      -0.6 * math.Cos(math.Pi*4/3),
				Radius: 0.4,
			},
		},
	}
	mesh.AddMesh(model3d.SolidToMesh(body, 0.05, 4, 1.0, 5))

	ioutil.WriteFile("spinner.stl", mesh.EncodeSTL(), 0755)

	RenderMesh(mesh)
}

func RenderMesh(m *model3d.Mesh) {
	log.Println("Saving rendering of mesh...")
	image := image.NewGray(image.Rect(0, 0, 400, 300))
	model3d.RenderRayCast(model3d.MeshToCollider(m), image,
		model3d.Coord3D{X: 0, Y: 1.5, Z: 1.5},
		model3d.Coord3D{X: 1},
		model3d.Coord3D{Y: 1, Z: -1},
		model3d.Coord3D{Y: -1, Z: -1},
		math.Pi/3,
	)
	w, err := os.Create("rendering.png")
	essentials.Must(err)
	defer w.Close()
	essentials.Must(png.Encode(w, image))
}

type EtchedImage struct {
	Image  image.Image
	X      float64
	Y      float64
	Radius float64
}

func (e *EtchedImage) Min() model3d.Coord3D {
	return model3d.Coord3D{X: e.X - e.Radius, Y: e.Y - e.Radius, Z: -Thickness}
}

func (e *EtchedImage) Max() model3d.Coord3D {
	return model3d.Coord3D{X: e.X + e.Radius, Y: e.Y + e.Radius, Z: Thickness}
}

func (e *EtchedImage) Contains(c model3d.Coord3D) bool {
	if c.X < e.X-e.Radius || c.Y < e.Y-e.Radius || c.X > e.X+e.Radius || c.Y > e.Y+e.Radius {
		return false
	}
	if math.Abs(c.Z) > Thickness || math.Abs(c.Z) < Thickness-EtchInset {
		return false
	}

	axis1 := model3d.Coord2D{X: e.X, Y: e.Y}
	axis1 = axis1.Scale(1 / (axis1.Norm() * e.Radius))
	axis2 := model3d.Coord2D{X: -axis1.Y, Y: axis1.X}
	if c.Z < 0 {
		axis2 = axis2.Scale(-1)
	}
	p := model3d.Coord2D{X: c.X - e.X, Y: c.Y - e.Y}

	x := int(math.Round(float64(e.Image.Bounds().Dx()) * (axis2.Dot(p) + 1) / 2))
	y := int(math.Round(float64(e.Image.Bounds().Dy()) * (axis1.Dot(p) + 1) / 2))

	if x < 0 {
		return false
	} else if x >= e.Image.Bounds().Dx() {
		return false
	}

	if y < 0 {
		return false
	} else if y >= e.Image.Bounds().Dy() {
		return false
	}

	r, _, _, _ := e.Image.At(x, y).RGBA()
	if r > 0xffff/2 {
		return false
	}
	return math.Abs(c.Z) > Thickness-EtchInset
}
