package shorthand

import (
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

func Cylinder(p1, p2 C3, r float64) *model3d.Cylinder {
	return &model3d.Cylinder{P1: p2, P2: p2, Radius: r}
}

func Sphere(c C3, r float64) *model3d.Sphere {
	return &model3d.Sphere{Center: c, Radius: r}
}

func Rect2(min, max C2) *model2d.Rect {
	return model2d.NewRect(min, max)
}

func Rect3(min, max C3) *model3d.Rect {
	return model3d.NewRect(min, max)
}
