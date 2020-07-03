// This is an experiment to see if a convex polytope
// can somewhat model the shape of the earth. It was
// a failure.
package main

import (
	"image/png"
	"log"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const LandDepth = 0.03

func main() {
	f, err := os.Open("map.png")
	essentials.Must(err)
	defer f.Close()
	img, err := png.Decode(f)
	essentials.Must(err)
	eq := toolbox3d.NewEquirect(img)

	log.Println("Creating polytope...")
	var poly model3d.ConvexPolytope
	for i := 0; i < 10000; i++ {
		norm := model3d.NewCoord3DRandUnit()
		radius := 1.0
		if r, _, _, _ := eq.At(norm.Geo()).RGBA(); r > 0xf000 {
			radius += LandDepth
		}
		poly = append(poly, &model3d.LinearConstraint{
			Normal: norm,
			Max:    radius,
		})
	}

	// Using the polytope's Mesh() method results in a
	// truly gigantic mesh that takes a long time to
	// build. Instead, we use an algorithm that does not
	// produce larger meshes for larger polytopes.
	log.Println("Creating mesh...")
	maxVal := model3d.XYZ(1+LandDepth, 1+LandDepth, 1+LandDepth)
	solid := &PolySolid{Polytope: poly, MinVal: maxVal.Scale(-1), MaxVal: maxVal}
	mesh := model3d.MarchingCubesSearch(solid, 0.01, 8)

	log.Println("Saving...")
	mesh.SaveGroupedSTL("mesh.stl")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

type PolySolid struct {
	Polytope model3d.ConvexPolytope
	MinVal   model3d.Coord3D
	MaxVal   model3d.Coord3D
}

func (p *PolySolid) Min() model3d.Coord3D {
	return p.MinVal
}

func (p *PolySolid) Max() model3d.Coord3D {
	return p.MaxVal
}

func (p *PolySolid) Contains(c model3d.Coord3D) bool {
	return model3d.InBounds(p, c) && p.Polytope.Contains(c)
}
