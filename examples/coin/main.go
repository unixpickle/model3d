package main

import (
	"flag"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"math"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
)

const (
	NumSlices = 100
	NumStops  = 200
)

func main() {
	var outFile string
	var minHeight float64
	var maxHeight float64
	var radius float64
	var template string

	flag.StringVar(&outFile, "out", "coin.stl", "output file name")
	flag.Float64Var(&minHeight, "min-height", 0.1, "minimum height")
	flag.Float64Var(&maxHeight, "max-height", 0.13, "maximum height")
	flag.Float64Var(&radius, "radius", 0.5, "radius of coin")
	flag.StringVar(&template, "template", "example.png", "coin depth map image")

	flag.Parse()

	f, err := os.Open(template)
	essentials.Must(err)
	templateImg, _, err := image.Decode(f)
	f.Close()
	essentials.Must(err)

	hFunc := &HeightFunc{
		Img:       templateImg,
		MinHeight: minHeight,
		MaxHeight: maxHeight,
	}

	m := CreateRoundMesh(hFunc, radius)
	for i := 0; i < 11; i++ {
		Subdivide(m, hFunc, radius)
	}
	FillVolume(m)

	essentials.Must(ioutil.WriteFile(outFile, m.EncodeSTL(), 0755))
}

func CreateRoundMesh(h *HeightFunc, radius float64) *model3d.Mesh {
	m := model3d.NewMesh()
	midHeight := h.Height(0, 0)
	for i := 0; i < NumSlices; i++ {
		theta := 2 * math.Pi * float64(i) / NumSlices
		nextTheta := 2 * math.Pi * float64((i+1)%NumSlices) / NumSlices
		m.Add(&model3d.Triangle{
			model3d.Coord3D{X: 0, Y: 0, Z: midHeight},
			h.Coord(nextTheta, 1, radius),
			h.Coord(theta, 1, radius),
		})
	}
	return m
}

func Subdivide(m *model3d.Mesh, h *HeightFunc, radius float64) {
	subdivider := model3d.NewSubdivider()
	subdivider.AddFiltered(m, func(p1, p2 model3d.Coord3D) bool {
		return !h.IsFlat(p1, p2, radius)
	})
	subdivider.Subdivide(m, func(p1, p2 model3d.Coord3D) model3d.Coord3D {
		x := (p1.X + p2.X) / 2
		y := (p1.Y + p2.Y) / 2
		theta := math.Atan2(y, x)
		r := math.Sqrt(x*x+y*y) / radius
		return h.Coord(theta, r, radius)
	})
}

func FillVolume(m *model3d.Mesh) {
	m.Iterate(func(t *model3d.Triangle) {
		t1 := *t
		for i := range t1 {
			t1[i].Z = 0
		}

		// Create sides for edge triangles.
		for i := 0; i < 3; i++ {
			a := i
			b := (i + 1) % 3
			if len(m.Find(t[a], t[b])) == 1 {
				m.Add(&model3d.Triangle{t[b], t[a], t1[a]})
				m.Add(&model3d.Triangle{t[b], t1[a], t1[b]})
			}
		}

		// Flip normal for bottom face.
		t1[1], t1[2] = t1[2], t1[1]

		m.Add(&t1)
	})
}

type HeightFunc struct {
	Img       image.Image
	MinHeight float64
	MaxHeight float64
}

func (h *HeightFunc) Height(theta, radius float64) float64 {
	x := math.Round((math.Cos(theta)*radius + 1) * float64(h.Img.Bounds().Dx()) / 2)
	y := math.Round((1 - math.Sin(theta)*radius) * float64(h.Img.Bounds().Dy()) / 2)
	c := h.Img.At(int(x), int(y))
	r, g, b, _ := color.RGBAModel.Convert(c).RGBA()
	relHeight := 1 - float64(r+g+b)/(3*0xffff)
	return h.MinHeight + (h.MaxHeight-h.MinHeight)*relHeight
}

func (h *HeightFunc) Coord(theta, radius, radiusScale float64) model3d.Coord3D {
	return model3d.Coord3D{
		X: radiusScale * radius * math.Cos(theta),
		Y: radiusScale * radius * math.Sin(theta),
		Z: h.Height(theta, radius),
	}
}

func (h *HeightFunc) IsFlat(p1, p2 model3d.Coord3D, radius float64) bool {
	if p1.Z != p2.Z {
		return false
	}

	totalDist := p1.Dist(p2)

	for t := 0.0; t < totalDist; t += radius / NumStops {
		frac := t / totalDist
		x := p1.X*(1-frac) + p2.X*frac
		y := p1.Y*(1-frac) + p2.Y*frac
		theta := math.Atan2(y, x)
		r := math.Sqrt(x*x+y*y) / radius
		if h.Height(theta, r) != p1.Z {
			return false
		}
	}

	return true
}
