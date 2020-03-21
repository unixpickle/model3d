package model3d

import (
	"math"
)

const (
	DefaultDecimatorMinAspectRatio = 0.1
	DefaultDecimatorFeatureAngle   = 0.5
)

// DecimateSimple decimates a mesh using a specified
// distance epsilon combined with default parameters.
//
// For more fine-grained control, use Decimator.
func DecimateSimple(m *Mesh, epsilon float64) *Mesh {
	d := Decimator{PlaneDistance: epsilon, BoundaryDistance: epsilon}
	return d.Decimate(m)
}

// Decimator implements a decimation algorithm to simplify
// triangle meshes.
//
// This may only be applied to closed, manifold meshes.
// Thus, all edges are touching exactly two triangles, and
// there are no singularities or holes.
//
// The algorithm is described in:
// "Decimation of Triangle Meshes" - William J. Schroeder,
// Jonathan A. Zarge and William E. Lorensen.
// https://webdocs.cs.ualberta.ca/~lin/ABProject/papers/4.pdf.
type Decimator struct {
	// The minimum dihedral angle between two triangles
	// to consider an edge a "feature edge".
	//
	// If 0, DefaultDecimatorFeatureAngle is used.
	//
	// This is measured in radians.
	FeatureAngle float64

	// The maximum distance for a vertex to be from its
	// average plane for it to be deleted.
	PlaneDistance float64

	// The maximum distance for a vertex to be from the
	// line defining a feature edge.
	BoundaryDistance float64

	// If true, use PlaneDistance to evaluate all vertices
	// rather than consulting BoundaryDistance.
	NoEdgePreservation bool

	// If true, eliminate corner vertices.
	EliminateCorners bool

	// MinimumAspectRatio is the minimum aspect ratio for
	// triangulation splits.
	//
	// If 0, a default of DefaultDecimatorMinAspectRatio
	// is used.
	MinimumAspectRatio float64
}

// Decimate applies the decimation algorithm to m,
// producing a new mesh.
func (d *Decimator) Decimate(m *Mesh) *Mesh {
	pm := newPtrMeshMesh(m)
	d.decimatePtrMesh(pm)
	return pm.Mesh()
}

func (d *Decimator) decimatePtrMesh(p *ptrMesh) int {
	coords := map[*ptrCoord]struct{}{}
	p.Iterate(func(t *ptrTriangle) {
		for _, c := range t.Coords {
			coords[c] = struct{}{}
		}
	})
	var eliminated int
	for c := range coords {
		v := newDecVertex(c, d.FeatureAngle)
		if d.canRemoveVertex(v) && d.attemptRemoveVertex(p, v) {
			eliminated++
		}
	}
	return eliminated
}

func (d *Decimator) canRemoveVertex(v *decVertex) bool {
	if v.Simple() || (v.Edge() && d.NoEdgePreservation) || (v.Corner() && d.EliminateCorners) {
		// Use the distance to plane metric.
		return math.Abs(v.AvgPlane.Eval(v.Vertex.Coord3D)) < d.PlaneDistance
	} else if v.Edge() {
		// Use the distance to edge metric.
		seg := NewSegment(v.Loop[v.FeatureEndpoints[0]].Coord3D,
			v.Loop[v.FeatureEndpoints[1]].Coord3D)
		return seg.Dist(v.Vertex.Coord3D) < d.BoundaryDistance
	}
	return false
}

func (d *Decimator) attemptRemoveVertex(p *ptrMesh, v *decVertex) bool {
	var newTriangles []*ptrTriangle

	// Only preserve interior edge when connecting
	// the two points wouldn't cause an empty loop.
	if v.Edge() && v.FeatureEndpoints[1] != v.FeatureEndpoints[0]+1 &&
		v.FeatureEndpoints[0] != (v.FeatureEndpoints[1]+1)%len(v.Loop) {
		loop1, loop2, ratio := d.createSubloops(v.AvgPlane, v.Loop, v.FeatureEndpoints[0],
			v.FeatureEndpoints[1])
		if ratio == 0 {
			return false
		}
		newTriangles = d.fillLoops(v.AvgPlane, loop1, loop2)
	} else {
		newTriangles = d.fillLoop(v.AvgPlane, v.Loop)
	}

	if newTriangles == nil {
		return false
	}

	oldTriangles := append([]*ptrTriangle{}, v.Vertex.Triangles...)
	for _, t := range oldTriangles {
		p.Remove(t)
		t.RemoveCoords()
	}
	for _, t := range newTriangles {
		// fillLoop(s) explicitly don't add the triangles
		// to their coords to avoid unnecessary undoing.
		t.AddCoords()

		p.Add(t)
	}

	rollBack := func() {
		for _, t := range newTriangles {
			p.Remove(t)
			t.RemoveCoords()
		}
		for _, t := range oldTriangles {
			t.AddCoords()
			p.Add(t)
		}
	}

	// It is possible to eliminate the mesh too much, and
	// create a flattened section (duplicated triangle).
	for _, t := range newTriangles {
		for _, t1 := range t.Coords[0].Triangles {
			if t1 != t && t1.Contains(t.Coords[0]) && t1.Contains(t.Coords[1]) &&
				t1.Contains(t.Coords[2]) {
				rollBack()
				return false
			}
		}
	}

	// Also make sure we don't create duplicate edges.
	for _, t := range newTriangles {
		for _, s := range t.Segments() {
			if len(s.Triangles()) != 2 {
				rollBack()
				return false
			}
		}
	}

	return true
}

func (d *Decimator) fillLoop(avgPlane *plane, coords []*ptrCoord) []*ptrTriangle {
	if len(coords) < 3 {
		panic("invalid number of loop coordinates")
	} else if len(coords) == 3 {
		return []*ptrTriangle{
			{Coords: [3]*ptrCoord{coords[0], coords[2], coords[1]}},
		}
	}

	var bestAspectRatio float64
	var bestLoop1, bestLoop2 []*ptrCoord
	for i := range coords {
		for j := i + 2; j < len(coords); j++ {
			if i+len(coords)-j < 2 {
				continue
			}
			loop1, loop2, aspectRatio := d.createSubloops(avgPlane, coords, i, j)
			if aspectRatio == 0 {
				continue
			}
			if bestAspectRatio == 0 || math.Abs(aspectRatio-1) < math.Abs(bestAspectRatio-1) {
				bestAspectRatio = aspectRatio
				bestLoop1, bestLoop2 = loop1, loop2
			}
		}
	}

	if bestAspectRatio == 0 {
		return nil
	}

	minRatio := d.MinimumAspectRatio
	if minRatio == 0 {
		minRatio = DefaultDecimatorMinAspectRatio
	}
	if bestAspectRatio < minRatio {
		return nil
	}

	return d.fillLoops(avgPlane, bestLoop1, bestLoop2)
}

func (d *Decimator) createSubloops(avgPlane *plane, coords []*ptrCoord, i, j int) (loop1,
	loop2 []*ptrCoord, aspectRatio float64) {
	c1 := coords[i]
	c2 := coords[j]

	sepLine := c2.Coord3D.Sub(c1.Coord3D)
	sepNormal := sepLine.Cross(avgPlane.Normal).Normalize()
	sepPlane := newPlanePoint(sepNormal, c1.Coord3D)

	loop1 = createSubloop(coords, i, j)
	sign1, minAbs1 := subloopSplitDist(loop1, sepPlane)
	if sign1 == 0 {
		return nil, nil, 0
	}
	loop2 = createSubloop(coords, j, i)
	sign2, minAbs2 := subloopSplitDist(loop2, sepPlane)
	if sign2 == 0 || sign2 == sign1 {
		return nil, nil, 0
	}
	aspectRatio = math.Min(minAbs1, minAbs2) / sepLine.Norm()
	return
}

func (d *Decimator) fillLoops(avgPlane *plane, loop1, loop2 []*ptrCoord) []*ptrTriangle {
	tris1 := d.fillLoop(avgPlane, loop1)
	if tris1 == nil {
		return nil
	}
	tris2 := d.fillLoop(avgPlane, loop2)
	if tris2 == nil {
		return nil
	}
	return append(tris1, tris2...)
}

func createSubloop(coords []*ptrCoord, start, end int) []*ptrCoord {
	if end < start {
		end += len(coords)
	}
	res := make([]*ptrCoord, 0, end-start+1)
	for i := start; i <= end; i++ {
		res = append(res, coords[i%len(coords)])
	}
	return res
}

func subloopSplitDist(coords []*ptrCoord, p *plane) (sign int, minAbs float64) {
	for i, c := range coords[1 : len(coords)-1] {
		dist := p.Eval(c.Coord3D)
		curSign := 1
		if dist == 0 {
			// Touching the separating plane.
			return 0, 0
		} else if dist < 0 {
			curSign = -1
		}
		if i == 0 {
			sign = curSign
			minAbs = math.Abs(dist)
		} else {
			if sign != curSign {
				// There is an edge passing the boundary.
				return 0, 0
			}
			minAbs = math.Min(minAbs, math.Abs(dist))
		}
	}
	return
}

// decVertex stores info relavant for deleting a given
// vertex.
type decVertex struct {
	// The vertex to consider for decimation.
	Vertex *ptrCoord

	// A loop of points around the vertex.
	Loop []*ptrCoord

	// AvgPlane is the average plane around the vertex.
	AvgPlane *plane

	// Loop point indices that are part of feature edges.
	FeatureEndpoints []int
}

func newDecVertex(v *ptrCoord, featureAngle float64) *decVertex {
	if featureAngle == 0 {
		featureAngle = DefaultDecimatorFeatureAngle
	}

	res := &decVertex{
		Vertex:   v,
		Loop:     v.SortLoops(),
		AvgPlane: newPlaneAvg(v.Triangles),
	}

	nextNormal := v.Triangles[0].Triangle().Normal()
	for i := range v.Triangles {
		t := v.Triangles[(i+1)%len(v.Triangles)]
		normal := nextNormal
		nextNormal = t.Triangle().Normal()
		angle := math.Acos(normal.Dot(nextNormal))
		if angle > featureAngle {
			res.FeatureEndpoints = append(res.FeatureEndpoints, i)
		}
	}

	return res
}

func (d *decVertex) Simple() bool {
	return len(d.FeatureEndpoints) == 0
}

func (d *decVertex) Edge() bool {
	return len(d.FeatureEndpoints) == 2
}

func (d *decVertex) Corner() bool {
	return !d.Simple() && !d.Edge()
}

// plane implements the plane Normal*X - Bias = 0.
type plane struct {
	Normal Coord3D
	Bias   float64
}

func newPlaneAvg(tris []*ptrTriangle) *plane {
	var normal Coord3D
	var avgPoint Coord3D
	var totalWeight float64
	for _, t := range tris {
		tri := t.Triangle()
		weight := tri.Area()
		totalWeight += weight
		normal = normal.Add(tri.Normal().Scale(weight))
		avgPoint = avgPoint.Add(tri[0].Add(tri[1]).Add(tri[2]).Scale(weight / 3.0))
	}
	normal = normal.Normalize()
	avgPoint = avgPoint.Scale(1 / totalWeight)

	return newPlanePoint(normal, avgPoint)
}

func newPlanePoint(normal, point Coord3D) *plane {
	return &plane{
		Normal: normal,
		Bias:   point.Dot(normal),
	}
}

// Eval evaluates the signed distance from the plane,
// assuming a unit normal.
func (p *plane) Eval(c Coord3D) float64 {
	return p.Normal.Dot(c) - p.Bias
}
