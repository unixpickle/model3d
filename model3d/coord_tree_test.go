// Generated from templates/coord_tree_test.template

package model3d

import (
	"math"
	"math/rand"
	"testing"

	"github.com/unixpickle/essentials"
)

func TestCoordTreeContains(t *testing.T) {
	coords := make([]Coord3D, 1000)
	for i := range coords {
		coords[i] = NewCoord3DRandNorm()
	}
	tree := NewCoordTree(coords)

	for _, c := range coords {
		if !tree.Contains(c) {
			t.Errorf("missing coordinate: %v", c)
		}
	}
	for i := 0; i < 1000; i++ {
		if tree.Contains(NewCoord3DRandNorm()) {
			t.Error("random point should not be contained (very low probability)")
		}
	}
}

func TestCoordTreeContainsDuplicates(t *testing.T) {
	coords := make([]Coord3D, 1000)
	for i := range coords {
		coords[i] = NewCoord3DRandNorm()
	}
	coords = append(coords, coords[0:100]...)
	coords = append(coords, coords[0:100]...)
	tree := NewCoordTree(coords)

	for _, c := range coords {
		if !tree.Contains(c) {
			t.Errorf("missing coordinate: %v", c)
		}
	}
	for i := 0; i < 1000; i++ {
		if tree.Contains(NewCoord3DRandNorm()) {
			t.Error("random point should not be contained (very low probability)")
		}
	}
}

func TestCoordTreeNearestNeighbor(t *testing.T) {
	coords := make([]Coord3D, 1000)
	for i := range coords {
		coords[i] = NewCoord3DRandNorm()
	}
	coords = append(coords, coords[0:100]...)
	coords = append(coords, coords[0:100]...)

	naiveNearest := func(c Coord3D) Coord3D {
		bestDist := math.Inf(1)
		bestCoord := Coord3D{}
		for _, p := range coords {
			dist := p.Dist(c)
			if dist < bestDist {
				bestDist = dist
				bestCoord = p
			}
		}
		return bestCoord
	}

	tree := NewCoordTree(coords)

	for _, c := range coords {
		if tree.NearestNeighbor(c) != c {
			t.Errorf("bad neighbor for coordinate: %v", c)
		}
	}
	for i := 0; i < 1000; i++ {
		p := NewCoord3DRandNorm()
		if tree.NearestNeighbor(p) != naiveNearest(p) {
			t.Error("incorrect nearest neighbor for random point")
		}
	}
	// Make sure axis-value collisions don't break
	// the algorithm.
	for i := 0; i < 1000; i++ {
		p := coords[rand.Intn(len(coords))]
		p.X = rand.NormFloat64()
		if tree.NearestNeighbor(p) != naiveNearest(p) {
			t.Error("incorrect nearest neighbor for axis-similar point")
		}
	}
}

func TestCoordTreeKNN(t *testing.T) {
	coords := make([]Coord3D, 200)
	for i := range coords {
		coords[i] = NewCoord3DRandNorm()
	}
	coords = append(coords, coords[0:50]...)
	coords = append(coords, coords[0:50]...)

	naiveKNN := func(k int, c Coord3D) []Coord3D {
		dists := make([]float64, len(coords))
		for i, p := range coords {
			dists[i] = p.Dist(c)
		}
		sorted := append([]Coord3D{}, coords...)
		essentials.VoodooSort(dists, func(i, j int) bool {
			return dists[i] < dists[j]
		}, sorted)
		return sorted[:essentials.MinInt(k, len(sorted))]
	}
	pointsEqual := func(p1, p2 []Coord3D) bool {
		if len(p1) != len(p2) {
			return false
		}
		for i, x := range p1 {
			if x != p2[i] {
				return false
			}
		}
		return true
	}

	tree := NewCoordTree(coords)

	for _, c := range coords {
		if tree.KNN(1, c)[0] != c {
			t.Errorf("bad 1-nn for coordinate: %v", c)
		}
	}
	for i := 0; i < 1000; i++ {
		p := NewCoord3DRandNorm()
		k := rand.Intn(10)
		if !pointsEqual(tree.KNN(k, p), naiveKNN(k, p)) {
			t.Errorf("incorrect nearest neighbor for random point, k=%d", k)
		}
	}
	// Make sure axis-value collisions don't break
	// the algorithm.
	for i := 0; i < 1000; i++ {
		p := coords[rand.Intn(len(coords))]
		p.X = rand.NormFloat64()
		k := rand.Intn(10)
		if !pointsEqual(tree.KNN(k, p), naiveKNN(k, p)) {
			t.Errorf("incorrect nearest neighbor for axis-similar point, k=%d", k)
		}
	}
	// Test case when we don't find enough points.
	p := NewCoord3DRandNorm()
	if !pointsEqual(tree.KNN(len(coords)*2, p), naiveKNN(len(coords)*2, p)) {
		t.Error("incorrect nearest neighbor for k > len(coords)")
	}
}

func TestCoordTreeSphereCollision(t *testing.T) {
	coords := make([]Coord3D, 1000)
	for i := range coords {
		coords[i] = NewCoord3DRandNorm()
	}
	coords = append(coords, coords[0:100]...)
	coords = append(coords, coords[0:100]...)

	tree := NewCoordTree(coords)

	randomRadius := func(c Coord3D) (float64, bool) {
		actualRadius := c.Dist(tree.NearestNeighbor(c))
		scale := rand.Float64() * 2
		radius := actualRadius * scale
		return radius, scale >= 1
	}
	checkCollision := func(c Coord3D) {
		radius, expected := randomRadius(c)
		if tree.SphereCollision(c, radius) != expected {
			t.Errorf("expected collision %v but got %v", expected, !expected)
		}
	}

	for i := 0; i < 1000; i++ {
		p := NewCoord3DRandNorm()
		checkCollision(p)
	}
	// Make sure axis-value collisions don't break
	// the algorithm.
	for i := 0; i < 1000; i++ {
		p := coords[rand.Intn(len(coords))]
		p.X = rand.NormFloat64()
		checkCollision(p)
	}
}
