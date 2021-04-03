// Generated from templates/coord_tree_test.template

package model3d

import (
	"math"
	"math/rand"
	"testing"
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
