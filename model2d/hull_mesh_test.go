package model2d

import (
	"math"
	"math/rand"
	"testing"
)

func TestConvexHullMeshBasic(t *testing.T) {
	cases := []struct {
		name   string
		points []Coord
		want   []Segment
	}{
		{
			name:   "single",
			points: []Coord{XY(1, 2)},
			want:   nil,
		},
		{
			name:   "two_points",
			points: []Coord{XY(0, 0), XY(2, 0)},
			want:   []Segment{{XY(0, 0), XY(2, 0)}},
		},
		{
			name:   "triangle",
			points: []Coord{XY(0, 0), XY(1, 0), XY(0, 1)},
			want: []Segment{
				{XY(0, 0), XY(1, 0)},
				{XY(1, 0), XY(0, 1)},
				{XY(0, 1), XY(0, 0)},
			},
		},
		{
			name:   "square_with_interior",
			points: []Coord{XY(0, 0), XY(1, 0), XY(1, 1), XY(0, 1), XY(0.5, 0.5)},
			want: []Segment{
				{XY(0, 0), XY(1, 0)},
				{XY(1, 0), XY(1, 1)},
				{XY(1, 1), XY(0, 1)},
				{XY(0, 1), XY(0, 0)},
			},
		},
		{
			name:   "colinear",
			points: []Coord{XY(0, 0), XY(1, 0), XY(2, 0), XY(0.5, 0), XY(2, 0)},
			want:   []Segment{{XY(0, 0), XY(2, 0)}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ConvexHullMesh(tc.points)
			want := NewMesh()
			for i := range tc.want {
				seg := tc.want[i]
				want.Add(&seg)
			}
			if !meshEdgeSetEqual(got, want) {
				t.Fatalf("edge sets differ: got=%v want=%v",
					meshEdgeSet(got), meshEdgeSet(want))
			}
		})
	}
}

func TestConvexHullMeshRandomMatchesNaive(t *testing.T) {
	rng := rand.New(rand.NewSource(1337))
	for i := 0; i < 5; i++ {
		points := make([]Coord, 10)
		for j := range points {
			points[j] = XY(rng.Float64()*2-1, rng.Float64()*2-1)
		}
		got := ConvexHullMesh(points)
		want := naiveConvexHullMesh(points)
		MustValidateMesh(t, got, true)
		MustValidateMesh(t, want, true)
		if !meshEdgeSetEqual(got, want) {
			t.Fatalf("random draw %d mismatch: got=%v want=%v",
				i, meshEdgeSet(got), meshEdgeSet(want))
		}
	}
}

func naiveConvexHullMesh(points []Coord) *Mesh {
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

	start := unique[0]
	for _, p := range unique[1:] {
		if p.Y < start.Y || (p.Y == start.Y && p.X < start.X) {
			start = p
		}
	}

	scale := maxCoordAbs(unique)
	if scale == 0 {
		scale = 1
	}
	eps := scale * scale * 1e-12

	hull := []Coord{start}
	curr := start
	for {
		var next Coord
		found := false
		for _, p := range unique {
			if p == curr {
				continue
			}
			if !found {
				next = p
				found = true
				continue
			}
			cross := cross2(curr, next, p)
			if cross < -eps {
				next = p
			} else if math.Abs(cross) <= eps {
				if curr.SquaredDist(p) > curr.SquaredDist(next) {
					next = p
				}
			}
		}
		if !found || next == start {
			break
		}
		hull = append(hull, next)
		curr = next
		if len(hull) > len(unique)+1 {
			break
		}
	}

	if len(hull) < 2 {
		return NewMesh()
	}
	if len(hull) == 2 {
		m := NewMesh()
		m.Add(&Segment{hull[0], hull[1]})
		return m
	}
	if !isClockwise(hull) {
		for i, j := 0, len(hull)-1; i < j; i, j = i+1, j-1 {
			hull[i], hull[j] = hull[j], hull[i]
		}
	}

	m := NewMesh()
	for i := 0; i < len(hull); i++ {
		p1 := hull[i]
		p2 := hull[(i+1)%len(hull)]
		m.Add(&Segment{p1, p2})
	}
	return m
}

func meshEdgeSet(m *Mesh) map[Segment]struct{} {
	res := map[Segment]struct{}{}
	for _, seg := range m.SegmentSlice() {
		a, b := seg[0], seg[1]
		if lessCoord(b, a) {
			a, b = b, a
		}
		res[Segment{a, b}] = struct{}{}
	}
	return res
}

func meshEdgeSetEqual(m1, m2 *Mesh) bool {
	s1 := meshEdgeSet(m1)
	s2 := meshEdgeSet(m2)
	if len(s1) != len(s2) {
		return false
	}
	for k := range s1 {
		if _, ok := s2[k]; !ok {
			return false
		}
	}
	return true
}

func lessCoord(a, b Coord) bool {
	if a.X != b.X {
		return a.X < b.X
	}
	return a.Y < b.Y
}

func isClockwise(poly []Coord) bool {
	var area2 float64
	for i, p := range poly {
		q := poly[(i+1)%len(poly)]
		area2 += p.X*q.Y - q.X*p.Y
	}
	return area2 < 0
}
