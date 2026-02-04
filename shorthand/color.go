package shorthand

import "github.com/unixpickle/model3d/render3d"

type Color = render3d.Color

func RGB(r, g, b float64) Color {
	return render3d.NewColorRGB(r, g, b)
}

func Gray(a float64) Color {
	return render3d.NewColor(a)
}
