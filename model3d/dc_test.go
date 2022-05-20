package model3d

import (
	"testing"
)

func TestDcCubeLayout(t *testing.T) {
	layout := newDcCubeLayout(XYZ(-3.0, -2.0, -5.0), XYZ(4.0, 3.0, 1.0), 1.0)

	cubes, _ := layout.FirstRow()
	prevCubes := append([]*dcCube{}, cubes...)
	for i := 0; i < 2; i++ {
		prevCubes, _ = layout.NextRow(prevCubes, i)
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
			for j, edge := range cube.Edges {
				var found bool
				for _, c1 := range edge.Cubes {
					if c1 == cube {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("cube %d (row %d) edge %d not connected to cube "+
						"(cube=%p, cubes=%v, coord=%v)",
						i, cubeRow, j, cube, edge.Cubes, cube.Corners[0].Coord)
				}
			}
		}
	})

	t.Run("EdgeCubesHaveCorners", func(t *testing.T) {
		for i, cube := range cubes {
			cubeRow := i / ((len(layout.Xs) - 1) * (len(layout.Ys) - 1))
			for j, edge := range cube.Edges {
				for k, c1 := range edge.Cubes {
					if c1 == nil {
						continue
					}
					var found1, found2 bool
					for _, corner := range c1.Corners {
						if corner == edge.Corners[0] {
							found1 = true
						} else if corner == edge.Corners[1] {
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
		c2c := map[Coord3D]*dcCorner{}
		for _, cube := range cubes {
			for _, corner := range cube.Corners {
				if x, ok := c2c[corner.Coord]; ok {
					if x != corner {
						t.Fatalf("duplicated coordinate %v", corner.Coord)
					}
				} else {
					c2c[corner.Coord] = corner
				}
			}
		}
	})

	t.Run("UniqueEdges", func(t *testing.T) {
		c2c := map[Segment]*dcEdge{}
		for _, cube := range cubes {
			for _, edge := range cube.Edges {
				k := NewSegment(edge.Corners[0].Coord, edge.Corners[1].Coord)
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
			corner := cube.Corners[0].Coord
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
				corners := cube.Edges[j].Corners
				a := NewSegment(corners[0].Coord, corners[1].Coord)
				if a != x {
					t.Fatalf("cube %d edge %d: expected %v but got %v", i, j, x, a)
				}
			}
		}
	})
}

func TestDualContouring(t *testing.T) {
	t.Run("Sphere", func(t *testing.T) {
		solid := &Sphere{Radius: 1.0}
		dc := &DualContouring{Delta: 0.04}
		mesh := dc.Mesh(solid)
		// In general, meshes from Dual Contouring might not be
		// manifold, but this one should be.
		MustValidateMesh(t, mesh, true)
	})
}
