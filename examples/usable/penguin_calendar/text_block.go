package main

import (
	"fmt"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func CreateBlocks() {
	joinedMesh := model3d.NewMesh()
	colorMap := model3d.NewCoordToCoord()

	curX := 0.0
	addBlock := func(block *model3d.Mesh, colorFunc toolbox3d.CoordColorFunc) {
		offset := block.Min().Scale(-1).Add(model3d.X(curX))
		block = block.Translate(offset)
		colorFunc = colorFunc.Transform(&model3d.Translate{Offset: offset})

		joinedMesh.AddMesh(block)
		block.IterateVertices(func(c model3d.Coord3D) {
			colorMap.Store(c, colorFunc(c))
		})

		curX = block.Max().X + 0.1
	}

	addBlock(TextBlock(model3d.XYZ(0.8, 0.8, 0.8), 0.05, 1.0/100.0, [6]string{
		"numbers/0.png",
		"numbers/1.png",
		"numbers/2.png",
		"numbers/3.png",
		"numbers/4.png",
		"numbers/5.png",
	}))
	addBlock(TextBlock(model3d.XYZ(0.8, 0.8, 0.8), 0.05, 1.0/100.0, [6]string{
		"numbers/0.png",
		"numbers/1.png",
		"numbers/6.png",
		"numbers/7.png",
		"numbers/8.png",
		"numbers/9.png",
	}))
	for i := 1; i < 12; i += 4 {
		addBlock(TextBlock(model3d.XYZ(1.6, 0.3, 0.3), 0.02, 1.0/300.0, [6]string{
			"",
			"",
			fmt.Sprintf("dates/%d.png", i),
			fmt.Sprintf("dates/%d.png", i+2),
			fmt.Sprintf("dates/%d.png", i+1),
			fmt.Sprintf("dates/%d.png", i+3),
		}))
	}

	coords := model3d.NewCoordTree(joinedMesh.VertexSlice())
	colorFunc := toolbox3d.CoordColorFunc(func(c model3d.Coord3D) render3d.Color {
		return colorMap.Value(coords.NearestNeighbor(c))
	})

	triColor := colorFunc.Cached().TriangleColor
	joinedMesh.SaveMaterialOBJ("blocks.zip", triColor)

	mp := joinedMesh.Min().Mid(joinedMesh.Max())
	render3d.SaveRendering("rendering_blocks.png", joinedMesh, mp.Add(model3d.YZ(3.0, 8.0)),
		500, 500, func(c model3d.Coord3D, rc model3d.RayCollision) render3d.Color {
			rgb := triColor(rc.Extra.(*model3d.TriangleCollision).Triangle)
			return render3d.NewColorRGB(rgb[0], rgb[1], rgb[2])
		})
}

func TextBlock(size model3d.Coord3D, inset float64, imgScale float64,
	faces [6]string) (*model3d.Mesh, toolbox3d.CoordColorFunc) {
	resMesh := model3d.NewMesh()
	insideMap := model3d.NewCoordToBool()

	for axis := 0; axis < 3; axis++ {
		var origin model3d.Coord3D
		var side1, side2 model3d.Coord3D
		var depth model3d.Coord3D

		if axis == 0 {
			side1 = model3d.Y(size.Y)
			side2 = model3d.Z(size.Z)
			depth = model3d.X(size.X)
		} else if axis == 1 {
			side1 = model3d.X(-size.X)
			side2 = model3d.Z(size.Z)
			depth = model3d.Y(size.Y)
		} else {
			side1 = model3d.X(size.X)
			side2 = model3d.Y(size.Y)
			depth = model3d.Z(size.Z)
		}
		origin = side1.Add(side2).Add(depth).Scale(-0.5)

		for side := 0; side < 2; side++ {
			boundsMax := model2d.XY(side1.Norm(), side2.Norm())
			boundsMesh := model2d.NewMeshRect(model2d.XY(0, 0), boundsMax)

			var img *model2d.Mesh
			if faces[side+axis*2] == "" {
				img = model2d.NewMesh()
			} else {
				img = model2d.MustReadBitmap("labels/"+faces[side+axis*2], nil).Mesh().SmoothSq(20)
				img = img.Scale(imgScale)
				img = img.Translate(img.Min().Mid(img.Max()).Scale(-1).Add(boundsMax.Scale(0.5)))

				if img.Max().X >= boundsMax.X || img.Max().Y >= boundsMax.Y {
					panic(fmt.Sprintf("image out of bounds for axis %d and side %d", axis, side))
				}
			}

			invertedImg := model2d.NewMesh()
			img.Iterate(func(s *model2d.Segment) {
				invertedImg.Add(&model2d.Segment{s[1], s[0]})
			})
			invertedImg.AddMesh(boundsMesh)

			invertedMesh := model2d.TriangulateMesh(invertedImg)
			imgMesh := model2d.TriangulateMesh(img)
			b1, b2, b3 := side1.Normalize(), side2.Normalize(), depth.Normalize()

			for _, tri2d := range invertedMesh {
				tri3d := &model3d.Triangle{}
				for i, c := range tri2d {
					tri3d[i] = origin.Add(b1.Scale(c.X)).Add(b2.Scale(c.Y))
				}
				resMesh.Add(tri3d)
			}
			innerMesh := model3d.NewMesh()
			for _, tri2d := range imgMesh {
				tri3d := &model3d.Triangle{}
				for i, c := range tri2d {
					c3d := origin.Add(b3.Scale(inset)).Add(b1.Scale(c.X)).Add(b2.Scale(c.Y))
					insideMap.Store(c3d, true)
					tri3d[i] = c3d
				}
				innerMesh.Add(tri3d)
			}
			innerMesh.Iterate(func(t *model3d.Triangle) {
				for i := 0; i < 3; i++ {
					p2, p1 := t[i], t[(i+1)%3]
					if len(innerMesh.Find(p1, p2)) == 1 {
						// Create a small inset that shares the same color
						// as the outside, since we color by vertex.
						for j, insetScale := range []float64{-inset * 0.9, -inset * 0.1} {
							p3, p4 := p1.Add(b3.Scale(insetScale)), p2.Add(b3.Scale(insetScale))
							resMesh.AddQuad(p1, p2, p4, p3)
							p1, p2 = p3, p4
							if j == 0 {
								insideMap.Store(p1, true)
								insideMap.Store(p2, true)
							}
						}
					}
				}
			})
			resMesh.AddMesh(innerMesh)

			origin = origin.Add(depth).Add(side1)
			side1 = side1.Scale(-1)
			depth = depth.Scale(-1)
		}
	}
	resMesh = resMesh.Repair(1e-5)

	coordTree := model3d.NewCoordTree(resMesh.VertexSlice())
	return resMesh, func(c model3d.Coord3D) render3d.Color {
		if insideMap.Value(coordTree.NearestNeighbor(c)) {
			return render3d.NewColor(0.0)
		} else {
			return render3d.NewColor(1.0)
		}
	}
}
