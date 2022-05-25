package model3d

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

func TestDualContouringBasic(t *testing.T) {
	runTests := func(t *testing.T, gos, bufSize int) {
		t.Run("Sphere", func(t *testing.T) {
			solid := &Sphere{Radius: 1.0}
			dc := &DualContouring{
				S:          SolidSurfaceEstimator{Solid: solid},
				Delta:      0.04,
				MaxGos:     gos,
				BufferSize: bufSize,
			}
			mesh := dc.Mesh()
			// In general, meshes from Dual Contouring might not be
			// manifold, but this one should be.
			MustValidateMesh(t, mesh, false)

			volume := mesh.Volume()
			expected := 4.0 / 3.0 * math.Pi
			if math.Abs(volume-expected) > 5e-2 {
				t.Errorf("expected volume %f but got %f", expected, volume)
			}
		})

		t.Run("Rect", func(t *testing.T) {
			solid := NewRect(Ones(-1), Ones(1))
			dc := &DualContouring{
				S:          SolidSurfaceEstimator{Solid: solid},
				Delta:      0.04,
				MaxGos:     gos,
				BufferSize: bufSize,
			}
			mesh := dc.Mesh()
			// In general, meshes from Dual Contouring might not be
			// manifold, but this one should be.
			MustValidateMesh(t, mesh, false)

			volume := mesh.Volume()
			expected := 2.0 * 2.0 * 2.0
			if math.Abs(volume-expected) > 1e-2 {
				t.Errorf("expected volume %f but got %f", expected, volume)
			}
		})
	}
	for _, gos := range []int{0, 1, 8} {
		for _, buf := range []int{5000, 1000000} {
			t.Run(fmt.Sprintf("MaxGos%dBuf%d", gos, buf), func(t *testing.T) {
				runTests(t, gos, buf)
			})
		}
	}
}

func TestDualContouringSingular(t *testing.T) {
	runTransform := func(t *testing.T, transform Transform) {
		var solid Solid
		solid = JoinedSolid{
			NewRect(XYZ(0, 0, 0), XYZ(1, 1, 1)),
			NewRect(XYZ(1, 0, 1), XYZ(2, 1, 2)),
			NewRect(XYZ(-1, -1, -1), XYZ(0, 0, 0)),
		}
		if transform != nil {
			solid = TransformSolid(transform, solid)
		}
		dc := &DualContouring{
			S:      SolidSurfaceEstimator{Solid: solid},
			Delta:  0.04,
			Repair: false,
		}
		badMesh := dc.Mesh()
		if !badMesh.NeedsRepair() {
			t.Fatal("bad mesh should have singular edges")
		}
		if len(badMesh.SingularVertices()) == 0 {
			t.Fatal("bad mesh should have singular vertices")
		}
		dc.Repair = true
		goodMesh := dc.Mesh()
		if goodMesh.NeedsRepair() {
			t.Error("good mesh should have no singular edges")
		}
		if len(goodMesh.SingularVertices()) != 0 {
			t.Error("good mesh should have no singular vertices")
		}
	}
	t.Run("Basic", func(t *testing.T) {
		runTransform(t, nil)
	})
	t.Run("Rotated", func(t *testing.T) {
		runTransform(t, Rotation(XYZ(1.0, 2.0, 3.0).Normalize(), 123.0))
	})
	t.Run("Random", func(t *testing.T) {
		rand.Seed(1337)
		dc := &DualContouring{
			S: SolidSurfaceEstimator{
				Solid: randomSolid{},
			},
			Delta: 0.04,
			// Both of these are necessary to guarantee
			// manifold meshes in the general case.
			Repair: true,
			Clip:   true,
		}
		mesh := dc.Mesh()
		if mesh.NeedsRepair() {
			t.Error("mesh has singular edges")
		}
		if len(mesh.SingularVertices()) > 0 {
			t.Error("mesh has singular vertices")
		}
	})
}

func BenchmarkDualContouring(b *testing.B) {
	runBench := func(b *testing.B, gos int, repair bool) {
		solid := &CylinderSolid{
			P1:     XYZ(1, 2, 3),
			P2:     XYZ(3, 1, 4),
			Radius: 0.5,
		}
		dc := &DualContouring{
			S:      SolidSurfaceEstimator{Solid: solid},
			Delta:  0.025,
			MaxGos: gos,
			Repair: repair,
		}
		for i := 0; i < b.N; i++ {
			dc.Mesh()
		}
	}
	runRepair := func(b *testing.B, repair bool) {
		b.Run("MaxGos1", func(b *testing.B) {
			runBench(b, 1, repair)
		})
		b.Run("MaxGos0", func(b *testing.B) {
			runBench(b, 0, repair)
		})
	}
	b.Run("NoRepair", func(b *testing.B) {
		runRepair(b, false)
	})
	b.Run("Repair", func(b *testing.B) {
		runRepair(b, true)
	})
}

func TestDcCubeLayout(t *testing.T) {
	layout := newDcCubeLayout(XYZ(-1, -1, -1), XYZ(1, 1, 1), 0.04, false, 5000)
	for layout.Remaining() > 0 {
		for cubeIdx := range layout.Cubes {
			coord := layout.Corner(layout.CubeCorners(dcCubeIdx(cubeIdx))[0]).Coord
			offset := []Coord3D{
				XYZ(0, 0, 0),
				XYZ(1, 0, 0),
				XYZ(0, 1, 0),
				XYZ(1, 1, 0),
				XYZ(0, 0, 1),
				XYZ(1, 0, 1),
				XYZ(0, 1, 1),
				XYZ(1, 1, 1),
			}
			for i, off := range offset {
				expected := coord.Add(off.Scale(0.04))
				actual := layout.Corner(layout.CubeCorners(dcCubeIdx(cubeIdx))[i]).Coord
				if expected.Dist(actual) > 1e-5 {
					t.Fatalf("cube %d has inconsistent corners", cubeIdx)
				}
			}
		}

		for cubeIdx := range layout.Cubes {
			counts := map[dcCornerIdx]int{}
			for _, edge := range layout.CubeEdges(dcCubeIdx(cubeIdx)) {
				for _, corner := range layout.EdgeCorners(edge) {
					counts[corner]++
				}
			}
			for _, corner := range layout.CubeCorners(dcCubeIdx(cubeIdx)) {
				if counts[corner] != 3 {
					t.Fatalf("invalid count of corners from edges: %d", counts[corner])
					break
				}
			}
		}

		for edgeIdx := range layout.Edges {
			edge := dcEdgeIdx(edgeIdx)
			corners := layout.EdgeCorners(edge)
			c1 := layout.Corner(corners[0]).Coord
			c2 := layout.Corner(corners[1]).Coord
			diff := c2.Sub(c1).Normalize()
			if math.Abs(diff.Sum()-1) > 1e-5 {
				t.Fatalf("edge spans non-line %v to %v (offset=%d)", c1, c2, layout.ZOffset)
			}
		}

		for edgeIdx := range layout.Edges {
			edge := dcEdgeIdx(edgeIdx)
			for _, cube := range layout.EdgeCubes(edge) {
				if cube == -1 {
					continue
				}
				var found bool
				for _, edge1 := range layout.CubeEdges(cube) {
					if edge == edge1 {
						found = true
						break
					}
				}
				if !found {
					t.Fatal("EdgeCubes inconsistent with CubeEdges")
				}
			}
		}
		for cubeIdx := range layout.Cubes {
			cube := dcCubeIdx(cubeIdx)
			for _, edge := range layout.CubeEdges(cube) {
				var found bool
				for _, cube1 := range layout.EdgeCubes(edge) {
					if cube == cube1 {
						found = true
						break
					}
				}
				if !found {
					t.Fatal("EdgeCubes inconsistent with CubeEdges")
				}
			}
		}
		layout.Shift()
	}
}
