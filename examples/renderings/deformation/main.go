package main

import (
	"image"
	"image/color"
	"image/gif"
	"log"
	"math"
	"os"

	"github.com/unixpickle/model3d/model2d"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	FrameSkip = 4
	ImageSize = 200
)

func main() {
	mesh := CreateMesh()

	rotation := model2d.BezierCurve{
		model2d.Coord{X: 0, Y: 0},
		model2d.Coord{X: 0.5, Y: math.Pi / 4},
		model2d.Coord{X: 1.0, Y: math.Pi / 4},
		model2d.Coord{X: 2.0, Y: math.Pi / 4},
		model2d.Coord{X: 2.5, Y: 0.1},
		model2d.Coord{X: 3.0, Y: 0},
	}
	translateX := model2d.BezierCurve{
		model2d.Coord{X: 0, Y: 0},
		model2d.Coord{X: 0.5, Y: 0},
		model2d.Coord{X: 1.0, Y: -0.6},
		model2d.Coord{X: 2.0, Y: -0.5},
		model2d.Coord{X: 2.5, Y: 0.6},
		model2d.Coord{X: 3.0, Y: 0.5},
	}
	translateZ := model2d.BezierCurve{
		model2d.Coord{X: 0, Y: 0},
		model2d.Coord{X: 0.5, Y: -0.3},
		model2d.Coord{X: 1.0, Y: 0},
		model2d.Coord{X: 1.5, Y: 0.3},
		model2d.Coord{X: 2.0, Y: 0},
		model2d.Coord{X: 3.0, Y: 0},
	}

	a := model3d.NewARAP(mesh)
	df := a.SeqDeformer()

	var g gif.GIF
	var frame int
	for t := 0.0; t < 3.0; t += 0.05 {
		log.Println("Frame", frame, "...")
		rotation := rotation.EvalX(t)
		translate := model3d.Coord3D{X: translateX.EvalX(t), Z: translateZ.EvalX(t)}
		transform := model3d.JoinedTransform{
			&model3d.Matrix3Transform{
				Matrix: model3d.NewMatrix3Rotation(model3d.Z(1), rotation),
			},
			&model3d.Translate{Offset: translate},
		}
		deformed := df(Constraints(mesh, transform))
		if frame%FrameSkip == 0 {
			g.Image = append(g.Image, RenderFrame(deformed))
			g.Delay = append(g.Delay, 10*FrameSkip)
		}
		frame++
	}

	w, err := os.Create("output.gif")
	essentials.Must(err)
	defer w.Close()
	essentials.Must(gif.EncodeAll(w, &g))
}

func RenderFrame(mesh *model3d.Mesh) *image.Paletted {
	renderer := &render3d.RayCaster{
		Camera: render3d.NewCameraAt(model3d.Coord3D{Y: -3}, model3d.Coord3D{}, math.Pi/3.6),
		Lights: []*render3d.PointLight{
			{
				Origin: model3d.Coord3D{Y: -100},
				Color:  render3d.NewColor(1.0),
			},
		},
	}
	img := render3d.NewImage(ImageSize, ImageSize)
	renderer.Render(img, render3d.Objectify(mesh, nil))

	var palette []color.Color
	for i := 0; i < 256; i++ {
		palette = append(palette, color.Gray{Y: uint8(i)})
	}
	fullImg := img.Gray()
	outImg := image.NewPaletted(image.Rect(0, 0, img.Width, img.Height), palette)
	for y := 0; y < img.Height; y++ {
		for x := 0; x < img.Width; x++ {
			outImg.Set(x, y, fullImg.At(x, y))
		}
	}
	return outImg
}

func CreateMesh() *model3d.Mesh {
	box := model3d.NewMesh()

	squareSize := 0.05
	width := 0.4
	height := 1.0

	addQuad := func(p model3d.Coord3D, normalAxis int) {
		ax1 := model3d.X(squareSize)
		ax2 := model3d.Y(squareSize)
		if normalAxis == 0 {
			ax1 = model3d.Z(squareSize)
		} else if normalAxis == 1 {
			ax2 = model3d.Z(squareSize)
		}
		box.Add(&model3d.Triangle{p, p.Add(ax1), p.Add(ax2)})
		box.Add(&model3d.Triangle{p.Add(ax1), p.Add(ax2), p.Add(ax1).Add(ax2)})
	}

	// All but top two faces.
	for x := -width; x < width-1e-8; x += squareSize {
		for z := -height; z < height-1e-8; z += squareSize {
			for _, y := range []float64{-width, width} {
				addQuad(model3d.XYZ(x, y, z), 1)
				addQuad(model3d.XYZ(y, x, z), 0)
			}
		}
	}

	// Top two faces.
	for x := -width; x < width-1e-8; x += squareSize {
		for y := -width; y < width-1e-8; y += squareSize {
			addQuad(model3d.XYZ(x, y, -height), 2)
			addQuad(model3d.XYZ(x, y, height), 2)
		}
	}

	// Fix holes due to rounding errors.
	box = box.Repair(1e-8)

	// Fix normals due to arbitrary triangle ordering.
	box, _ = box.RepairNormals(1e-8)

	// Don't let the box face the camera head on,
	// allowing us to see more detail.
	rotate := model3d.NewMatrix3Rotation(model3d.Z(1), 0.4)
	return box.MapCoords(rotate.MulColumn)
}

func Constraints(mesh *model3d.Mesh, transform model3d.Transform) model3d.ARAPConstraints {
	min, max := mesh.Min(), mesh.Max()
	control := model3d.ARAPConstraints{}
	for _, v := range mesh.VertexSlice() {
		if v.Z == min.Z {
			control[v] = v
		} else if v.Z == max.Z {
			control[v] = transform.Apply(v)
		}
	}
	return control
}
