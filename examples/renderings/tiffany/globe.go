package main

import (
	"image/png"
	"math"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

type Globe struct {
	render3d.Object
	Image *toolbox3d.Equirect
}

func NewGlobe() *Globe {
	r, err := os.Open("../../decoration/globe/map.png")
	essentials.Must(err)
	defer r.Close()
	mapImage, err := png.Decode(r)
	essentials.Must(err)

	return &Globe{
		Image: toolbox3d.NewEquirect(mapImage),
		Object: &render3d.ColliderObject{
			Collider: &model3d.Sphere{
				Center: model3d.Z(0.5),
				Radius: 1.5,
			},
		},
	}
}

func (g *Globe) Cast(r *model3d.Ray) (model3d.RayCollision, render3d.Material, bool) {
	collision, material, ok := g.Object.Cast(r)
	if !ok {
		return collision, material, ok
	}

	point := collision.Normal
	point = model3d.NewMatrix3Rotation(model3d.Z(1), -math.Pi/2).MulColumn(point)
	point.Y, point.Z = point.Z, -point.Y

	red, green, blue, _ := g.Image.At(point.Geo()).RGBA()
	material = &render3d.PhongMaterial{
		Alpha:         5,
		SpecularColor: render3d.NewColor(0.1),
		DiffuseColor: render3d.NewColorRGB(float64(red)/0xffff, float64(green)/0xffff,
			float64(blue)/0xffff).Scale(0.45),
	}
	return collision, material, ok
}
