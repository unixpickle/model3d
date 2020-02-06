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

func TestMeshEliminateMinimal(t *testing.T) {
	m := NewMesh()
	m.Add(&Triangle{
		Coord3D{0, 0, 1},
		Coord3D{1, 0, 0},
		Coord3D{0, 1, 0},
	})
	m.Add(&Triangle{
		Coord3D{0, 0, 0},
		Coord3D{1, 0, 0},
		Coord3D{0, 1, 0},
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
	if m.NeedsRepair() {
		t.Fatal("invalid initial mesh")
	}

	elim := m.EliminateEdges(func(m *Mesh, s Segment) bool {
		return true
	})
	if len(elim.triangles) != len(m.triangles) {
		t.Error("invalid reduction")
	}
}

func TestMeshEliminateCoplanar(t *testing.T) {
	cyl := &CylinderSolid{
		P1:     Coord3D{0, 0, -1},
		P2:     Coord3D{0, 0, 1},
		Radius: 0.5,
	}
	m1 := SolidToMesh(cyl, 0.1, 2, 0.8, 1)
	m2 := m1.EliminateCoplanar(1e-5)
	if len(m2.triangles) >= len(m1.triangles) {
		t.Fatal("reduction failed")
	}

	// Make sure the meshes have the same geometries.
	c1 := MeshToCollider(m1)
	c2 := MeshToCollider(m2)
	for i := 0; i < 1000; i++ {
		ray := &Ray{
			Origin:    Coord3D{rand.NormFloat64(), rand.NormFloat64(), rand.NormFloat64()},
			Direction: Coord3D{rand.NormFloat64(), rand.NormFloat64(), rand.NormFloat64()},
		}
		if c1.RayCollisions(ray) != c2.RayCollisions(ray) {
			t.Fatal("mismatched ray collisions", c1.RayCollisions(ray), c2.RayCollisions(ray))
		}
		r := rand.Float64()
		if c1.SphereCollision(ray.Origin, r) != c2.SphereCollision(ray.Origin, r) {
			t.Fatal("mismatched sphere collision")
		}
	}
}

func TestMeshFlatten(t *testing.T) {
	solid := JoinedSolid{
		&RectSolid{MaxVal: Coord3D{X: 2, Y: 1, Z: 0.5}},
		&RectSolid{
			MinVal: Coord3D{X: 1, Y: 1, Z: 0},
			MaxVal: Coord3D{X: 2, Y: 1, Z: 0.5},
		},
	}
	m := SolidToMesh(solid, 0.025, 0, -1, 5)
	for i := 0; i < 10; i++ {
		m = m.LassoSolid(solid, 0.025, 4, 0, 0.2)
	}
	if m.SelfIntersections() != 0 {
		t.Fatal("invalid starting mesh (intersections)")
	} else if _, n := m.RepairNormals(1e-8); n != 0 {
		t.Fatal("invalid starting mesh (normals)")
	}

	flat := m.FlattenBase(0)

	if flat.SelfIntersections() != 0 {
		t.Error("flattened mesh has self-intersections")
	} else if _, n := flat.RepairNormals(1e-8); n != 0 {
		t.Fatal("flattened mesh has invalid normals")
	}

	c1 := NewColliderSolid(MeshToCollider(m))
	c2 := NewColliderSolid(MeshToCollider(flat))
	for i := 0; i < 1000; i++ {
		p := Coord3D{X: rand.Float64(), Y: rand.Float64(), Z: rand.Float64()}
		p = p.Mul(solid.Max())
		if c1.Contains(p) && !c2.Contains(p) {
			t.Error("flattened solid is not strictly larger")
		}
	}
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

func BenchmarkMeshSmoothArea(b *testing.B) {
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
	mesh := SolidToMesh(cyl, 0.1, 2, 0.8, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mesh.EliminateCoplanar(1e-5)
	}
}
