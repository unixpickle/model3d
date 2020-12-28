package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	Thickness = 0.1
	Spacing   = 0.15

	Depth  = 2.0
	Width  = 1.0
	Height = 0.8
)

func main() {
	path := CreatePath()
	rs := PathToRects(path)
	rs.Mesh().SaveGroupedSTL("table.stl")
	render3d.SaveRandomGrid("rendering.png", rs.Mesh(), 3, 3, 300, nil)
}

func CreatePath() []model3d.Coord3D {
	current := model3d.Coord3D{}
	path := []model3d.Coord3D{current}

	addOffset := func(x, y, z float64) {
		current = current.Add(model3d.XYZ(x, y, z))
		path = append(path, current)
	}

	addOffset(0, Depth-Spacing, 0)
	addOffset(-Width, 0, 0)
	addOffset(0, -Depth, 0)
	addOffset(Width-Spacing, 0, 0)
	addOffset(0, 0, Height-Spacing)
	addOffset(-Width, 0, 0)
	addOffset(0, 0, -(Height - Spacing*2))
	addOffset(0, Depth, 0)
	addOffset(0, 0, Height-Spacing)
	addOffset(0, -(Depth - Spacing), 0)
	addOffset(Width, 0, 0)
	addOffset(0, Depth, 0)
	addOffset(-(Width - Spacing), 0, 0)
	addOffset(0, 0, -(Height - Spacing))
	addOffset(Width, 0, 0)
	addOffset(0, 0, Height-Spacing*2)
	addOffset(0, -(Depth), 0)
	addOffset(0, 0, -(Height - Spacing))
	return path
}

func PathToRects(path []model3d.Coord3D) *toolbox3d.RectSet {
	res := toolbox3d.NewRectSet()
	for i := 1; i < len(path); i++ {
		p1 := path[i-1]
		p2 := path[i]
		min := p1.Min(p2).Array()
		max := p1.Max(p2).Array()
		for i := 0; i < 3; i++ {
			min[i] -= Thickness / 2
			max[i] += Thickness / 2
		}
		res.Add(&model3d.Rect{
			MinVal: model3d.NewCoord3DArray(min),
			MaxVal: model3d.NewCoord3DArray(max),
		})
	}
	return res
}
