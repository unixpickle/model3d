package main

import (
	"flag"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	var outFile string
	var renderFile string
	var minHeight float64
	var maxHeight float64
	var msResolution float64
	var smoothIters int
	var radius float64
	var template string

	flag.StringVar(&outFile, "out", "coin.stl", "output file name")
	flag.StringVar(&renderFile, "render", "rendering.png", "rendered output file name")
	flag.Float64Var(&minHeight, "min-height", 0.1, "minimum height")
	flag.Float64Var(&maxHeight, "max-height", 0.13, "maximum height")
	flag.Float64Var(&msResolution, "resolution", 0.01,
		"resolution of marching squares, relative to radius")
	flag.IntVar(&smoothIters, "smooth-iters", 50,
		"number of mesh smoothing iterations")
	flag.Float64Var(&radius, "radius", 0.5, "radius of coin")
	flag.StringVar(&template, "template", "example.png", "coin design image")

	flag.Parse()

	log.Println("Creating 2D mesh from template...")
	mesh := ReadTemplateIntoMesh(template, msResolution, smoothIters, radius)

	mesh3d := model3d.NewMesh()

	// Create a separate 3D entity for each outer mesh.
	// There should only be one, but the template could
	// be unusual and arbitrary.
	for _, h := range model2d.MeshToHierarchy(mesh) {
		topTriangles := model2d.TriangulateMesh(h.FullMesh())

		log.Println("Triangulating bottom face...")
		bottomTriangles := model2d.TriangulateMesh(h.Mesh)

		log.Println("Triangulating mid face...")
		childTriangles := [][3]model2d.Coord{}
		for _, child := range h.Children {
			subTris := model2d.TriangulateMesh(child.FullMesh().Invert())
			childTriangles = append(childTriangles, subTris...)
		}

		log.Println("Creating faces...")
		faces := [][][3]model2d.Coord{bottomTriangles, childTriangles, topTriangles}
		for i, z := range []float64{0, minHeight, maxHeight} {
			face2d := faces[i]
			for _, t := range face2d {
				t3d := &model3d.Triangle{}
				for j, c := range t {
					t3d[j] = model3d.XYZ(c.X, c.Y, z)
				}
				if z != 0 {
					t3d[0], t3d[1] = t3d[1], t3d[0]
				}
				mesh3d.Add(t3d)
			}
		}
	}

	log.Println("Connecting top face to other faces...")
	mesh3d.Iterate(func(t *model3d.Triangle) {
		if t[0].Z == maxHeight {
			return
		}
		for i := 0; i < 3; i++ {
			s0, s1 := t[i], t[(i+1)%3]
			if len(mesh3d.Find(s0, s1)) == 1 {
				top0, top1 := s0, s1
				top0.Z = maxHeight
				top1.Z = maxHeight
				// Normals will always be correct because of
				// the ordering of the face triangles.
				mesh3d.AddQuad(s0, top0, top1, s1)
			}
		}
	})

	log.Println("Saving...")
	essentials.Must(mesh3d.SaveGroupedSTL(outFile))
	essentials.Must(render3d.SaveRandomGrid(renderFile, mesh3d, 4, 4, 200, nil))
}

func ReadTemplateIntoMesh(filename string, msResolution float64, smoothIters int,
	radius float64) *model2d.Mesh {
	bmp := model2d.MustReadBitmap(filename, func(c color.Color) bool {
		r, _, _, _ := c.RGBA()
		return r < 0xffff/2
	}).FlipY()
	m := bmp.Mesh().SmoothSq(smoothIters)
	m = m.MapCoords(m.Min().Mid(m.Max()).Sub)
	m = m.Scale(radius / math.Max(m.Max().X, m.Max().Y))

	// Re-mesh this to be constrained to a circle.
	solid := model2d.IntersectedSolid{
		model2d.NewColliderSolid(model2d.MeshToCollider(m)),
		&model2d.Circle{Radius: radius},
	}
	return model2d.MarchingSquaresSearch(solid, radius*msResolution, 8)
}
