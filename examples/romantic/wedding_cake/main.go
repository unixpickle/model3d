package main

import (
	"log"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	. "github.com/unixpickle/model3d/shorthand"
	"github.com/unixpickle/model3d/toolbox3d"
)

const GlobalScale = 0.5

func main() {
	log.Println("Creating scene...")
	layers := model3d.JoinedSolid{}
	colors := []toolbox3d.CoordColorFunc{}
	curZ := 0.0
	for _, fn := range []func() (Solid3, toolbox3d.CoordColorFunc){
		FlowerLayer, HexRoughLayer, DotsLayer, MarbleHexagonLayer, RoughRoundLayer,
	} {
		solid, color := fn()
		solid = Translate3(solid, Z(curZ))
		color = Translate3(color, Z(curZ))
		curZ = solid.Max().Z - 0.075 // Make each layer sink into the last.
		layers = append(layers, solid)
		colors = append(colors, color)
	}
	s, c := LetterCircle(layers[0])
	layers = append(layers, s)
	colors = append(colors, c)

	// Create base under the cake.
	layers = append(layers, Cylinder(
		Z(-0.2),
		Z(0.03),
		3.5,
	))
	colors = append(colors, toolbox3d.ConstantCoordColorFunc(render3d.NewColor(1.0)))

	log.Println("Creating mesh and texture...")
	mesh, interior := model3d.DualContourInterior(layers, 0.01, true, false)
	if mesh.NeedsRepair() || len(mesh.SingularVertices()) > 0 {
		panic("mesh is bad")
	}
	var solidsAndColors []any
	for i, x := range layers {
		solidsAndColors = append(solidsAndColors, x, colors[i])
	}
	fullColor := toolbox3d.JoinedSolidCoordColorFunc(
		interior,
		solidsAndColors...,
	)
	oldCount := mesh.NumTriangles()
	mesh = mesh.EliminateCoplanarFiltered(1e-5, fullColor.ChangeFilterFunc(mesh, 0.05))
	newCount := mesh.NumTriangles()
	log.Printf(" => went from %d to %d triangles", oldCount, newCount)

	mesh = mesh.Scale(GlobalScale)
	fullColor = fullColor.Transform(&model3d.Scale{Scale: GlobalScale})

	log.Println("Saving...")
	mesh.SaveMaterialOBJ("cake.zip", fullColor.Cached().QuantizedTriangleColor(mesh, 128))

	log.Println("Rendering...")
	render3d.SaveRotatingGIF("panning.gif", mesh, Z(1), model3d.XZ(1, 0.4).Normalize(),
		300, 50, 10.0, fullColor.RenderColor)
}
