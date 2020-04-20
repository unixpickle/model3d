package model3d

import (
	"bufio"
	"math"
	"os"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
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
// A Mesh can be read safely from concurrent Goroutines,
// but modifications must not be performed concurrently
// with any mesh operations.
type Mesh struct {
	triangles map[*Triangle]bool

	// Stores a map[Coord3D][]*Triangle
	vertexToTriangle atomic.Value
	v2tCreateLock    sync.Mutex
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

// NewMeshRect creates a new mesh around the rectangular
// bounds.
func NewMeshRect(min, max Coord3D) *Mesh {
	mesh := NewMesh()

	point := func(x, y, z int) Coord3D {
		res := min
		if x == 1 {
			res.X = max.X
		}
		if y == 1 {
			res.Y = max.Y
		}
		if z == 1 {
			res.Z = max.Z
		}
		return res
	}

	addFace := func(p1, p2, p3, p4 Coord3D) {
		mesh.Add(&Triangle{p1, p3, p2})
		mesh.Add(&Triangle{p2, p3, p4})
	}

	// Front and back faces.
	addFace(min, point(0, 0, 1), point(1, 0, 0), point(1, 0, 1))
	addFace(max, point(0, 1, 1), point(1, 1, 0), point(0, 1, 0))

	// Left and right faces.
	addFace(min, point(0, 1, 0), point(0, 0, 1), point(0, 1, 1))
	addFace(max, point(1, 1, 0), point(1, 0, 1), point(1, 0, 0))

	// Top and bottom faces.
	addFace(min, point(1, 0, 0), point(0, 1, 0), point(1, 1, 0))
	addFace(max, point(1, 0, 1), point(0, 1, 1), point(0, 0, 1))

	return mesh
}

// Add adds the triangle t to the mesh.
func (m *Mesh) Add(t *Triangle) {
	v2t := m.getVertexToTriangleOrNil()
	if v2t == nil {
		m.triangles[t] = true
		return
	} else if m.triangles[t] {
		return
	}

	for _, p := range t {
		v2t[p] = append(v2t[p], t)
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
	v2t := m.getVertexToTriangleOrNil()
	if v2t != nil {
		for _, p := range t {
			m.removeTriangleFromVertex(v2t, t, p)
		}
	}
}

func (m *Mesh) removeTriangleFromVertex(v2t map[Coord3D][]*Triangle, t *Triangle, p Coord3D) {
	s := v2t[p]
	for i, t1 := range s {
		if t1 == t {
			essentials.UnorderedDelete(&s, i)
			break
		}
	}
	v2t[p] = s
}

// Contains checks if t has been added to the mesh.
func (m *Mesh) Contains(t *Triangle) bool {
	_, ok := m.triangles[t]
	return ok
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

// IterateVertices calls f for every vertex in m in an
// arbitrary order.
//
// If f adds or removes vertices, they will not be
// visited.
func (m *Mesh) IterateVertices(f func(c Coord3D)) {
	v2t := m.getVertexToTriangle()
	for _, c := range m.VertexSlice() {
		if _, ok := v2t[c]; ok {
			f(c)
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
	counts := m.neighborsWithCounts(t)
	res := make([]*Triangle, 0, len(counts))
	for t1, count := range counts {
		if count > 1 {
			res = append(res, t1)
		}
	}
	return res
}

func (m *Mesh) neighborsWithCounts(t *Triangle) map[*Triangle]int {
	counts := map[*Triangle]int{}
	for _, p := range t {
		for _, t1 := range m.getVertexToTriangle()[p] {
			if t1 != t {
				counts[t1]++
			}
		}
	}
	return counts
}

// Find gets all the triangles that contain all of the
// passed points.
//
// For example, to find all triangles containing a line
// from p1 to p2, you could do m.Find(p1, p2).
func (m *Mesh) Find(ps ...Coord3D) []*Triangle {
	if len(ps) == 1 {
		return append([]*Triangle{}, m.getVertexToTriangle()[ps[0]]...)
	}

	tris := m.getVertexToTriangle()[ps[0]]
	res := make([]*Triangle, 0, len(tris))

TriLoop:
	for _, t := range tris {
		for _, p := range ps[1:] {
			if p != t[0] && p != t[1] && p != t[2] {
				continue TriLoop
			}
		}
		res = append(res, t)
	}

	return res
}

// Scale creates a new mesh by scaling the coordinates by
// a factor s.
func (m *Mesh) Scale(s float64) *Mesh {
	return m.MapCoords(Coord3D{X: s, Y: s, Z: s}.Mul)
}

// MapCoords creates a new mesh by transforming all of the
// coordinates according to the function f.
func (m *Mesh) MapCoords(f func(Coord3D) Coord3D) *Mesh {
	mapping := map[Coord3D]Coord3D{}
	if v2t := m.getVertexToTriangleOrNil(); v2t != nil {
		for c := range v2t {
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

// Transform applies t to the coordinates.
func (m *Mesh) Transform(t Transform) *Mesh {
	return m.MapCoords(t.Apply)
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

// SaveGroupedSTL writes the mesh to an STL file with the
// triangles grouped in such a way that the file can be
// compressed efficiently.
func (m *Mesh) SaveGroupedSTL(path string) error {
	w, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "save grouped STL")
	}
	defer w.Close()

	bufWriter := bufio.NewWriter(w)

	tris := m.TriangleSlice()
	GroupTriangles(tris)
	if err := WriteSTL(bufWriter, tris); err != nil {
		return errors.Wrap(err, "save grouped STL")
	}
	if err := bufWriter.Flush(); err != nil {
		return errors.Wrap(err, "save grouped STL")
	}
	return nil
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

// VertexSlice gets a snapshot of all the vertices
// currently in the mesh.
//
// The result is a copy and is in no way connected to the
// mesh in memory.
func (m *Mesh) VertexSlice() []Coord3D {
	v2t := m.getVertexToTriangle()
	vertices := make([]Coord3D, 0, len(v2t))
	for v := range v2t {
		vertices = append(vertices, v)
	}
	return vertices
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
	v2t := m.getVertexToTriangleOrNil()
	if v2t != nil {
		return v2t
	}

	// Use a lock to ensure two different maps aren't
	// created and returned on different Goroutines.
	m.v2tCreateLock.Lock()
	defer m.v2tCreateLock.Unlock()

	// Another goroutine could have created a map while we
	// waited on the lock.
	v2t = m.getVertexToTriangleOrNil()
	if v2t != nil {
		return v2t
	}

	v2t = map[Coord3D][]*Triangle{}
	for t := range m.triangles {
		for _, p := range t {
			v2t[p] = append(v2t[p], t)
		}
	}
	m.vertexToTriangle.Store(v2t)

	return v2t
}

func (m *Mesh) getVertexToTriangleOrNil() map[Coord3D][]*Triangle {
	res := m.vertexToTriangle.Load()
	if res == nil {
		return nil
	}
	return res.(map[Coord3D][]*Triangle)
}
