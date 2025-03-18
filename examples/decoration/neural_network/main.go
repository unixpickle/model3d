package main

import (
	"flag"
	"math"
	"math/rand/v2"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

type Args struct {
	NeuronRadius      float64 `default:"0.2"`
	SynapseRadius     float64 `default:"0.05"`
	SynapseWiggle     float64 `default:"0.05"`
	SynapseWiggleFreq float64 `default:"2.0"`
	LayerSpace        float64 `default:"0.8"`
	NeuronSpace       float64 `default:"0.6"`
	Delta             float64 `default:"0.02"`
	BaseStemRadius    float64 `default:"0.05"`
	BaseRadius        float64 `default:"0.4"`
	BaseThickness     float64 `default:"0.1"`
}

func main() {
	var a Args
	toolbox3d.AddFlags(&a, nil)
	flag.Parse()

	smallLayerHeight := a.NeuronRadius*2 + a.NeuronSpace
	bigLayerHeight := a.NeuronRadius*2 + a.NeuronSpace*2
	smallLayerZ := (bigLayerHeight - smallLayerHeight) / 2

	var solid model3d.JoinedSolid
	layers := [][]model3d.Coord3D{
		{
			model3d.XYZ(0.0, 0.0, smallLayerZ),
			model3d.XYZ(0.0, 0.0, smallLayerZ+a.NeuronSpace),
		},
		{
			model3d.XYZ(a.LayerSpace, 0.0, 0.0),
			model3d.XYZ(a.LayerSpace, 0.0, a.NeuronSpace),
			model3d.XYZ(a.LayerSpace, 0.0, 2*a.NeuronSpace),
		},
		{
			model3d.XYZ(a.LayerSpace*2, 0.0, smallLayerZ),
			model3d.XYZ(a.LayerSpace*2, 0.0, smallLayerZ+a.NeuronSpace),
		},
	}
	for _, layer := range layers {
		for _, c := range layer {
			solid = append(solid, &model3d.Sphere{Center: c, Radius: a.NeuronRadius})
		}
	}
	for sourceLayer := 0; sourceLayer < 2; sourceLayer++ {
		layer1 := layers[sourceLayer]
		layer2 := layers[sourceLayer+1]
		for _, c1 := range layer1 {
			for _, c2 := range layer2 {
				solid = append(solid, Synapse(&a, c1, c2))
			}
		}
	}
	for _, neuron := range layers[0] {
		solid = append(solid, Synapse(&a, neuron, neuron.Sub(model3d.X(a.NeuronRadius*1.5))))
	}
	for _, neuron := range layers[2] {
		solid = append(solid, Synapse(&a, neuron, neuron.Add(model3d.X(a.NeuronRadius*1.5))))
	}
	baseStart := layers[1][0]
	baseX := baseStart.X
	baseStart.X = 0
	solid = append(
		solid,
		model3d.TranslateSolid(
			model3d.JoinedSolid{
				&model3d.Cylinder{
					P1:     baseStart,
					P2:     baseStart.Add(model3d.Z(-0.5)),
					Radius: a.BaseStemRadius,
				},
				model3d.VecScaleSolid(&model3d.Cylinder{
					P1:     baseStart.Add(model3d.Z(-0.5)),
					P2:     baseStart.Add(model3d.Z(-0.5 + -a.BaseThickness)),
					Radius: a.BaseRadius,
				}, model3d.XYZ(2.0, 1.0, 1.0)),
			},
			model3d.X(baseX),
		),
	)
	mesh := model3d.DualContour(solid, a.Delta, false, false)
	render3d.SaveRendering("rendering.png", mesh, model3d.XYZ(a.LayerSpace*1.5, -a.LayerSpace*4, a.LayerSpace*3), 512, 512, nil)
	mesh.SaveGroupedSTL("neural_network.stl")
}

func Synapse(a *Args, start, end model3d.Coord3D) model3d.Solid {
	dist := end.Dist(start)
	direction := end.Sub(start).Normalize()

	offset := rand.Float64()

	point := func(t float64) model3d.Coord3D {
		linePoint := start.Add(direction.Scale(t))
		z := math.Sin(t*math.Pi*2*dist*a.SynapseWiggleFreq+offset) * a.SynapseWiggle
		return linePoint.Add(model3d.Y(z))
	}

	var segments []model3d.Segment
	delta := 0.02
	for t := 0.0; t < dist; t += delta {
		segments = append(segments, model3d.NewSegment(point(t), point(t+delta)))
	}

	return toolbox3d.LineJoin(a.SynapseRadius, segments...)
}
