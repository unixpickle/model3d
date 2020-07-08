package model3d

import (
	"math"
	"math/rand"
	"testing"
)

func TestMeshRepair(t *testing.T) {
	t.Run("EdgeCase", func(t *testing.T) {
		m := NewMesh()
		// An example where the numbers round to different
		// things even though they are close.
		// Numbers are 1.7164450046354633 and
		// 1.7164449974385279.
		m.Add(&Triangle{
			{2.8934311810738533, 1.8152061242737787, 1.5906772555075124},
			{0, 0, 0},
			{2.9520256962330107, 1.7164450046354633, 1.6228898626401937},
		})
		m.Add(&Triangle{
			{2.8934311810738533, 1.8152061242737787, 1.5906772555075124},
			{2.95202569111261, 1.7164449974385279, 1.6228898570817343},
			{1, 1, 1},
		})
		m1 := m.Repair(1e-5)
		tris := m1.TriangleSlice()
		if tris[0][1].X != 0 {
			tris[0], tris[1] = tris[1], tris[0]
		}
		if len(m1.Find(tris[0][0], tris[0][2])) != 2 {
			t.Fatal("Repair failed", tris[0][0], tris[0][2], tris[1][0], tris[1][1])
		}
	})
	t.Run("Large", func(t *testing.T) {
		m := NewMesh()
		NewMeshPolar(func(g GeoCoord) float64 {
			return 3 + math.Cos(g.Lat)*math.Sin(g.Lon)
		}, 100).Iterate(func(t *Triangle) {
			t[0].X += rand.Float64() * 1e-8
			t[0].Y += rand.Float64() * 1e-8
			t[0].Z += rand.Float64() * 1e-8
			m.Add(t)
		})
		if !m.NeedsRepair() {
			t.Error("should need repair")
		}
		if m.Repair(1e-5).NeedsRepair() {
			t.Error("should not need repair")
		}
	})
}

func TestMeshRepairNormals(t *testing.T) {
	mesh := NewMeshPolar(func(g GeoCoord) float64 {
		return 1 + math.Sin(g.Lat*4)*0.1 + math.Cos(g.Lon*4)*0.13
	}, 30)

	mesh1, numRepairs := mesh.RepairNormals(1e-8)
	if numRepairs > 0 {
		t.Errorf("expected 0 repairs but got: %d", numRepairs)
	}
	if !meshesEqual(mesh, mesh1) {
		t.Error("meshes are not equal")
	}

	flipped := NewMesh()
	expectedFlipped := 0
	mesh.Iterate(func(t *Triangle) {
		if rand.Intn(2) == 0 {
			flipped.Add(t)
		} else {
			t1 := *t
			t1[0], t1[2] = t1[2], t1[0]
			flipped.Add(&t1)
			expectedFlipped++
		}
	})
	mesh1, numRepairs = flipped.RepairNormals(1e-8)
	if numRepairs != expectedFlipped {
		t.Errorf("expected %d repairs but got %d", expectedFlipped, numRepairs)
	}
	if !meshesEqual(mesh, mesh1) {
		t.Error("meshes are not equal")
	}
}

func TestMeshEliminateMinimal(t *testing.T) {
	m := NewMesh()
	m.Add(&Triangle{
		Coord3D{0, 0, 1},
		Coord3D{1, 0, 0},
		Coord3D{0, 1, 0},
	})
	m.Add(&Triangle{
		Coord3D{0, 0, 0},
		Coord3D{0, 1, 0},
		Coord3D{1, 0, 0},
	})
	m.Add(&Triangle{
		Coord3D{0, 0, 0},
		Coord3D{0, 0, 1},
		Coord3D{0, 1, 0},
	})
	m.Add(&Triangle{
		Coord3D{0, 0, 0},
		Coord3D{1, 0, 0},
		Coord3D{0, 0, 1},
	})

	// Sanity check.
	MustValidateMesh(t, m, true)

	elim := m.EliminateEdges(func(m *Mesh, s Segment) bool {
		return true
	})
	if !meshesEqual(elim, m) {
		t.Error("invalid reduction: meshes not equal")
	}
}

func TestMeshEliminateCoplanar(t *testing.T) {
	cyl := &CylinderSolid{
		P1:     Coord3D{0, 0, -1},
		P2:     Coord3D{0, 0, 1},
		Radius: 0.5,
	}
	m1 := MarchingCubesSearch(cyl, 0.025, 32)
	m2 := m1.EliminateCoplanar(1e-8)
	if len(m2.triangles) >= len(m1.triangles) {
		t.Fatal("reduction failed")
	}
	if _, n := m2.RepairNormals(1e-8); n != 0 {
		t.Error("reduction has bad normals")
	}

	// Make sure the meshes have the same geometries.
	s1 := MeshToSDF(m1)
	s2 := MeshToSDF(m2)
	for i := 0; i < 1000; i++ {
		origin := NewCoord3DRandNorm()
		if math.Abs(s1.SDF(origin)-s2.SDF(origin)) > 1e-5 {
			t.Fatal("mismatched SDFs", s1.SDF(origin), s2.SDF(origin))
		}
	}
}

func TestMeshFlipDelaunay(t *testing.T) {
	mesh := testingNonDelaunayMesh()
	isDelaunay := func(m *Mesh) bool {
		result := true
		m.Iterate(func(t *Triangle) {
			for _, seg := range t.Segments() {
				tris := m.Find(seg[0], seg[1])
				if len(tris) != 2 {
					return
				}
				var sum float64
				for _, t := range tris {
					other := seg.other(t)
					v1 := seg[0].Sub(other)
					v2 := seg[1].Sub(other)
					sum += math.Acos(v1.Normalize().Dot(v2.Normalize()))
				}
				if sum > math.Pi+2e-8 {
					result = false
				}
			}
		})
		return result
	}
	if isDelaunay(mesh) {
		t.Fatal("initial mesh should be non-delaunay")
	}
	mesh1 := mesh.FlipDelaunay()
	if !isDelaunay(mesh1) {
		t.Fatal("flipped mesh is non-delaunay")
	}
	MustValidateMesh(t, mesh1, false)
	verts1 := mesh.VertexSlice()
	verts2 := mesh1.VertexSlice()
	if len(verts1) != len(verts2) {
		t.Fatal("vertex count is different")
	}
	v1Set := map[Coord3D]bool{}
	for _, v := range verts1 {
		v1Set[v] = true
	}
	for _, v := range verts2 {
		if !v1Set[v] {
			t.Fatal("vertices are different")
		}
	}
}

func meshesEqual(m1, m2 *Mesh) bool {
	seg1 := meshOrderedSegments(m1)
	seg2 := meshOrderedSegments(m2)
	if len(seg1) != len(seg2) {
		return false
	}
	for s, c := range seg1 {
		if seg2[s] != c {
			return false
		}
	}
	return true
}

func meshOrderedSegments(m *Mesh) map[[2]Coord3D]int {
	res := map[[2]Coord3D]int{}
	m.Iterate(func(t *Triangle) {
		for i := 0; i < 3; i++ {
			seg := [2]Coord3D{t[i], t[(i+1)%3]}
			res[seg]++
		}
	})
	return res
}

func TestMeshFlattenBase(t *testing.T) {
	t.Run("Topology", func(t *testing.T) {
		m := readNonIntersectingHook()
		flat := m.FlattenBase(0)
		MustValidateMesh(t, flat, true)
	})

	t.Run("Containment", func(t *testing.T) {
		solid := JoinedSolid{
			&RectSolid{MaxVal: XYZ(2, 1, 0.5)},
			&RectSolid{
				MinVal: XYZ(1, 1, 0),
				MaxVal: XYZ(2, 1, 0.5),
			},
		}
		m := MarchingCubesSearch(solid, 0.025, 8).Blur(-1, -1, -1, -1, -1)
		flat := m.FlattenBase(0)
		c1 := NewColliderSolid(MeshToCollider(m))
		c2 := NewColliderSolid(MeshToCollider(flat))
		for i := 0; i < 1000; i++ {
			p := XYZ(rand.Float64(), rand.Float64(), rand.Float64())
			p = p.Mul(solid.Max())
			if c1.Contains(p) && !c2.Contains(p) {
				t.Error("flattened solid is not strictly larger")
			}
		}
	})
}

func BenchmarkMeshBlur(b *testing.B) {
	m := NewMeshPolar(func(g GeoCoord) float64 {
		return 3 + math.Cos(g.Lat)*math.Sin(g.Lon)
	}, 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Blur(0.8, 0.8, 0.8, 0.8, 0.8, 0.8, 0.8)
	}
}

func BenchmarkMeshSmoothAreas(b *testing.B) {
	m := NewMeshPolar(func(g GeoCoord) float64 {
		return 3 + math.Cos(g.Lat)*math.Sin(g.Lon)
	}, 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.SmoothAreas(0.1, 7)
	}
}

func BenchmarkMeshRepair(b *testing.B) {
	m := NewMesh()
	NewMeshPolar(func(g GeoCoord) float64 {
		return 3 + math.Cos(g.Lat)*math.Sin(g.Lon)
	}, 100).Iterate(func(t *Triangle) {
		t[0].X += rand.Float64() * 1e-8
		t[0].Y += rand.Float64() * 1e-8
		t[0].Z += rand.Float64() * 1e-8
		m.Add(t)
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Repair(1e-5)
	}
}

func BenchmarkEliminateCoplanar(b *testing.B) {
	cyl := &CylinderSolid{
		P1:     Coord3D{0, 1, -1},
		P2:     Coord3D{0, 1, 1},
		Radius: 0.5,
	}
	mesh := MarchingCubesSearch(cyl, 0.025, 8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mesh.EliminateCoplanar(1e-5)
	}
}

func BenchmarkMeshFlattenBase(b *testing.B) {
	solid := JoinedSolid{
		&RectSolid{MaxVal: XYZ(2, 1, 0.5)},
		&RectSolid{
			MinVal: XYZ(1, 1, 0),
			MaxVal: XYZ(2, 1, 0.5),
		},
	}
	m := MarchingCubesSearch(solid, 0.025, 8).Blur(-1, -1, -1, -1, -1, -1, -1, -1, -1, -1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.FlattenBase(0)
	}
}

func BenchmarkMeshFlipDelaunay(b *testing.B) {
	mesh := testingNonDelaunayMesh()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mesh.FlipDelaunay()
	}
}

func testingNonDelaunayMesh() *Mesh {
	return MarchingCubesSearch(JoinedSolid{
		&Cylinder{
			P1:     XY(0.2, 0.3),
			P2:     XZ(0.3, 0.5),
			Radius: 0.1,
		},
		&Cylinder{
			P1:     X(0.2),
			P2:     XZ(0.3, 0.5),
			Radius: 0.1,
		},
		&Sphere{Center: XZ(0.25, 0.25), Radius: 0.2},
	}, 0.02, 8)
}
