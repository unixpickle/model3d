package main

import (
	"flag"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/render3d"
)

type Args struct {
	Out          string  `default:"coin.stl" usage:"output filename"`
	Render       string  `default:"rendering.png" usage:"rendered output filename"`
	MinHeight    float64 `default:"0.1" usage:"minimum height"`
	MaxHeight    float64 `default:"0.13" usage:"maximum height"`
	Resolution   float64 `default:"0.01" usage:"resolution of marching squares, relative to radius"`
	McDelta      float64 `default:"0.01" usage:"resolution of marching cubes when rounding the solid"`
	SmoothIters  int     `default:"50" usage:"number of mesh smoothing iterations"`
	Radius       float64 `default:"0.5" usage:"radius of coin"`
	Template     string  `default:"example.png" usage:"coin design image"`
	Rounded      bool    `default:"false" usage:"use a rounded design instead of flat"`
	RoundSamples int     `default:"10000" usage:"number of samples to use for rounded design"`
}

func main() {
	var args Args
	toolbox3d.AddFlags(&args, nil)
	flag.Parse()

	log.Println("Creating 2D mesh from template...")
	mesh := ReadTemplateIntoMesh(args.Template, args.Resolution, args.SmoothIters, args.Radius)

	var mesh3d *model3d.Mesh
	if args.Rounded {
		mesh3d = RoundedModel(mesh, args.MinHeight, args.MaxHeight, args.McDelta, args.RoundSamples)
	} else {
		mesh3d = UnroundedModel(mesh, args.MinHeight, args.MaxHeight)
	}

	log.Println("Saving...")
	essentials.Must(mesh3d.SaveGroupedSTL(args.Out))
	essentials.Must(render3d.SaveRandomGrid(args.Render, mesh3d, 4, 4, 200, nil))
}

func UnroundedModel(mesh *model2d.Mesh, minHeight, maxHeight float64) *model3d.Mesh {
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

	return mesh3d
}

func RoundedModel(mesh *model2d.Mesh, minHeight, maxHeight, mcDelta float64,
	roundSamples int) *model3d.Mesh {
	// Create the outline for the base.
	log.Println("Creating outline solid...")
	outerMeshes := model2d.MeshToHierarchy(mesh)
	outerSolid := model2d.JoinedSolid{}
	for _, m := range outerMeshes {
		outerSolid = append(outerSolid, model2d.NewColliderSolid(model2d.MeshToCollider(m.Mesh)))
	}

	// Create heightmap for the etching on the top.
	log.Println("Creating heightmap...")
	sdf2d := model2d.MeshToSDF(mesh)
	hm := toolbox3d.NewHeightMap(sdf2d.Min(), sdf2d.Max(), 1024)
	for i := 0; i < roundSamples; i++ {
		p := model2d.NewCoordRandBounds(hm.Min, hm.Max)
		if sdf2d.SDF(p) < 0 {
			i--
			continue
		}
		proj := model2d.ProjectMedialAxis(sdf2d, p, 16, 0)
		hm.AddSphereFill(proj, sdf2d.SDF(proj), maxHeight-minHeight)
	}

	log.Println("Creating mesh...")
	solid := model3d.JoinedSolid{
		model3d.ProfileSolid(outerSolid, -minHeight, 1e-5),
		toolbox3d.HeightMapToSolid(hm),
	}
	mesh3d := model3d.MarchingCubesSearch(solid, mcDelta, 8)

	log.Println("Simplifying mesh...")
	mesh3d = mesh3d.EliminateCoplanar(1e-5)

	return mesh3d
}

func ReadTemplateIntoMesh(filename string, msResolution float64, smoothIters int,
	radius float64) *model2d.Mesh {
	bmp := model2d.MustReadBitmap(filename, func(c color.Color) bool {
		r, _, _, _ := c.RGBA()
		return r < 0xffff/2
	}).FlipY()
	m := bmp.Mesh().SmoothSq(smoothIters)
	m = m.MapCoords(m.Min().Mid(m.Max()).Sub)
	m = m.Scale(radius / m.Max().MaxCoord())

	// Re-mesh this to be constrained to a circle.
	solid := model2d.IntersectedSolid{
		model2d.NewColliderSolid(model2d.MeshToCollider(m)),
		&model2d.Circle{Radius: radius},
	}
	return model2d.MarchingSquaresSearch(solid, radius*msResolution, 8)
}
