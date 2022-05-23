package model3d

import (
	"math"
	"testing"
)

func TestDualContouring(t *testing.T) {
	runTests := func(t *testing.T, gos int) {
		t.Run("Sphere", func(t *testing.T) {
			solid := &Sphere{Radius: 1.0}
			dc := &DualContouring{S: SolidSurfaceEstimator{Solid: solid}, Delta: 0.04, MaxGos: gos}
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
			dc := &DualContouring{S: SolidSurfaceEstimator{Solid: solid}, Delta: 0.04, MaxGos: gos}
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
	t.Run("MaxGos1", func(t *testing.T) {
		runTests(t, 1)
	})
	t.Run("MaxGos8", func(t *testing.T) {
		runTests(t, 8)
	})
	t.Run("MaxGos0", func(t *testing.T) {
		runTests(t, 0)
	})
}

func BenchmarkDualContouring(b *testing.B) {
	runBench := func(b *testing.B, gos int) {
		solid := &CylinderSolid{
			P1:     XYZ(1, 2, 3),
			P2:     XYZ(3, 1, 4),
			Radius: 0.5,
		}
		dc := &DualContouring{S: SolidSurfaceEstimator{Solid: solid}, Delta: 0.025, MaxGos: gos}
		for i := 0; i < b.N; i++ {
			dc.Mesh()
		}
	}
	b.Run("MaxGos1", func(b *testing.B) {
		runBench(b, 1)
	})
	b.Run("MaxGos0", func(b *testing.B) {
		runBench(b, 0)
	})
}

func TestDcCubeLayout(t *testing.T) {
	layout := newDcCubeLayout(XYZ(-3.0, -2.0, -5.0), XYZ(4.0, 3.0, 1.0), 1.0, true)
	heap := &dcHeap{}

	cubes, _ := layout.FirstRow(heap)
	prevCubes := append([]dcCubeIdx{}, cubes...)
	for i := 0; i < 2; i++ {
		prevCubes, _ = layout.NextRow(heap, prevCubes, i)
		cubes = append(cubes, prevCubes...)
	}

	// We should now have three rows of cubes.
	expectedLen := (len(layout.Xs) - 1) * (len(layout.Ys) - 1) * 3
	if len(cubes) != expectedLen {
		t.Fatalf("expected %d cubes but got %d", expectedLen, len(cubes))
	}

	t.Run("EdgeRefsCube", func(t *testing.T) {
		for i, cube := range cubes {
			cubeRow := i / ((len(layout.Xs) - 1) * (len(layout.Ys) - 1))
			for j, edge := range heap.Cube(cube).Edges {
				var found bool
				for _, c1 := range heap.Edge(edge).Cubes {
					if c1 == cube {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("cube %d (row %d) edge %d not connected to cube "+
						"(cube=%d, cubes=%v, coord=%v)",
						i, cubeRow, j, cube, heap.Edge(edge).Cubes,
						heap.Corner(heap.Cube(cube).Corners[0]).Coord)
				}
			}
		}
	})

	t.Run("EdgeCubesHaveCorners", func(t *testing.T) {
		for i, cube := range cubes {
			cubeRow := i / ((len(layout.Xs) - 1) * (len(layout.Ys) - 1))
			for j, edge := range heap.Cube(cube).Edges {
				for k, c1 := range heap.Edge(edge).Cubes {
					if c1 < 0 {
						continue
					}
					var found1, found2 bool
					for _, corner := range heap.Cube(c1).Corners {
						if corner == heap.Edge(edge).Corners[0] {
							found1 = true
						} else if corner == heap.Edge(edge).Corners[1] {
							found2 = true
						}
					}
					if !found1 || !found2 {
						t.Fatalf("cube %d (row %d) edge %d cube %d does not have corners (%v, %v)",
							i, cubeRow, j, k, found1, found2)
					}
				}
			}
		}
	})

	t.Run("UniqueCorners", func(t *testing.T) {
		c2c := map[Coord3D]dcCornerIdx{}
		for _, cube := range cubes {
			for _, corner := range heap.Cube(cube).Corners {
				if x, ok := c2c[heap.Corner(corner).Coord]; ok {
					if x != corner {
						t.Fatalf("duplicated coordinate %v", heap.Corner(corner).Coord)
					}
				} else {
					c2c[heap.Corner(corner).Coord] = corner
				}
			}
		}
	})

	t.Run("UniqueEdges", func(t *testing.T) {
		c2c := map[Segment]dcEdgeIdx{}
		for _, cube := range cubes {
			for _, edge := range heap.Cube(cube).Edges {
				k := NewSegment(
					heap.Corner(heap.Edge(edge).Corners[0]).Coord,
					heap.Corner(heap.Edge(edge).Corners[1]).Coord,
				)
				if x, ok := c2c[k]; ok {
					if x != edge {
						t.Fatalf("duplicated segment %v", k)
					}
				} else {
					c2c[k] = edge
				}
			}
		}
	})

	t.Run("CorrectEdges", func(t *testing.T) {
		for i, cube := range cubes {
			corner := heap.Corner(heap.Cube(cube).Corners[0]).Coord
			bottom := corner.Add(Z(1))
			edges := [12]Segment{
				// Top edges.
				NewSegment(corner, corner.Add(X(1))),
				NewSegment(corner, corner.Add(Y(1))),
				NewSegment(corner.Add(X(1)), corner.Add(XY(1, 1))),
				NewSegment(corner.Add(Y(1)), corner.Add(XY(1, 1))),
				// Vertical edges.
				NewSegment(corner, bottom),
				NewSegment(corner.Add(X(1)), bottom.Add(X(1))),
				NewSegment(corner.Add(Y(1)), bottom.Add(Y(1))),
				NewSegment(corner.Add(XY(1, 1)), bottom.Add(XY(1, 1))),
				// Bottom edges.
				NewSegment(bottom, bottom.Add(X(1))),
				NewSegment(bottom, bottom.Add(Y(1))),
				NewSegment(bottom.Add(X(1)), bottom.Add(XY(1, 1))),
				NewSegment(bottom.Add(Y(1)), bottom.Add(XY(1, 1))),
			}
			for j, x := range edges {
				corners := heap.Edge(heap.Cube(cube).Edges[j]).Corners
				a := NewSegment(heap.Corner(corners[0]).Coord, heap.Corner(corners[1]).Coord)
				if a != x {
					t.Fatalf("cube %d edge %d: expected %v but got %v", i, j, x, a)
				}
			}
		}
	})

	t.Run("NoExtraAllocs", func(t *testing.T) {
		if len(cubes) != len(heap.cubes) {
			t.Errorf("got %d cubes but allocated %d", len(cubes), len(heap.cubes))
		}
		foundEdges := map[dcEdgeIdx]bool{}
		for _, c := range cubes {
			for _, edge := range heap.Cube(c).Edges {
				foundEdges[edge] = true
			}
		}
		for i := range heap.edges {
			if !foundEdges[dcEdgeIdx(i)] {
				t.Error("superfluous edges detected")
				break
			}
		}
	})

	// NOTE: this test should come at the bottom, since
	// it modifies the heap.
	t.Run("DeleteRefCount", func(t *testing.T) {
		checkRefCount := func() {
			counts := map[dcCornerIdx]int{}
			freeCubes := map[int]bool{}
			for _, x := range heap.freeCubes {
				freeCubes[x] = true
			}
			for i, c := range heap.cubes {
				if freeCubes[i] {
					continue
				}
				for _, corner := range c.Corners {
					counts[corner]++
				}
			}
			for idx, expected := range counts {
				if actual := heap.Corner(idx).Refs; actual != expected {
					t.Fatalf("corner %d has ref count %d but is referenced %d times",
						idx, actual, expected)
				}
			}
		}
		checkRefCount()
		// Try removing some corners.
		if actual := len(heap.freeCorners); actual != 0 {
			t.Errorf("original free corners should be 0 but got %d", actual)
		}
		heap.UnlinkCube(0)
		heap.UnlinkCube(1)
		heap.UnlinkCube(dcCubeIdx(len(heap.cubes) / 2))
		heap.UnlinkCube(dcCubeIdx(len(heap.cubes) - 15))
		if actual := len(heap.freeCorners); actual == 0 {
			t.Errorf("final free corners should be > 0 but got %d", actual)
		}
		if actual := len(heap.freeEdges); actual == 0 {
			t.Errorf("final free edges should be > 0 but got %d", actual)
		}
		checkRefCount()
	})
}
