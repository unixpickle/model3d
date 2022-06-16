package main

import (
	"math"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

type LampLight struct {
	Mesh        *model3d.Mesh
	Object      model3d.Collider
	Solid       model3d.Solid
	Color       render3d.Color
	Samples     int
	SmoothIters int
	Amplify     float64
	Ambient     float64
}

func NewLampLight() *LampLight {
	polytope := model3d.ConvexPolytope{
		&model3d.LinearConstraint{
			Normal: model3d.YZ(-2.0, 1.0).Normalize(),
			Max:    model3d.YZ(-2.0, 1.0).Normalize().Dot(model3d.Y(-0.5)),
		},
		&model3d.LinearConstraint{
			Normal: model3d.YZ(2.0, 1.0).Normalize(),
			Max:    model3d.YZ(2.0, 1.0).Normalize().Dot(model3d.Y(0.5)),
		},
		&model3d.LinearConstraint{
			Normal: model3d.X(1),
			Max:    1.0,
		},
		&model3d.LinearConstraint{
			Normal: model3d.X(-1),
			Max:    1.0,
		},
		&model3d.LinearConstraint{
			Normal: model3d.Z(1),
			Max:    0.7,
		},
		&model3d.LinearConstraint{
			Normal: model3d.Z(-1),
			Max:    0.0,
		},
	}
	offset := model3d.Z(2)
	lampMesh := polytope.Mesh().Translate(offset)
	return &LampLight{
		Mesh:        lampMesh,
		Object:      model3d.MeshToCollider(lampMesh),
		Solid:       model3d.TranslateSolid(polytope.Solid(), offset),
		Color:       render3d.NewColorRGB(1.0, 1.0, 0.95),
		Samples:     200,
		SmoothIters: 5,
		Amplify:     10.0,
		Ambient:     0.3,
	}
}

func (l *LampLight) Recolor(s model3d.Solid, f toolbox3d.CoordColorFunc) toolbox3d.CoordColorFunc {
	size := s.Max().Sub(s.Min())
	delta := math.Max(math.Max(size.X, size.Y), size.Z) / 100.0
	mesh := model3d.MarchingCubesSearch(s, delta, 8)
	vertexColors := l.Cast(mesh)
	tree := model3d.NewCoordTree(mesh.VertexSlice())

	lampSDF := model3d.MeshToSDF(l.Mesh)

	return func(c model3d.Coord3D) render3d.Color {
		if lampSDF.SDF(c) > -0.02 {
			return l.Color
		}
		nearest := tree.NearestNeighbor(c)
		vc := vertexColors[nearest]
		origColor := f(c)
		return origColor.Scale(1 - l.Ambient).Mul(vc).Add(origColor.Scale(l.Ambient))
	}
}

func (l *LampLight) Cast(m *model3d.Mesh) map[model3d.Coord3D]render3d.Color {
	collider := model3d.MeshToCollider(m)
	vertices := m.VertexSlice()
	normals := m.VertexNormals()
	colors := make([]render3d.Color, len(vertices))

	essentials.ConcurrentMap(0, len(vertices), func(i int) {
		v := vertices[i]
		normal := normals.Value(v)

		var colorSum render3d.Color
		for i := 0; i < l.Samples; i++ {
			dir := model3d.NewCoord3DRandUnit()
			if dir.Dot(normal) < 0 {
				dir = dir.Scale(-1)
			}
			origin := v.Add(normal.Scale(1e-8))
			ray := &model3d.Ray{Origin: origin, Direction: dir}
			rc, ok := l.Object.FirstRayCollision(ray)
			if ok {
				// See if something else is in the way of the light.
				rc1, ok1 := collider.FirstRayCollision(ray)
				if ok1 && rc1.Scale < rc.Scale {
					ok = false
				}
			}
			if ok {
				colorSum = colorSum.Add(l.Color)
			}
		}
		colors[i] = colorSum.Scale(l.Amplify / float64(l.Samples))
	})

	res := map[model3d.Coord3D]render3d.Color{}
	for i, v := range vertices {
		res[v] = colors[i]
	}
	for i := 0; i < l.SmoothIters; i++ {
		newRes := map[model3d.Coord3D]render3d.Color{}
		for _, v := range vertices {
			neighborScale := 0.1
			neighbors := map[model3d.Coord3D]bool{}
			color := res[v]
			for _, t := range m.Find(v) {
				for _, c := range t {
					if c != v {
						if !neighbors[c] {
							neighbors[c] = true
							color = color.Add(res[c].Scale(neighborScale))
						}
					}
				}
			}
			color = color.Scale(1 / (1 + neighborScale*float64(len(neighbors))))
			newRes[v] = color
		}
		res = newRes
	}
	clippedRes := map[model3d.Coord3D]render3d.Color{}
	for k, v := range res {
		clippedRes[k] = v.Min(model3d.XYZ(1.0, 1.0, 1.0))
	}
	return clippedRes
}
