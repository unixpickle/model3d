package model3d

import "math"

const DefaultDecimatorMinAspectRatio = 0.1

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
	// This is measured in radians.
	// It is unused if NoEdgePreservation and CornerEdges
	// are both true.
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

	// If true, eliminate corner edges.
	CornerEdges bool

	// MinimumAspectRatio is the minimum aspect ratio for
	// triangulation splits.
	//
	// If 0, a default of DefaultDecimatorMinAspectRatio
	// is used.
	MinimumAspectRatio float64
}

func (d *Decimator) fillLoop(avgPlane *plane, coords []*ptrCoord) []*ptrTriangle {
	if len(coords) < 3 {
		panic("invalid number of loop coordinates")
	} else if len(coords) == 3 {
		return []*ptrTriangle{newPtrTriangle(coords[0], coords[1], coords[2])}
	}

	var bestAspectRatio float64
	var bestLoop1, bestLoop2 []*ptrCoord
	for i, c1 := range coords {
		for j := i + 2; j < len(coords); j++ {
			c2 := coords[j]
			sepLine := c2.Coord3D.Sub(c1.Coord3D)
			sepNormal := sepLine.Cross(avgPlane.Normal)
			sepPlane := newPlanePoint(sepNormal, c1.Coord3D)

			loop1 := createSubloop(coords, i, j)
			sign1, minAbs1 := subloopSplitDist(loop1, sepPlane)
			if sign1 == 0 {
				continue
			}
			loop2 := createSubloop(coords, j, i)
			sign2, minAbs2 := subloopSplitDist(loop2, sepPlane)
			if sign2 == 0 || sign2 == sign1 {
				continue
			}

			aspectRatio := math.Min(minAbs1, minAbs2) / sepLine.Norm()
			if bestAspectRatio == 0 || math.Abs(aspectRatio-1) < math.Abs(bestAspectRatio-1) {
				bestAspectRatio = aspectRatio
				bestLoop1, bestLoop2 = loop1, loop2
			}
		}
	}

	if bestLoop1 == nil {
		return nil
	}

	minRatio := d.MinimumAspectRatio
	if minRatio == 0 {
		minRatio = DefaultDecimatorMinAspectRatio
	}
	if bestAspectRatio < minRatio {
		return nil
	}

	tris1 := d.fillLoop(avgPlane, bestLoop1)
	if tris1 == nil {
		return nil
	}
	tris2 := d.fillLoop(avgPlane, bestLoop2)
	if tris2 == nil {
		for _, t := range tris1 {
			t.RemoveCoords()
		}
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
