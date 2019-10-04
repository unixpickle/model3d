package model3d

import (
	"math"
	"sort"

	"github.com/unixpickle/essentials"
)

// A Mesh is a collection of triangles.
//
// The triangles are uniquely identified as pointers, not
// as values. This is important for methods which
// reference existing triangles, such as Remove and
// Neighbors.
//
// Triangles in a mesh are "connected" when they contain
// exactly identical points. Thus, small rounding errors
// can cause triangles to incorrectly be disassociated
// with each other.
//
// It is not safe to access a Mesh from multiple
// Goroutines at once, even for reading.
type Mesh struct {
	triangles        map[*Triangle]bool
	vertexToTriangle map[Coord3D][]*Triangle
}

// NewMesh creates an empty mesh.
func NewMesh() *Mesh {
	return &Mesh{
		triangles: map[*Triangle]bool{},
	}
}

// NewMeshTriangles creates a mesh with the given
// collection of triangles.
func NewMeshTriangles(ts []*Triangle) *Mesh {
	m := NewMesh()
	for _, t := range ts {
		m.Add(t)
	}
	return m
}

// NewMeshPolar creates a mesh with a 3D polar function.
func NewMeshPolar(radius func(g GeoCoord) float64, stops int) *Mesh {
	res := NewMesh()
	lonStep := math.Pi * 2 / float64(stops)
	latStep := math.Pi / float64(stops)
	latFunc := func(i int) float64 {
		return -math.Pi/2 + float64(i)*latStep
	}
	lonFunc := func(i int) float64 {
		if i == stops {
			// Make rounding match up at the edges, since
			// sin(-pi) != sin(pi) in the stdlib.
			return -math.Pi
		}
		return -math.Pi + float64(i)*lonStep
	}
	for lonIdx := 0; lonIdx < stops; lonIdx++ {
		for latIdx := 0; latIdx < stops; latIdx++ {
			longitude := lonFunc(lonIdx)
			latitude := latFunc(latIdx)
			longitudeNext := lonFunc(lonIdx + 1)
			latitudeNext := latFunc(latIdx + 1)
			g := []GeoCoord{
				GeoCoord{Lat: latitude, Lon: longitude},
				GeoCoord{Lat: latitude, Lon: longitudeNext},
				GeoCoord{Lat: latitudeNext, Lon: longitudeNext},
				GeoCoord{Lat: latitudeNext, Lon: longitude},
			}
			p := make([]Coord3D, 4)
			for i, x := range g {
				p[i] = x.Coord3D().Scale(radius(x))
			}
			if latIdx == 0 {
				// p[0] and p[1] are technically equivalent,
				// but they are numerically slightly different,
				// so we must make it perfect.
				p[0] = Coord3D{X: 0, Y: -radius(GeoCoord{Lat: latitude, Lon: 0}), Z: 0}
			} else if latIdx == stops-1 {
				// p[2] and p[3] are technically equivalent,
				// but see note above.
				p[2] = Coord3D{X: 0, Y: radius(GeoCoord{Lat: latitude, Lon: 0}), Z: 0}
			}
			if latIdx != 0 {
				res.Add(&Triangle{p[0], p[1], p[2]})
			}
			if latIdx != stops-1 {
				res.Add(&Triangle{p[0], p[2], p[3]})
			}
		}
	}
	return res
}

// Add adds the triangle t to the mesh.
func (m *Mesh) Add(t *Triangle) {
	if m.vertexToTriangle == nil {
		m.triangles[t] = true
		return
	} else if m.triangles[t] {
		return
	}
	for _, p := range t {
		m.vertexToTriangle[p] = append(m.vertexToTriangle[p], t)
	}
	m.triangles[t] = true
}

// AddMesh adds all the triangles from m1 to m.
func (m *Mesh) AddMesh(m1 *Mesh) {
	m1.Iterate(m.Add)
}

// Remove removes the triangle t from the mesh.
//
// It looks at t as a pointer, so the pointer must be
// exactly the same as a triangle passed to Add.
func (m *Mesh) Remove(t *Triangle) {
	if !m.triangles[t] {
		return
	}
	delete(m.triangles, t)
	if m.vertexToTriangle == nil {
		return
	}
	for _, p := range t {
		s := m.vertexToTriangle[p]
		for i, t1 := range s {
			if t1 == t {
				essentials.UnorderedDelete(&s, i)
				break
			}
		}
		m.vertexToTriangle[p] = s
	}
}

// Iterate calls f for every triangle in m in an arbitrary
// order.
//
// If f adds or removes triangles, they will not be
// visited.
func (m *Mesh) Iterate(f func(t *Triangle)) {
	m.IterateSorted(f, nil)
}

// IterateSorted is like Iterate, but it first sorts all
// the triangles according to a less than function, cmp.
func (m *Mesh) IterateSorted(f func(t *Triangle), cmp func(t1, t2 *Triangle) bool) {
	all := m.TriangleSlice()
	if cmp != nil {
		sort.Slice(all, func(i, j int) bool {
			return cmp(all[i], all[j])
		})
	}
	for _, t := range all {
		if m.triangles[t] {
			f(t)
		}
	}
}

// Neighbors gets all the triangles with a side touching a
// given triangle t.
//
// The triangle t itself is not included in the results.
//
// The triangle t needn't be in the mesh. However, if it
// is not in the mesh, but an equivalent triangle is, then
// said equivalent triangle will be in the results.
func (m *Mesh) Neighbors(t *Triangle) []*Triangle {
	resSet := map[*Triangle]int{}
	for _, p := range t {
		for _, t1 := range m.getVertexToTriangle()[p] {
			if t1 != t {
				resSet[t1]++
			}
		}
	}
	res := make([]*Triangle, 0, len(resSet))
	for t1, count := range resSet {
		if count > 1 {
			res = append(res, t1)
		}
	}
	return res
}

// Find gets all the triangles that contain all of the
// passed points.
//
// For example, to find all triangles containing a line
// from p1 to p2, you could do m.Find(p1, p2).
func (m *Mesh) Find(ps ...Coord3D) []*Triangle {
	resSet := map[*Triangle]int{}
	for _, p := range ps {
		for _, t1 := range m.getVertexToTriangle()[p] {
			resSet[t1]++
		}
	}
	res := make([]*Triangle, 0, len(resSet))
	for t1, count := range resSet {
		if count == len(ps) {
			res = append(res, t1)
		}
	}
	return res
}

// MapCoords creates a new mesh by transforming all of the
// coordinates according to the function f.
func (m *Mesh) MapCoords(f func(Coord3D) Coord3D) *Mesh {
	mapping := map[Coord3D]Coord3D{}
	if m.vertexToTriangle != nil {
		for c := range m.vertexToTriangle {
			mapping[c] = f(c)
		}
	} else {
		for t := range m.triangles {
			for _, c := range t {
				if _, ok := mapping[c]; !ok {
					mapping[c] = f(c)
				}
			}
		}
	}
	m1 := NewMesh()
	m.Iterate(func(t *Triangle) {
		t1 := *t
		for i, p := range t {
			t1[i] = mapping[p]
		}
		m1.Add(&t1)
	})
	return m1
}

// EncodeSTL encodes the mesh as STL data.
func (m *Mesh) EncodeSTL() []byte {
	return EncodeSTL(m.TriangleSlice())
}

// EncodePLY encodes the mesh as a PLY file with color.
func (m *Mesh) EncodePLY(colorFunc func(c Coord3D) [3]uint8) []byte {
	return EncodePLY(m.TriangleSlice(), colorFunc)
}

// EncodeMaterialOBJ encodes the mesh as a zip file with
// per-triangle material.
func (m *Mesh) EncodeMaterialOBJ(colorFunc func(t *Triangle) [3]float64) []byte {
	return EncodeMaterialOBJ(m.TriangleSlice(), colorFunc)
}

// TriangleSlice gets a snapshot of all the triangles
// currently in the mesh. The resulting slice is a copy,
// and will not change as the mesh is updated.
func (m *Mesh) TriangleSlice() []*Triangle {
	ts := make([]*Triangle, 0, len(m.triangles))
	for t := range m.triangles {
		ts = append(ts, t)
	}
	return ts
}

// Min gets the component-wise minimum across all the
// vertices in the mesh.
func (m *Mesh) Min() Coord3D {
	if len(m.triangles) == 0 {
		return Coord3D{}
	}
	var result Coord3D
	var firstFlag bool
	for t := range m.triangles {
		for _, c := range t {
			if !firstFlag {
				result = c
				firstFlag = true
			} else {
				result = result.Min(c)
			}
		}
	}
	return result
}

// Max gets the component-wise maximum across all the
// vertices in the mesh.
func (m *Mesh) Max() Coord3D {
	if len(m.triangles) == 0 {
		return Coord3D{}
	}
	var result Coord3D
	var firstFlag bool
	for t := range m.triangles {
		for _, c := range t {
			if !firstFlag {
				result = c
				firstFlag = true
			} else {
				result = result.Max(c)
			}
		}
	}
	return result
}

func (m *Mesh) getVertexToTriangle() map[Coord3D][]*Triangle {
	if m.vertexToTriangle != nil {
		return m.vertexToTriangle
	}
	m.vertexToTriangle = map[Coord3D][]*Triangle{}
	for t := range m.triangles {
		for _, p := range t {
			m.vertexToTriangle[p] = append(m.vertexToTriangle[p], t)
		}
	}
	return m.vertexToTriangle
}
