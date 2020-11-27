package model2d

import (
	"image/color"
	"io/ioutil"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/unixpickle/essentials"
)

func TestTriangulate(t *testing.T) {
	poly := []Coord{
		{0, 0},
		{1, 2},
		{3, 3},
		{1, 5},
		{-1, 3},
		{2, 3},
		{0, 2},
	}
	contained := []Coord{
		{0.2, 1},
		{0.5, 2},
		{2.5, 3},
		{1, 4},
		{1, 4.5},
		{-0.9, 3.01},
	}
	notContained := []Coord{
		{1, 0},
		{0, 2.5},
		{1, 1},
		{2.5, 4},
	}
	for i := 0; i < len(poly); i++ {
		p1 := append(append([]Coord{}, poly[i:]...), poly[:i]...)
		p2 := append([]Coord{}, p1...)
		essentials.Reverse(p2)
		for j, p := range [][]Coord{p1, p2} {
			tris := Triangulate(p)
			checkContained := func(c Coord) bool {
				for _, t := range tris {
					if triangle2DContains(t, c) {
						return true
					}
				}
				return false
			}
			for _, c := range contained {
				if !checkContained(c) {
					t.Errorf("mismatched result for contained point %v (case %d %d)", c, i, j)
				}
			}
			for _, c := range notContained {
				if checkContained(c) {
					t.Errorf("mismatched result for uncontained point %v (case %d %d)", c, i, j)
				}
			}
		}
	}
}

func TestTriangulateMeshBasic(t *testing.T) {
	mesh := NewMeshPolar(func(theta float64) float64 {
		return math.Cos(theta) + 1.5
	}, 30)
	tris := TriangulateMesh(mesh)

	testTriangulatedEdgeCounts(t, tris, mesh)
	testTriangulatedContainment(t, tris, mesh, 1000)
}

func TestTriangulateMeshComplex(t *testing.T) {
	// Create a testing mesh with holes, etc.
	bitmap := MustReadBitmap("test_data/test_bitmap_small.png", func(c color.Color) bool {
		r, g, b, _ := c.RGBA()
		return r == 0 && g == 0 && b == 0
	})
	mesh := bitmap.Mesh().SmoothSq(30)
	if !mesh.Manifold() {
		t.Fatal("non-manifold mesh")
	}

	tris := TriangulateMesh(mesh)
	if len(tris) == 0 {
		panic("no triangles")
	}

	testTriangulatedEdgeCounts(t, tris, mesh)
	testTriangulatedContainment(t, tris, mesh, 1000)

	for _, tri := range tris {
		if !isPolygonClockwise(tri[:]) {
			t.Errorf("triangle is not clockwise: %v", tri)
			break
		}
	}
}

func TestTriangulateMeshErrorCase(t *testing.T) {
	data, err := ioutil.ReadFile("test_data/triangulate_breaker.txt")
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(strings.TrimSpace(string(data)), " ")
	mesh := NewMesh()
	for _, part := range parts {
		coordStrs := strings.Split(part, ",")
		coords := [4]float64{}
		for i, s := range coordStrs {
			coords[i], err = strconv.ParseFloat(s, 64)
			if err != nil {
				t.Fatal(err)
			}
		}
		mesh.Add(&Segment{XY(coords[0], coords[1]), XY(coords[2], coords[3])})
	}

	// The interior mesh is the one with issues.
	mesh = MeshToHierarchy(mesh)[0].Children[0].FullMesh()
	mesh.Iterate(func(s *Segment) {
		mesh.Remove(s)
		mesh.Add(&Segment{s[1], s[0]})
	})

	tris := TriangulateMesh(mesh)
	if len(tris) == 0 {
		panic("no triangles")
	}

	testTriangulatedEdgeCounts(t, tris, mesh)
	testTriangulatedContainment(t, tris, mesh, 100)
}

func TestTriangulateMeshErrorCase2(t *testing.T) {
	// This error case used to trigger a NaN when
	// computing angles between segments.
	bitmap := MustReadBitmap("test_data/test_bitmap.png", func(c color.Color) bool {
		r, g, b, _ := c.RGBA()
		return r == 0 && g == 0 && b == 0
	})
	mesh := bitmap.Mesh()
	tris := TriangulateMesh(mesh)
	testTriangulatedEdgeCounts(t, tris, mesh)
	testTriangulatedContainment(t, tris, mesh, 100)
}

func triangle2DContains(tri [3]Coord, p Coord) bool {
	v1 := tri[0].Sub(tri[1])
	v2 := tri[2].Sub(tri[1])
	mat := (&Matrix2{v1.X, v2.X, v1.Y, v2.Y}).Inverse()
	coords := mat.MulColumn(p.Sub(tri[1]))
	return coords.X >= 0 && coords.Y >= 0 && coords.X+coords.Y <= 1
}

func testTriangulatedEdgeCounts(t *testing.T, tris [][3]Coord, m *Mesh) {
	dupMesh := NewMesh()
	for _, t := range tris {
		for i := 0; i < 3; i++ {
			dupMesh.Add(&Segment{t[i], t[(i+1)%3]})
		}
	}
	broken := NewMesh()
	dupMesh.Iterate(func(s *Segment) {
		dupCount := len(dupMesh.Find(s[0], s[1]))
		if len(m.Find(s[0], s[1])) == 1 {
			// This is an exterior edge.
			if dupCount != 1 {
				t.Errorf("expected exactly 1 copy of exterior edge but got %d", dupCount)
			}
		} else {
			if dupCount != 2 {
				t.Errorf("expected exactly 2 copies of interior edge but got %d", dupCount)
				broken.Add(s)
			}
		}
	})
	m.Iterate(func(s *Segment) {
		if n := len(dupMesh.Find(s[0], s[1])); n != 1 {
			t.Errorf("exterior edge should have count 1 but got %d", n)
		}
	})
}

func testTriangulatedContainment(t *testing.T, tris [][3]Coord, m *Mesh, n int) {
	solid := NewColliderSolid(MeshToCollider(m))
	for i := 0; i < n; i++ {
		c := NewCoordRandBounds(solid.Min(), solid.Max())
		expected := solid.Contains(c)
		actual := false
		for _, t := range tris {
			if triangle2DContains(t, c) {
				actual = true
			}
		}
		if actual != expected {
			t.Fatalf("point %v: contains=%v but got %v", c, expected, actual)
		}
	}
}

func BenchmarkTriangulate7(b *testing.B) {
	poly := []Coord{
		{0, 0},
		{1, 2},
		{3, 3},
		{1, 5},
		{-1, 3},
		{2, 3},
		{0, 2},
	}
	mesh := NewMesh()
	for i := 0; i < len(poly); i++ {
		mesh.Add(&Segment{poly[(i+1)%len(poly)], poly[i]})
	}

	b.Run("Triangulate", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Triangulate(poly)
		}
	})
	b.Run("TriangulateMesh", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			TriangulateMesh(mesh)
		}
	})
}

func BenchmarkTriangulate50(b *testing.B) {
	poly := []Coord{}
	const numPoints = 50
	for i := 0; i < numPoints; i++ {
		theta := math.Pi * 2 * float64(i) / numPoints
		coord := NewCoordPolar(theta, math.Abs(math.Cos(11*theta))+0.1)
		poly = append(poly, coord)
	}
	mesh := NewMesh()
	for i := 0; i < len(poly); i++ {
		mesh.Add(&Segment{poly[(i+1)%len(poly)], poly[i]})
	}

	b.Run("Triangulate", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Triangulate(poly)
		}
	})
	b.Run("TriangulateMesh", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			TriangulateMesh(mesh)
		}
	})
}

func BenchmarkTriangulateLarge(b *testing.B) {
	bitmap := MustReadBitmap("test_data/test_bitmap.png", func(c color.Color) bool {
		r, g, b, _ := c.RGBA()
		return r == 0 && g == 0 && b == 0
	})
	mesh := bitmap.Mesh()
	b.Run("Triangulate", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			TriangulateMesh(mesh)
		}
	})
}
