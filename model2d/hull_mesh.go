package model2d

import (
	"math"
	"sort"
)

// ConvexHullMesh computes the convex hull of points and returns it as a mesh.
//
// The hull edges are ordered clockwise (assuming the y-axis points up), so
// segment normals face outward. Colinear points on the hull are removed.
func ConvexHullMesh(points []Coord) *Mesh {
	unique := make([]Coord, 0, len(points))
	seen := NewCoordToNumber[int]()
	for _, p := range points {
		if _, ok := seen.Load(p); !ok {
			seen.Store(p, 1)
			unique = append(unique, p)
		}
	}

	if len(unique) < 2 {
		return NewMesh()
	}

	pivotIdx := 0
	for i := 1; i < len(unique); i++ {
		p, q := unique[i], unique[pivotIdx]
		if p.Y < q.Y || (p.Y == q.Y && p.X < q.X) {
			pivotIdx = i
		}
	}
	unique[0], unique[pivotIdx] = unique[pivotIdx], unique[0]
	pivot := unique[0]

	scale := maxCoordAbs(unique)
	if scale == 0 {
		scale = 1
	}
	eps := scale * scale * 1e-12

	others := append([]Coord{}, unique[1:]...)
	sort.Slice(others, func(i, j int) bool {
		a, b := others[i], others[j]
		cross := cross2(pivot, a, b)
		if math.Abs(cross) > eps {
			return cross > 0
		}
		da := pivot.SquaredDist(a)
		db := pivot.SquaredDist(b)
		if da != db {
			return da < db
		}
		if a.X != b.X {
			return a.X < b.X
		}
		return a.Y < b.Y
	})

	hull := []Coord{pivot}
	for _, p := range others {
		for len(hull) >= 2 {
			c := cross2(hull[len(hull)-2], hull[len(hull)-1], p)
			if c > eps {
				break
			}
			hull = hull[:len(hull)-1]
		}
		hull = append(hull, p)
	}

	if len(hull) < 2 {
		return NewMesh()
	}
	if len(hull) == 2 {
		m := NewMesh()
		m.Add(&Segment{hull[0], hull[1]})
		return m
	}

	for i, j := 0, len(hull)-1; i < j; i, j = i+1, j-1 {
		hull[i], hull[j] = hull[j], hull[i]
	}

	m := NewMesh()
	for i := 0; i < len(hull); i++ {
		p1 := hull[i]
		p2 := hull[(i+1)%len(hull)]
		m.Add(&Segment{p1, p2})
	}
	return m
}

func cross2(a, b, c Coord) float64 {
	return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
}

func maxCoordAbs(points []Coord) float64 {
	var maxAbs float64
	for _, p := range points {
		ax := math.Abs(p.X)
		ay := math.Abs(p.Y)
		if ax > maxAbs {
			maxAbs = ax
		}
		if ay > maxAbs {
			maxAbs = ay
		}
	}
	return maxAbs
}
