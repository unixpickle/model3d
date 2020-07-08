package main

import (
	"io/ioutil"
	"log"
	"sync"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

var Green = [3]float64{
	float64(0x1b) / 255.0,
	float64(0xad) / 255.0,
	float64(0x64) / 255.0,
}

const (
	PickleLength = 2.0
	PickleWidth  = PickleLength / 2
)

func main() {
	var solid model3d.Solid
	solid = &PickleSolid{F: NewPickleFunction()}
	inscription := NewInscription()

	for _, color := range []bool{true, false} {
		if color {
			log.Print("Creating colored model...")
		} else {
			log.Print("Creating uncolored model...")
		}

		if !color {
			solid = &model3d.SubtractedSolid{
				Positive: solid,
				Negative: inscription,
			}
		}

		log.Println(" - creating mesh...")
		mesh := model3d.MarchingCubesSearch(solid, 0.006, 8)

		renderOrigin := mesh.Min().Mid(mesh.Max()).Add(model3d.Coord3D{Z: 2.5, Y: -1})
		if !color {
			log.Println(" - saving mesh...")
			mesh.SaveGroupedSTL("pickle.stl")
			log.Println(" - rendering...")
			render3d.SaveRendering("rendering_etched.png", mesh, renderOrigin, 500, 500, nil)
		} else {
			log.Println(" - saving mesh...")
			colorFunc := model3d.VertexColorsToTriangle(inscription.ColorAt)
			ioutil.WriteFile("pickle.zip", mesh.EncodeMaterialOBJ(colorFunc), 0755)
			log.Println(" - rendering...")
			render3d.SaveRendering("rendering_color.png", mesh, renderOrigin, 500, 500,
				render3d.TriangleColorFunc(colorFunc))
		}
	}
}

type PickleSolid struct {
	F *PickleFunction
}

func (p *PickleSolid) Min() model3d.Coord3D {
	min := p.F.Collider.Min()
	return model3d.XYZ(min.X, min.Y, -PickleWidth)
}

func (p *PickleSolid) Max() model3d.Coord3D {
	max := p.F.Collider.Max()
	return model3d.XYZ(max.X, max.Y, PickleWidth)
}

func (p *PickleSolid) Contains(c model3d.Coord3D) bool {
	radius := p.F.RadiusAt(c.Y)
	center := p.F.CenterAt(c.Y)
	dist := c.Dist(model3d.Coord3D{X: center, Y: c.Y})
	return dist < radius
}

type PickleFunction struct {
	Collider model2d.Collider

	// A synchronized map[float64][2]float64{}.
	cache sync.Map
}

func NewPickleFunction() *PickleFunction {
	bmp := model2d.MustReadBitmap("pickle.png", nil).FlipY()
	mesh := bmp.Mesh().SmoothSq(100)
	mesh = mesh.MapCoords(func(c model2d.Coord) model2d.Coord {
		return c.Scale(PickleLength / float64(bmp.Height))
	})
	collider := model2d.MeshToCollider(mesh)
	return &PickleFunction{Collider: collider}
}

func (p *PickleFunction) RadiusAt(y float64) float64 {
	min, max := p.minMaxAt(y)
	return (max - min) / 2
}

func (p *PickleFunction) CenterAt(y float64) float64 {
	min, max := p.minMaxAt(y)
	return (min + max) / 2
}

func (p *PickleFunction) minMaxAt(y float64) (float64, float64) {
	if val, ok := p.cache.Load(y); ok {
		val := val.([2]float64)
		return val[0], val[1]
	}

	r := &model2d.Ray{
		Origin:    model2d.Coord{Y: y},
		Direction: model2d.Coord{X: 1},
	}

	min := 0.0
	max := 0.0
	p.Collider.RayCollisions(r, func(rc model2d.RayCollision) {
		if rc.Scale < min || min == 0 {
			min = rc.Scale
		}
		if rc.Scale > max {
			max = rc.Scale
		}
	})

	p.cache.Store(y, [2]float64{min, max})
	return min, max
}

type Inscription struct {
	Solid model2d.Solid
}

func NewInscription() *Inscription {
	bmp := model2d.MustReadBitmap("inscription.png", nil).FlipY()
	mesh := bmp.Mesh().SmoothSq(20)
	collider := model2d.MeshToCollider(mesh)
	scale := PickleLength / float64(bmp.Height)
	return &Inscription{
		Solid: model2d.ScaleSolid(model2d.NewColliderSolid(collider), scale),
	}
}

func (i *Inscription) Min() model3d.Coord3D {
	min := i.Solid.Min()
	return model3d.XYZ(min.X, min.Y, -PickleWidth)
}

func (i *Inscription) Max() model3d.Coord3D {
	max := i.Solid.Max()
	return model3d.XYZ(max.X, max.Y, PickleWidth)
}

func (i *Inscription) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(i, c) {
		return false
	}
	return i.Solid.Contains(c.XY())
}

func (i *Inscription) ColorAt(c model3d.Coord3D) [3]float64 {
	if c.Z < 0 {
		return Green
	}
	if i.Solid.Contains(c.XY()) {
		return [3]float64{1, 1, 1}
	} else {
		return Green
	}
}
