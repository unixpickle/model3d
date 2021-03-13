package main

import (
	"image/color"
	"log"
	"math"

	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/model3d/render3d"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

const (
	EggHeight       = 1.0
	ZigZagInset     = 0.03
	ZigZagThickness = 0.01
	NumRadialPlanes = 22
	RadialPlaneZig  = 0.03
)

func main() {
	solid := model3d.JoinedSolid{
		model3d.IntersectedSolid{
			CreateEgg(),
			model3d.JoinedSolid{
				CreateZigZag(),
				CreateRadialPlanes(),
			},
		},
		// Stand
		&model3d.Cylinder{
			P1:     model3d.Coord3D{},
			P2:     model3d.Z(0.3),
			Radius: 0.2,
		},
	}
	log.Println("Creating mesh...")

	// Fix artifacts on the base
	squeeze := &toolbox3d.AxisPinch{
		Axis:  toolbox3d.AxisZ,
		Min:   -0.01,
		Max:   0.01,
		Power: 0.25,
	}

	mesh := model3d.MarchingCubesConj(solid, 0.005, 8, squeeze)

	// Only take the outside part of the mesh,
	// since some crazy stuff is going on inside.
	log.Println("Removing mesh internals...")
	mesh = TakeOutsideMesh(mesh)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("egg.stl")
	log.Println("Rendering mesh...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func CreateEgg() model3d.Solid {
	radius := model2d.BezierCurve{
		model2d.Coord{},
		model2d.Y(0.6),
		model2d.XY(1.0, 0.3),
		model2d.X(1.0),
	}
	maxRadius := 0.0
	for t := 0.0; t < 1.0; t += 0.001 {
		maxRadius = math.Max(maxRadius, radius.EvalX(t))
	}
	return model3d.CheckedFuncSolid(
		model3d.XY(-maxRadius, -maxRadius),
		model3d.XYZ(maxRadius, maxRadius, EggHeight),
		func(c model3d.Coord3D) bool {
			rad := radius.EvalX(c.Z / EggHeight)
			return c.XY().Norm() < rad
		},
	)
}

func CreateInnerEgg() model3d.Solid {
	egg := CreateEgg()
	egg2d := model3d.CrossSectionSolid(egg, 1, 0)
	outerMesh := model2d.MarchingSquaresSearch(egg2d, 0.005, 8)
	outerCollider := model2d.MeshToCollider(outerMesh)
	solid := model2d.NewColliderSolidInset(outerCollider, ZigZagInset)
	return RevolveSolid(solid)
}

func CreateZigZag() model3d.Solid {
	egg := CreateEgg()
	egg2d := model3d.CrossSectionSolid(egg, 1, 0)
	outerMesh := model2d.MarchingSquaresSearch(egg2d, 0.005, 8)
	outerCollider := model2d.MeshToCollider(outerMesh)
	innerSolid := model2d.NewColliderSolidInset(outerCollider, ZigZagInset)
	innerMesh := model2d.MarchingSquaresSearch(innerSolid, 0.005, 8)
	innerCollider := model2d.MeshToCollider(innerMesh)

	startY := innerSolid.Max().Mid(innerSolid.Min()).Y
	zigZagMesh := model2d.NewMesh()
	for _, directionY := range []float64{1.0, -1.0} {
		ray := &model2d.Ray{
			Origin:    model2d.XY(outerCollider.Min().X-1, startY),
			Direction: model2d.X(1.0),
		}
		rc, ok := outerCollider.FirstRayCollision(ray)
		if !ok {
			panic("failed to find edge point")
		}
		ray.Origin = ray.Origin.Add(ray.Direction.Scale(rc.Scale))
		ray.Direction.Y = directionY

		zigTo := func(scale float64) {
			c1 := ray.Origin.Add(ray.Direction.Scale(scale))
			zigZagMesh.Add(&model2d.Segment{ray.Origin, c1})
			ray.Origin = c1
			ray.Direction.X *= -1
			if ray.Origin.Y > startY {
				// For the top half of the egg, it looks better
				// if the inward zigs slope at less than a 45
				// degree angle.
				if ray.Direction.X > 0 {
					height := outerCollider.Max().Y - outerCollider.Min().Y
					fracToTop := (ray.Origin.Y - startY) / (height / 2)
					ray.Direction.Y = directionY * (1 - fracToTop)
				} else {
					ray.Direction.Y = directionY
				}
			}
		}
		zigDone := func() {
			// Scale must be large enough to leave bounds.
			zigTo(0.3)
		}
		for {
			collision, ok := innerCollider.FirstRayCollision(ray)
			if !ok {
				zigDone()
				break
			}
			zigTo(collision.Scale)
			collision, ok = outerCollider.FirstRayCollision(ray)
			if !ok {
				zigDone()
				break
			}
			zigTo(collision.Scale)
		}
	}
	zigZagMesh.AddMesh(zigZagMesh.MapCoords(func(c model2d.Coord) model2d.Coord {
		return model2d.XY(-c.X, c.Y)
	}))
	bg := &model2d.Rect{MinVal: zigZagMesh.Min(), MaxVal: zigZagMesh.Max()}
	model2d.RasterizeColor("zig_zag.png", []interface{}{
		bg, outerMesh, innerMesh, zigZagMesh,
	}, []color.Color{
		color.Gray{Y: 0xff},
		color.RGBA{B: 0xff, A: 0xff},
		color.RGBA{G: 0xff, A: 0xff},
		color.RGBA{R: 0xff, A: 0xff},
	}, 200.0)
	solid := model2d.NewColliderSolidHollow(
		model2d.MeshToCollider(zigZagMesh),
		ZigZagThickness/2,
	)
	return RevolveSolid(solid)
}

func CreateRadialPlanes() model3d.Solid {
	egg := CreateEgg()
	baseSolid := model3d.CheckedFuncSolid(
		model3d.XYZ(-(ZigZagThickness+RadialPlaneZig), egg.Min().Y, 0),
		model3d.XYZ(ZigZagThickness+RadialPlaneZig, egg.Max().Y, egg.Max().Z),
		func(c model3d.Coord3D) bool {
			modZ := math.Mod(c.Z+100.0, RadialPlaneZig*2)
			if modZ > RadialPlaneZig {
				modZ = RadialPlaneZig*2 - modZ
			}
			return math.Abs(c.X-modZ) < ZigZagThickness
		},
	)
	res := model3d.JoinedSolid{}
	for i := 0; i < NumRadialPlanes; i++ {
		angle := (math.Pi * 2 / NumRadialPlanes) * float64(i)
		res = append(res, model3d.TransformSolid(
			model3d.Rotation(model3d.Z(1), angle),
			baseSolid,
		))
	}
	return res
}

func RevolveSolid(s model2d.Solid) model3d.Solid {
	min, max := s.Min(), s.Max()
	return model3d.CheckedFuncSolid(
		model3d.XYZ(min.X, min.X, min.Y),
		model3d.XYZ(max.X, max.X, max.Y),
		func(c model3d.Coord3D) bool {
			return s.Contains(model2d.XY(c.XY().Norm(), c.Z))
		},
	)
}

func TakeOutsideMesh(m *model3d.Mesh) *model3d.Mesh {
	res := model3d.NewMesh()
	min := m.Min()
	var minTriangle *model3d.Triangle
	m.Iterate(func(t *model3d.Triangle) {
		if t.Min().Z == min.Z {
			minTriangle = t
		}
	})
	queue := []*model3d.Triangle{minTriangle}
	for len(queue) > 0 {
		t := queue[0]
		m.Remove(t)
		res.Add(t)
		queue = queue[1:]
		for _, n := range m.Neighbors(t) {
			queue = append(queue, n)
			m.Remove(n)
		}
	}
	return res
}
