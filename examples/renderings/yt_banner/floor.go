package main

import (
	"image"
	"image/jpeg"
	"os"

	"github.com/unixpickle/model3d/model2d"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func CreateFloor() render3d.Object {
	return NewFloorObject(&render3d.ColliderObject{
		Collider: &model3d.Rect{
			MinVal: model3d.Coord3D{X: -100, Y: -100, Z: -0.01},
			MaxVal: model3d.Coord3D{X: 100, Y: 100, Z: 0},
		},
	})
}

type FloorObject struct {
	render3d.Object

	Texture image.Image
	Scale   float64
}

func NewFloorObject(obj render3d.Object) *FloorObject {
	r, err := os.Open("assets/marble.jpg")
	essentials.Must(err)
	defer r.Close()
	img, err := jpeg.Decode(r)
	essentials.Must(err)

	return &FloorObject{
		Object:  obj,
		Texture: img,
		Scale:   float64(img.Bounds().Dx()) / 4,
	}
}

func (f *FloorObject) Cast(ray *model3d.Ray) (model3d.RayCollision, render3d.Material, bool) {
	rc, mat, ok := f.Object.Cast(ray)
	if !ok {
		return rc, mat, ok
	}
	p := ray.Origin.Add(ray.Direction.Scale(rc.Scale))
	p2 := p.Coord2D()

	// Rotate the texture so it doesn't look repeated
	// in the distance.
	p2 = model2d.NewMatrix2Rotation(0.3).MulColumn(p2)

	// For modulus consistency around 0.
	p2.X += 1000
	p2.Y += 1000

	p2 = p2.Scale(f.Scale)
	x := int(p2.X)
	y := int(p2.Y)

	// Reflect the texture in both directions.
	bounds := f.Texture.Bounds()
	if (x/bounds.Dx())%2 == 0 {
		x = bounds.Dx() - (x % bounds.Dx()) - 1
	} else {
		x = x % bounds.Dx()
	}
	if (y/bounds.Dy())%2 == 0 {
		y = bounds.Dy() - (y % bounds.Dy()) - 1
	} else {
		y = y % bounds.Dy()
	}

	r, g, b, _ := f.Texture.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()
	color := render3d.NewColorRGB(float64(r)/0xffff, float64(g)/0xffff,
		float64(b)/0xffff)

	return rc, &render3d.PhongMaterial{
		Alpha:         100,
		SpecularColor: color.Scale(0.5),
		DiffuseColor:  color.Scale(0.5),
	}, ok
}
