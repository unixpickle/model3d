package main

import (
	"flag"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"math"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
)

func main() {
	var patternFile string
	var color1 string
	var color2 string
	var radius float64
	var length float64
	flag.StringVar(&patternFile, "pattern", "heart.png", "image to put on ends")
	flag.StringVar(&color1, "color1", "1.0,1.0,1.0", "color for half of pill")
	flag.StringVar(&color2, "coolr2", "1.0,0.0,0.0", "color for other half of pill")
	flag.Float64Var(&radius, "radius", 0.2, "radius of pill")
	flag.Float64Var(&length, "length", 1.0, "length of pill")
	flag.Parse()

	r, err := os.Open(patternFile)
	essentials.Must(err)
	img, _, err := image.Decode(r)
	r.Close()
	essentials.Must(err)
	imprinter := &Imprinter{Img: img, Radius: radius}

	mesh := model3d.NewMeshPolar(func(g model3d.GeoCoord) float64 {
		return radius
	}, 150)

	for i := 0; i < 3; i++ {
		subdiv := model3d.NewSubdivider()
		subdiv.AddFiltered(mesh, func(p1, p2 model3d.Coord3D) bool {
			return imprinter.AlphaAt(p1.Y, p1.Z) != imprinter.AlphaAt(p2.Y, p2.Z)
		})
		subdiv.Subdivide(mesh, func(p1, p2 model3d.Coord3D) model3d.Coord3D {
			midpoint := model3d.Coord3D{
				X: (p1.X + p2.X) / 2,
				Y: (p1.Y + p2.Y) / 2,
				Z: (p1.Z + p2.Z) / 2,
			}
			return midpoint.Geo().Coord3D().Scale(radius)
		})
	}

	mesh.Iterate(func(t *model3d.Triangle) {
		mesh.Remove(t)
		if math.Min(math.Min(t[0].X, t[1].X), t[2].X) < -1e-4 {
			t1 := *t
			t2 := *t
			for i := range t1 {
				t1[i].X -= length/2 - radius
				t2[i].X = -t1[i].X
			}
			mesh.Add(&t1)
			mesh.Add(&t2)
		}
	})

	mesh.Iterate(func(t *model3d.Triangle) {
		if t[0].X > 0 {
			return
		}
		for i := 0; i < 3; i++ {
			i1 := (i + 1) % 3
			if len(mesh.Find(t[i], t[i1])) == 1 {
				p1 := t[i]
				p2 := t[i1]
				p1.X *= -1
				p2.X *= -1
				// TODO: create four triangles here so the color
				// can change right in the middle.
				mesh.Add(&model3d.Triangle{t[i], p1, t[i1]})
				mesh.Add(&model3d.Triangle{p1, p2, t[i1]})
			}
		}
	})

	colorFunc := func(t *model3d.Triangle) [3]float64 {
		// TODO: look at pattern here.
		return [3]float64{1, 1, 1}
	}

	ioutil.WriteFile("mesh.zip", mesh.EncodeMaterialOBJ(colorFunc), 0755)
}

type Imprinter struct {
	Img    image.Image
	Radius float64
}

func (i *Imprinter) AlphaAt(y, z float64) bool {
	imgX := int(math.Round(float64(i.Img.Bounds().Dx()) * (y + i.Radius) / (i.Radius * 2)))
	imgY := int(math.Round(float64(i.Img.Bounds().Dy()) * (z + i.Radius) / (i.Radius * 2)))
	_, _, _, a := i.Img.At(imgX, imgY).RGBA()
	if a < 0xffff/2 {
		return false
	} else {
		return true
	}
}
