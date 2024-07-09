package main

import (
	"compress/gzip"
	"encoding/json"
	"flag"
	"io"
	"math"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/numerical"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

type Args struct {
	Path   string  `default:"paths/hello.json.gz"`
	LightX float64 `default:"-0.2"`
	LightY float64 `default:"-2.0"`
	LightZ float64 `default:"2.0"`

	Resolution float64 `default:"0.01"`

	MidDist   float64 `default:"2.2"`
	DistRange float64 `default:"0.2"`

	Segments  int     `default:"500"`
	Thickness float64 `default:"0.03"`
}

func main() {
	var args Args
	toolbox3d.AddFlags(&args, nil)
	flag.Parse()

	curve := PathCurve(&args)
	lightPos := model3d.XYZ(args.LightX, args.LightY, args.LightZ)

	distFunc := DistanceFunc(&args)
	pointFunc := func(t float64) model3d.Coord3D {
		xy := curve.Eval(t)
		projPoint := model3d.XZ(xy.X, xy.Y+1.0)
		ray := projPoint.Sub(lightPos).Normalize()
		return ray.Scale(distFunc(t)).Add(lightPos)
	}

	var segments []model3d.Segment
	for i := 0; i < args.Segments; i++ {
		t1 := float64(i) / float64(args.Segments)
		t2 := float64(i+1) / float64(args.Segments)
		segments = append(segments, model3d.NewSegment(pointFunc(t1), pointFunc(t2)))
	}

	const backY = 0.5

	attachToBack := func(t float64) {
		p0 := pointFunc(t)
		dir := p0.Sub(lightPos)
		scale := (backY - lightPos.Y) / dir.Y
		p1 := lightPos.Add(dir.Scale(scale))
		segments = append(segments, model3d.NewSegment(p0, p1))
	}
	attachToBack(0)
	attachToBack(1)

	maxYPoint, _ := (&numerical.LineSearch{
		Stops:      50,
		Recursions: 2,
	}).Maximize(0, 1, func(t float64) float64 {
		return curve.Eval(t).Y
	})
	attachToBack(maxYPoint)

	solid := model3d.JoinedSolid{
		toolbox3d.LineJoin(args.Thickness, segments...),
		// Back
		model3d.NewRect(
			model3d.XYZ(-1.0, backY, 0.0),
			model3d.XYZ(1.0, backY+0.1, 1.5),
		),
		// Bottom
		model3d.NewRect(
			model3d.XYZ(-1.0, -0.5, 0.0),
			model3d.XYZ(1.0, backY, 0.1),
		),
	}

	mesh := model3d.DualContour(solid, args.Resolution, true, false)
	mesh = mesh.EliminateCoplanar(1e-5)

	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
	render3d.SaveRendering("rendering_direct.png", mesh, lightPos, 512, 512, nil)

	// This looks less cool than I expected.
	// render3d.SaveRotatingGIF("rendering.gif", mesh, model3d.Z(1).ProjectOut(lightPos).Normalize(), lightPos.Scale(-1), 512, 50, 10.0, nil)

	mesh.SaveGroupedSTL("shadow_text.stl")
}

func DistanceFunc(a *Args) func(t float64) float64 {
	return func(t float64) float64 {
		return a.MidDist + a.DistRange*math.Sin(math.Sin(t*3+0.3)*math.Pi*4)
	}
}

func PathCurve(a *Args) model2d.Curve {
	path := LoadPath(a)
	path = path.Decimate(100)
	path.SmoothSq(10)
	path = path.SubdividePath(3)
	path = path.Scale(1 / (path.Max().X - path.Min().X))
	path = path.Translate(path.Min().Mid(path.Max()).Scale(-1))
	return model2d.NewSegmentCurveMesh(path)
}

func LoadPath(a *Args) *model2d.Mesh {
	f, err := os.Open(a.Path)
	essentials.Must(err)
	defer f.Close()
	r, err := gzip.NewReader(f)
	essentials.Must(err)
	data, err := io.ReadAll(r)
	essentials.Must(err)

	var path []struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}
	essentials.Must(json.Unmarshal(data, &path))

	res := model2d.NewMesh()
	for i, point := range path {
		if i > 0 {
			res.Add(&model2d.Segment{
				model2d.XY(path[i-1].X, -path[i-1].Y),
				model2d.XY(point.X, -point.Y),
			})
		}
	}
	return res
}
