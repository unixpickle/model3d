package main

import (
	"flag"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
)

func main() {
	var outFile string
	var patternFile string
	var color1 string
	var color2 string
	var radius float64
	var length float64
	flag.StringVar(&outFile, "out", "pill.zip", "output file name")
	flag.StringVar(&patternFile, "pattern", "heart.png", "image to put on ends")
	flag.StringVar(&color1, "color1", "1.0,1.0,1.0", "color for half of pill")
	flag.StringVar(&color2, "coolr2", "1.0,0.0,0.0", "color for other half of pill")
	flag.Float64Var(&radius, "radius", 0.2, "radius of pill")
	flag.Float64Var(&length, "length", 1.0, "length of pill")
	flag.Parse()

	parsedColor1 := ParseColor(color1)
	parsedColor2 := ParseColor(color2)

	r, err := os.Open(patternFile)
	essentials.Must(err)
	img, _, err := image.Decode(r)
	r.Close()
	essentials.Must(err)
	imprinter := &Imprinter{Img: img, Radius: radius / 2}

	mesh := model3d.NewMeshPolar(func(g model3d.GeoCoord) float64 {
		return radius
	}, 150)

	for i := 0; i < 5; i++ {
		subdiv := model3d.NewSubdivider()
		subdiv.AddFiltered(mesh, func(p1, p2 model3d.Coord3D) bool {
			return imprinter.AlphaAt(p1.Y, p1.Z) != imprinter.AlphaAt(p2.Y, p2.Z)
		})
		subdiv.Subdivide(mesh, func(p1, p2 model3d.Coord3D) model3d.Coord3D {
			return p1.Mid(p2).Geo().Coord3D().Scale(radius)
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
			t2[0], t2[1] = t2[1], t2[0]
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
				p3 := p1
				p4 := p2
				p1.X = 0
				p2.X = 0
				p3.X *= -1
				p4.X *= -1
				mesh.Add(&model3d.Triangle{t[i], p1, t[i1]})
				mesh.Add(&model3d.Triangle{p1, p2, t[i1]})
				mesh.Add(&model3d.Triangle{p1, p3, p2})
				mesh.Add(&model3d.Triangle{p3, p4, p2})
			}
		}
	})

	colorFunc := func(t *model3d.Triangle) [3]float64 {
		var alpha bool
		for _, p := range t {
			if imprinter.AlphaAt(p.Y, p.Z) {
				alpha = true
				break
			}
		}
		if t[0].X < 0 || t[1].X < 0 || t[2].X < 0 {
			if alpha {
				return parsedColor2
			} else {
				return parsedColor1
			}
		} else {
			if alpha {
				return parsedColor1
			} else {
				return parsedColor2
			}
		}
	}

	ioutil.WriteFile(outFile, mesh.EncodeMaterialOBJ(colorFunc), 0755)
}

func ParseColor(color string) [3]float64 {
	parts := strings.Split(color, ",")
	if len(parts) != 3 {
		essentials.Die("invalid color string: " + color)
	}
	var res [3]float64
	for i, p := range parts {
		x, err := strconv.ParseFloat(p, 64)
		if err != nil {
			essentials.Die("invalid color string: " + color)
		}
		res[i] = x
	}
	return res
}

type Imprinter struct {
	Img    image.Image
	Radius float64
}

func (i *Imprinter) AlphaAt(y, z float64) bool {
	if y <= -i.Radius || y >= i.Radius || z <= -i.Radius || z >= i.Radius {
		return false
	}
	imgX := int(math.Round(float64(i.Img.Bounds().Dx()) * (y + i.Radius) / (i.Radius * 2)))
	imgY := int(math.Round(float64(i.Img.Bounds().Dy()) * (z + i.Radius) / (i.Radius * 2)))
	_, _, _, a := i.Img.At(imgX, imgY).RGBA()
	if a < 0xffff/2 {
		return false
	} else {
		return true
	}
}
