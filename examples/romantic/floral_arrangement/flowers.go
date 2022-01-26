package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

type Flower struct {
	Solid     model3d.Solid
	ColorSDF  model3d.SDF
	ColorFunc toolbox3d.CoordColorFunc
	Tilt      float64
}

func NewBermudaButtercup() *Flower {
	// https://upload.wikimedia.org/wikipedia/commons/thumb/a/a5/Flower_poster_2.jpg/1920px-Flower_poster_2.jpg
	depthCurve := model2d.BezierCurve{
		model2d.XY(0.0, 0.3),
		model2d.XY(0.8, 0.3),
		model2d.XY(0.5, 0.0),
		model2d.XY(1.0, 0.0),
	}
	z := func(r float64) float64 {
		return depthCurve.EvalX(math.Max(0, math.Min(1, 1-r)))
	}
	r := func(th float64) float64 {
		return math.Max(0.3, math.Abs(math.Cbrt(math.Sin(2.5*th))))
	}
	heightMesh := polarHeightMap(0.03, r, z)
	s := model3d.NewColliderSolidHollow(model3d.MeshToCollider(heightMesh), 0.1)
	tilt := math.Pi / 4
	tilted := model3d.RotateSolid(s, model3d.Y(1), tilt)
	return &Flower{
		Solid:     tilted,
		ColorSDF:  ColorFuncSDF(tilted),
		ColorFunc: toolbox3d.ConstantCoordColorFunc(render3d.NewColorRGB(1.0, 1.0, 0.0)),
		Tilt:      tilt,
	}
}

func NewRose() *Flower {
	pedal := func(radius, startFrac, span float64) *model3d.Mesh {
		thetaStart := 2 * math.Pi * startFrac
		thetaEnd := thetaStart + 2*math.Pi*span
		r := func(theta float64) float64 {
			if theta < thetaStart {
				theta += math.Pi * 2
			}
			if theta > thetaEnd {
				return -1
			}
			frac := (theta - thetaStart) / (thetaEnd - thetaStart)
			return radius * (1 - math.Pow(frac, 6))
		}
		z := func(curRad float64) float64 {
			return math.Pow(curRad/radius, 6)
		}
		return polarHeightMap(0.03, r, z)
	}
	pedals := []*model3d.Mesh{
		pedal(0.2, 0, 0.4),
		pedal(0.25, 0.3, 0.4),
		pedal(0.18, 0.65, 0.4),
		pedal(0.4, 0.3, 0.5),
		pedal(0.4, 0.6, 0.2),
		pedal(0.4, 0.9, 0.2),
		pedal(0.6, 0.15, 0.25),
		pedal(0.65, 0.1, 0.5),
		pedal(0.63, 0.4, 0.2),
		pedal(0.65, 0.7, 0.2),
		pedal(0.62, 0.95, 0.25),
	}
	combined := model3d.NewMesh()
	for _, m := range pedals {
		combined.AddMesh(m)
	}
	solid := model3d.NewColliderSolidHollow(model3d.MeshToCollider(combined), 0.1)
	tilt := math.Pi / 3.5
	tilted := model3d.RotateSolid(solid, model3d.Y(1), tilt)
	return &Flower{
		Solid:     tilted,
		ColorSDF:  ColorFuncSDF(tilted),
		ColorFunc: toolbox3d.ConstantCoordColorFunc(render3d.NewColorRGB(1.0, 0.0, 0.0)),
		Tilt:      tilt,
	}
}

func NewPurpleRowFlower() *Flower {
	// Based on https://commons.wikimedia.org/wiki/File:Purple_flower_(4764445139).jpg.
	// Color: #dd4dcd

	pedalShape := model2d.BezierCurve{
		model2d.XY(1.0, 0),
		model2d.XY(1.0, 0.3*0.7),
		model2d.XY(0.2, 0.2*0.7),
		model2d.XY(0.0, 0.05*0.7),
	}
	solid2d := model2d.CheckedFuncSolid(
		model2d.XY(0, -1),
		model2d.XY(1, 1),
		func(c model2d.Coord) bool {
			y := pedalShape.EvalX(c.X)
			return math.Abs(c.Y) <= y
		},
	)
	mesh2d := model2d.MarchingSquaresSearch(solid2d, 0.03, 8)

	depthCurve := model2d.BezierCurve{
		model2d.XY(0.0, 0.0),
		model2d.XY(0.5, 0.15),
		model2d.XY(1.0, 0.2),
	}
	pedalMesh := heightMap(mesh2d, 0.03, func(c model2d.Coord) float64 {
		return depthCurve.EvalX(c.Norm())
	})
	pedals := model3d.NewMesh()
	for i := 0; i < 8; i++ {
		theta := 2 * math.Pi * float64(i) / 8
		theta1 := theta + 2*math.Pi/16
		pedals.AddMesh(pedalMesh.Rotate(model3d.Z(1), theta))
		pedals.AddMesh(pedalMesh.Rotate(model3d.Z(1), theta1).Translate(model3d.Z(0.1)))
	}

	middleDiscCircle := &model3d.Cylinder{
		P1:     model3d.Z(0.2),
		P2:     model3d.Z(0.21),
		Radius: 0.2,
	}
	middleDisc := model3d.NewColliderSolidInset(middleDiscCircle, -0.15)

	solid := model3d.JoinedSolid{
		middleDisc,
		model3d.NewColliderSolidHollow(model3d.MeshToCollider(pedals), 0.1),
	}

	colorFunc := toolbox3d.CoordColorFunc(func(c model3d.Coord3D) render3d.Color {
		if middleDiscCircle.SDF(c) > -(0.15 + 0.005) {
			return render3d.NewColor(0.1)
		} else {
			return render3d.NewColorRGB(0xdd/255.0, 0x4d/255.0, 0xcd/255.0)
		}
	})

	tilt := math.Pi / 3.5
	xform := model3d.Rotation(model3d.Y(1), tilt)
	tilted := model3d.TransformSolid(xform, solid)
	return &Flower{
		Solid:     tilted,
		ColorSDF:  ColorFuncSDF(tilted),
		ColorFunc: colorFunc.Transform(xform),
		Tilt:      tilt,
	}
}

func (f *Flower) Place(pos model3d.Coord3D) *Flower {
	xform := model3d.JoinedTransform{
		model3d.Rotation(model3d.Z(1), math.Atan2(pos.Y, pos.X)),
		&model3d.Translate{Offset: pos},
	}
	return &Flower{
		Solid:     model3d.TransformSolid(xform, f.Solid),
		ColorSDF:  model3d.TransformSDF(xform, f.ColorSDF),
		ColorFunc: f.ColorFunc.Transform(xform),
		Tilt:      f.Tilt,
	}
}

func polarHeightMap(delta float64, r func(theta float64) float64, z func(r float64) float64) *model3d.Mesh {
	solid2d := model2d.CheckedFuncSolid(
		model2d.XY(-2, -2),
		model2d.XY(2, 2),
		func(c model2d.Coord) bool {
			return c.Norm() <= r(math.Atan2(c.Y, c.X))
		},
	)
	mesh2d := model2d.MarchingSquaresSearch(solid2d, delta, 8)

	return heightMap(mesh2d, delta, func(c model2d.Coord) float64 {
		return z(c.Norm())
	})
}

func heightMap(mesh2d *model2d.Mesh, delta float64, z func(c model2d.Coord) float64) *model3d.Mesh {
	// There may be some small artifacts/holes in the mesh, so
	// we use the outermost ring only.
	mesh2d = model2d.MeshToHierarchy(mesh2d)[0].Mesh

	polygon := []model3d.Coord3D{}
	c1 := mesh2d.VertexSlice()[0]
	for {
		segs := mesh2d.Find(c1)
		if len(segs) == 0 {
			break
		}
		var seg *model2d.Segment
		for _, s := range segs {
			if s[0] == c1 {
				seg = s
			}
		}
		if seg == nil {
			panic("mesh not manifold")
		}
		mesh2d.Remove(seg)
		polygon = append(polygon, model3d.XY(c1.X, c1.Y))
		c1 = seg[1]
	}
	if len(mesh2d.VertexSlice()) != 0 {
		panic("2d shape is not a single polygon")
	}
	mesh3d := model3d.NewMeshTriangles(model3d.TriangulateFace(polygon))
	mesh3d = mesh3d.MapCoords(func(c model3d.Coord3D) model3d.Coord3D {
		c.Z = z(c.XY())
		return c
	})
	for {
		subdiv := model3d.NewSubdivider()
		subdiv.AddFiltered(mesh3d, func(p1, p2 model3d.Coord3D) bool {
			return p1.Dist(p2) > delta
		})
		if subdiv.NumSegments() == 0 {
			break
		}
		subdiv.Subdivide(mesh3d, func(p1, p2 model3d.Coord3D) model3d.Coord3D {
			mp := p1.Mid(p2)
			mp.Z = z(mp.XY())
			return mp
		})
	}
	return mesh3d
}
