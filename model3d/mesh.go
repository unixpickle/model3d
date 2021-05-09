// Generated from templates/mesh.template

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
	"github.com/unixpickle/model3d/model2d"
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
	faces map[*Triangle]bool

	// Stores a *CoordToFaces
	vertexToFace  atomic.Value
	v2fCreateLock sync.Mutex
}

// NewMesh creates an empty mesh.
func NewMesh() *Mesh {
	return &Mesh{
		faces: map[*Triangle]bool{},
	}
}

// NewMeshTriangles creates a mesh with the given
// collection of triangles.
func NewMeshTriangles(faces []*Triangle) *Mesh {
	m := NewMesh()
	for _, f := range faces {
		m.Add(f)
	}
	return m
}

// NewMeshPolar creates a mesh with a 3D polar function.
//
// If radius is nil, a radius of 1 is used.
func NewMeshPolar(radius func(g GeoCoord) float64, stops int) *Mesh {
	if radius == nil {
		radius = func(g GeoCoord) float64 {
			return 1
		}
	}
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
				{Lat: latitude, Lon: longitude},
				{Lat: latitude, Lon: longitudeNext},
				{Lat: latitudeNext, Lon: longitudeNext},
				{Lat: latitudeNext, Lon: longitude},
			}
			p := make([]Coord3D, 4)
			for i, x := range g {
				p[i] = x.Coord3D().Scale(radius(x))
			}
			if latIdx == 0 {
				// p[0] and p[1] are technically equivalent,
				// but they are numerically slightly different,
				// so we must make it perfect.
				p[0] = XYZ(0, -radius(GeoCoord{Lat: latitude, Lon: 0}), 0)
			} else if latIdx == stops-1 {
				// p[2] and p[3] are technically equivalent,
				// but see note above.
				p[2] = XYZ(0, radius(GeoCoord{Lat: latitude, Lon: 0}), 0)
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

	// Front and back faces.
	mesh.AddQuad(min, point(1, 0, 0), point(1, 0, 1), point(0, 0, 1))
	mesh.AddQuad(max, point(1, 1, 0), point(0, 1, 0), point(0, 1, 1))

	// Left and right faces.
	mesh.AddQuad(min, point(0, 0, 1), point(0, 1, 1), point(0, 1, 0))
	mesh.AddQuad(max, point(1, 0, 1), point(1, 0, 0), point(1, 1, 0))

	// Top and bottom faces.
	mesh.AddQuad(min, point(0, 1, 0), point(1, 1, 0), point(1, 0, 0))
	mesh.AddQuad(max, point(0, 1, 1), point(0, 0, 1), point(1, 0, 1))

	return mesh
}

// NewMeshCylinder creates a 3D mesh representing a
// cylinder.
//
// This is done by discretizing the side circle into
// numStops segments and attaching triangles between them.
func NewMeshCylinder(p1, p2 Coord3D, radius float64, numStops int) *Mesh {
	z := p2.Sub(p1).Normalize()
	x, y := z.OrthoBasis()

	sidePoint := func(side Coord3D, i int) Coord3D {
		i = i % numStops
		theta := float64(i) * math.Pi * 2 / float64(numStops)
		offset := x.Scale(math.Cos(theta)).Add(y.Scale(math.Sin(theta))).Scale(radius)
		return side.Add(offset)
	}

	res := NewMesh()
	for i := 0; i < numStops; i++ {
		p11 := sidePoint(p1, i)
		p12 := sidePoint(p1, i+1)
		p21 := sidePoint(p2, i)
		p22 := sidePoint(p2, i+1)

		// Side faces
		res.Add(&Triangle{p11, p12, p1})
		res.Add(&Triangle{p21, p2, p22})

		// Connection along shaft
		res.AddQuad(p12, p11, p21, p22)
	}

	return res
}

// NewMeshTorus creates a 3D mesh representing a torus.
//
// This is done by discretizing the local ring into
// innerStops segments, and the larger ring into
// outerStops segments.
//
// The innerRadius must be smaller than the outerRadius,
// since doing otherwise would cause self-intersections.
func NewMeshTorus(center, axis Coord3D, innerRadius, outerRadius float64, innerStops,
	outerStops int) *Mesh {
	if innerRadius > outerRadius {
		panic("resulting mesh would have self-intersections, since innerRadius > outerRadius")
	}

	axis = axis.Normalize()
	x, y := axis.OrthoBasis()

	ringPoint := func(outerIndex, innerIndex int) Coord3D {
		outerTheta := float64(outerIndex%outerStops) * math.Pi * 2 / float64(outerStops)

		ringX := x.Scale(math.Cos(outerTheta)).Add(y.Scale(math.Sin(outerTheta)))
		ringY := axis

		theta := float64(innerIndex%innerStops) * math.Pi * 2 / float64(innerStops)
		offset := ringX.Scale(math.Cos(theta)).Add(ringY.Scale(math.Sin(theta))).Scale(innerRadius)
		return offset.Add(ringX.Scale(outerRadius)).Add(center)
	}

	res := NewMesh()
	for i := 0; i < outerStops; i++ {
		for j := 0; j < innerStops; j++ {
			p11 := ringPoint(i, j)
			p12 := ringPoint(i, j+1)
			p21 := ringPoint(i+1, j+1)
			p22 := ringPoint(i+1, j)
			res.AddQuad(p11, p12, p21, p22)
		}
	}

	return res
}

// NewMeshCone creates a 3D mesh representing a cone.
//
// This is done by discretizing the base circle into
// numStops segments and attaching triangles to the tip.
func NewMeshCone(tip, base Coord3D, radius float64, numStops int) *Mesh {
	z := tip.Sub(base).Normalize()
	x, y := z.OrthoBasis()

	ringPoint := func(i int) Coord3D {
		i = i % numStops
		theta := float64(i) * math.Pi * 2 / float64(numStops)
		offset := x.Scale(math.Cos(theta)).Add(y.Scale(math.Sin(theta)))
		return base.Add(offset.Scale(radius))
	}

	res := NewMesh()
	for i := 0; i < numStops; i++ {
		p1 := ringPoint(i)
		p2 := ringPoint(i + 1)
		res.Add(&Triangle{p1, p2, base})
		res.Add(&Triangle{p2, p1, tip})
	}

	return res
}

// ProfileMesh creates a 3D mesh from a 2D mesh by using
// the 2D mesh as a face surface and extending it along
// the Z axis.
//
// The 2D mesh must be manifold, closed, and oriented.
func ProfileMesh(m2d *model2d.Mesh, minZ, maxZ float64) *Mesh {
	tris := model2d.TriangulateMesh(m2d)
	m := NewMesh()
	for _, t := range tris {
		m.Add(&Triangle{
			XYZ(t[0].X, t[0].Y, minZ),
			XYZ(t[1].X, t[1].Y, minZ),
			XYZ(t[2].X, t[2].Y, minZ),
		})
		m.Add(&Triangle{
			XYZ(t[1].X, t[1].Y, maxZ),
			XYZ(t[0].X, t[0].Y, maxZ),
			XYZ(t[2].X, t[2].Y, maxZ),
		})
	}

	// Add sides to triangle edges with no neighbors.
	m.Iterate(func(t *Triangle) {
		if t[0].Z != minZ {
			return
		}
		for i := 0; i < 3; i++ {
			seg := [2]Coord3D{t[(i+1)%3], t[i]}
			if len(m.Find(seg[0], seg[1])) == 1 {
				// This needs to be connected from minZ to maxZ.
				p3, p4 := seg[1], seg[0]
				p3.Z = maxZ
				p4.Z = maxZ
				m.AddQuad(seg[0], seg[1], p3, p4)
			}
		}
	})
	return m
}

// Add adds the triangle f to the mesh.
func (m *Mesh) Add(f *Triangle) {
	v2f := m.getVertexToFaceOrNil()
	if v2f == nil {
		m.faces[f] = true
		return
	} else if m.faces[f] {
		return
	}

	for _, p := range f {
		addFaceToVertex(v2f, p, f)
	}
	m.faces[f] = true
}

// AddQuad adds a quadrilateral to the mesh.
//
// For correct normals, the vertices should be in counter-
// clockwise order as seen from the outside of the mesh.
func (m *Mesh) AddQuad(p1, p2, p3, p4 Coord3D) [2]*Triangle {
	res := [2]*Triangle{
		{p1, p2, p4},
		{p2, p3, p4},
	}
	m.Add(res[0])
	m.Add(res[1])
	return res
}

// AddMesh adds all the triangles from m1 to m.
func (m *Mesh) AddMesh(m1 *Mesh) {
	m1.Iterate(m.Add)
}

// Remove removes the triangle f from the mesh.
//
// It looks at f as a pointer, so the pointer must be
// exactly the same as one passed to Add.
func (m *Mesh) Remove(f *Triangle) {
	if !m.faces[f] {
		return
	}
	delete(m.faces, f)
	v2f := m.getVertexToFaceOrNil()
	if v2f != nil {
		for _, p := range f {
			m.removeFaceFromVertex(v2f, f, p)
		}
	}
}

func (m *Mesh) removeFaceFromVertex(v2f *CoordToFaces, f *Triangle, p Coord3D) {
	s := v2f.Value(p)
	for i, f1 := range s {
		if f1 == f {
			essentials.UnorderedDelete(&s, i)
			break
		}
	}
	if len(s) == 0 {
		v2f.Delete(p)
	} else {
		v2f.Store(p, s)
	}
}

// Contains checks if f has been added to the mesh.
func (m *Mesh) Contains(f *Triangle) bool {
	_, ok := m.faces[f]
	return ok
}

// Iterate calls f for every triangle in m in an arbitrary
// order.
//
// If f adds or removes triangles, they will not be visited.
func (m *Mesh) Iterate(f func(*Triangle)) {
	m.IterateSorted(f, nil)
}

// IterateSorted is like Iterate, but it first sorts all
// the triangles according to a less than function, cmp.
func (m *Mesh) IterateSorted(f func(*Triangle), cmp func(f1, f2 *Triangle) bool) {
	all := m.TriangleSlice()
	if cmp != nil {
		sort.Slice(all, func(i, j int) bool {
			return cmp(all[i], all[j])
		})
	}
	for _, face := range all {
		if m.faces[face] {
			f(face)
		}
	}
}

// IterateVertices calls f for every vertex in m in an
// arbitrary order.
//
// If f adds or removes vertices, they will not be
// visited.
func (m *Mesh) IterateVertices(f func(c Coord3D)) {
	v2f := m.getVertexToFace()
	for _, c := range m.VertexSlice() {
		if _, ok := v2f.Load(c); ok {
			f(c)
		}
	}
}

// Neighbors gets all the triangles with a side touching a
// given triangle f.
//
// The triangle f itself is not included in the results.
//
// The triangle f needn't be in the mesh. However, if it is
// not in the mesh, but an equivalent triangle is, then said
// equivalent triangle will be in the results.
func (m *Mesh) Neighbors(f *Triangle) []*Triangle {
	counts := m.neighborsWithCounts(f)
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
	v2f := m.getVertexToFace()
	for _, p := range t {
		for _, t1 := range v2f.Value(p) {
			if t1 != t {
				counts[t1]++
			}
		}
	}
	return counts
}

// Find gets all the triangles that contain all of the passed
// points.
//
// For example, to find all triangles containing a line from
// from p1 to p2, you could do m.Find(p1, p2).
func (m *Mesh) Find(ps ...Coord3D) []*Triangle {
	if len(ps) == 1 {
		return append([]*Triangle{}, m.getVertexToFace().Value(ps[0])...)
	}

	faces := m.getVertexToFace().Value(ps[0])
	res := make([]*Triangle, 0, len(faces))

FaceLoop:
	for _, t := range faces {
		for _, p := range ps[1:] {
			if p != t[0] && p != t[1] && p != t[2] {
				continue FaceLoop
			}
		}
		res = append(res, t)
	}

	return res
}

// Scale creates a new mesh by scaling the coordinates by
// a factor s.
func (m *Mesh) Scale(s float64) *Mesh {
	return m.MapCoords(XYZ(s, s, s).Mul)
}

// Translate returns a mesh with all coordinates added to
// a given vector.
func (m *Mesh) Translate(v Coord3D) *Mesh {
	return m.MapCoords(v.Add)
}

// MapCoords creates a new mesh by transforming all of the
// coordinates according to the function f.
func (m *Mesh) MapCoords(f func(Coord3D) Coord3D) *Mesh {
	mapping := map[Coord3D]Coord3D{}
	if v2f := m.getVertexToFaceOrNil(); v2f != nil {
		v2f.KeyRange(func(c Coord3D) bool {
			mapping[c] = f(c)
			return true
		})
	} else {
		for t := range m.faces {
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

// SaveMaterialOBJ saves the mesh to a zip file with a
// per-triangle material.
func (m *Mesh) SaveMaterialOBJ(path string, colorFunc func(t *Triangle) [3]float64) error {
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "save material OBJ")
	}
	defer f.Close()
	err = WriteMaterialOBJ(f, m.TriangleSlice(), colorFunc)
	if err != nil {
		return errors.Wrap(err, "save material OBJ")
	}
	return nil
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
	ts := make([]*Triangle, 0, len(m.faces))
	for t := range m.faces {
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
	v2f := m.getVertexToFace()
	vertices := make([]Coord3D, 0, v2f.Len())
	v2f.KeyRange(func(v Coord3D) bool {
		vertices = append(vertices, v)
		return true
	})
	return vertices
}

// Min gets the component-wise minimum across all the
// vertices in the mesh.
func (m *Mesh) Min() Coord3D {
	if len(m.faces) == 0 {
		return Coord3D{}
	}
	var result Coord3D
	var firstFlag bool
	for t := range m.faces {
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
	if len(m.faces) == 0 {
		return Coord3D{}
	}
	var result Coord3D
	var firstFlag bool
	for t := range m.faces {
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

func (m *Mesh) getVertexToFace() *CoordToFaces {
	v2f := m.getVertexToFaceOrNil()
	if v2f != nil {
		return v2f
	}

	// Use a lock to ensure two different maps aren't
	// created and returned on different Goroutines.
	m.v2fCreateLock.Lock()
	defer m.v2fCreateLock.Unlock()

	// Another goroutine could have created a map while we
	// waited on the lock.
	v2f = m.getVertexToFaceOrNil()
	if v2f != nil {
		return v2f
	}

	v2f = NewCoordToFaces()
	for t := range m.faces {
		for _, p := range t {
			addFaceToVertex(v2f, p, t)
		}
	}
	m.vertexToFace.Store(v2f)

	return v2f
}

func (m *Mesh) getVertexToFaceOrNil() *CoordToFaces {
	res := m.vertexToFace.Load()
	if res == nil {
		return nil
	}
	return res.(*CoordToFaces)
}

func addFaceToVertex(v2f *CoordToFaces, k Coord3D, v *Triangle) {
	f := v2f.Value(k)
	f = append(f, v)
	v2f.Store(k, f)
}
