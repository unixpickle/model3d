package model3d

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/unixpickle/model3d/model2d"
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

func TestDualContouringInterior(t *testing.T) {
	solid := &Sphere{Radius: 1.0}
	dc := &DualContouring{
		S:          SolidSurfaceEstimator{Solid: solid},
		Delta:      0.04,
		MaxGos:     8,
		BufferSize: 5000,
	}
	var mesh *Mesh
	var interior []Coord3D
	mesh, interior = dc.MeshInterior()
	if len(interior) == 0 {
		t.Fatal("no interior points")
	}
	for _, c := range interior {
		if !solid.Contains(c) {
			t.Fatalf("interior point not contained: %v", c)
		}
	}
	// In general, meshes from Dual Contouring might not be
	// manifold, but this one should be.
	MustValidateMesh(t, mesh, false)
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
		if !goodMesh.Orientable() || len(goodMesh.InconsistentEdges()) > 0 {
			t.Error("mesh is incorrectly oriented")
		}
	}
	t.Run("Basic", func(t *testing.T) {
		runTransform(t, nil)
	})
	t.Run("Rotated", func(t *testing.T) {
		runTransform(t, Rotation(XYZ(1.0, 2.0, 3.0).Normalize(), 123.0))
	})
	t.Run("Random", func(t *testing.T) {
		rng := rand.New(rand.NewSource(1337))
		dc := &DualContouring{
			S: SolidSurfaceEstimator{
				Solid: randomSolid{rng: rng},
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
		if !mesh.Orientable() || len(mesh.InconsistentEdges()) > 0 {
			t.Error("mesh is incorrectly oriented")
		}
	})
}

func TestDualContouringDuplicateVertices(t *testing.T) {
	blanketTassel := func() Solid {
		outsetPart := CheckedFuncSolid(
			XYZ(-5, 0, -3),
			XYZ(5, 5, 3),
			func(c Coord3D) bool {
				r := c.Mul(XZ(1, 2.0)).Norm()
				actualR := 1.2 * math.Sqrt(c.Y) / math.Sqrt(5)
				return r < actualR
			},
		)
		connector := &Sphere{Radius: 1}
		return JoinedSolid{outsetPart, connector}
	}
	solid := func() Solid {
		joined := JoinedSolid{
			TranslateSolid(
				RotateSolid(
					blanketTassel(),
					Z(1),
					-math.Pi/4,
				),
				XY(10, 10),
			),
			TranslateSolid(
				RotateSolid(
					blanketTassel(),
					Z(1),
					math.Pi/4,
				),
				XY(-10, 10),
			),
			TranslateSolid(
				RotateSolid(
					blanketTassel(),
					Z(1),
					math.Pi+math.Pi/4,
				),
				XY(10, -10),
			),
			TranslateSolid(
				RotateSolid(
					blanketTassel(),
					Z(1),
					math.Pi-math.Pi/4,
				),
				XY(-10, -10),
			),
		}
		baseSolid := joined.Optimize()
		distorted := CheckedFuncSolid(
			baseSolid.Min(),
			baseSolid.Max().Add(Z(3)),
			func(c Coord3D) bool {
				diagDot := c.XY().Dot(model2d.XY(0.3, 0.5)) - 7
				displaceScale := 0.5 * (c.XY().Dot(model2d.XY(-0.5, 0.7)) + 30) / 35
				displacement := math.Exp(-math.Pow(diagDot, 2))
				c.Z -= displacement * displaceScale

				diagDot = c.XY().Dot(model2d.XY(0.5, 0.2)) + 7
				displaceScale = (c.XY().Dot(model2d.XY(0.5, -0.7)) + 30) / 35
				displacement = math.Exp(-math.Pow(diagDot, 2))
				c.Z -= displacement * displaceScale

				return baseSolid.Contains(c)
			},
		)

		coarseMesh := MarchingCubesSearch(distorted, 0.1, 4)
		coarseCollider := MeshToCollider(coarseMesh)
		bottomFillIn := CheckedFuncSolid(
			coarseMesh.Min().Sub(Z(0.1)),
			coarseMesh.Max(),
			func(c Coord3D) bool {
				ray := &Ray{Origin: XYZ(c.X, c.Y, -10), Direction: Z(1)}
				rc, ok := coarseCollider.FirstRayCollision(ray)
				return ok && ray.Origin.Add(ray.Direction.Scale(rc.Scale)).Z >= c.Z-0.1
			},
		)

		outlineSolid := model2d.CheckedFuncSolid(
			coarseCollider.Min().XY(),
			coarseCollider.Max().XY(),
			func(c model2d.Coord) bool {
				ray := &Ray{Origin: XYZ(c.X, c.Y, -10), Direction: Z(1)}
				_, ok := coarseCollider.FirstRayCollision(ray)
				return ok
			},
		)
		outlineMesh := model2d.MarchingSquaresSearch(outlineSolid, 0.1, 8)
		outset := model2d.NewColliderSolidInset(model2d.MeshToCollider(outlineMesh), -1)
		outlineProfile := ProfileSolid(outset, bottomFillIn.Min().Z, bottomFillIn.Min().Z+1)

		fullSolid := JoinedSolid{distorted, bottomFillIn, outlineProfile}

		return ScaleSolid(fullSolid, 4)
	}
	mesh, _ := DualContourInterior(solid(), 0.5, true, false)
	if mesh.NeedsRepair() {
		t.Error("mesh has singular edges")
	}
	if len(mesh.SingularVertices()) > 0 {
		t.Error("mesh has singular vertices")
	}
	if !mesh.Orientable() || len(mesh.InconsistentEdges()) > 0 {
		t.Error("mesh is incorrectly oriented")
	}
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
