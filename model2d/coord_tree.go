// Generated from templates/coord_tree.template

package model2d

import (
	"math"
	"sort"
)

// A CoordTree is a k-d tree over Coords.
//
// A nil *CoordTree represents an empty tree.
type CoordTree struct {
	Coord Coord

	// SplitAxis is the dimension to split on for branches.
	SplitAxis int

	// At least one of these is non-nil for branches.
	LessThan     *CoordTree
	GreaterEqual *CoordTree
}

func NewCoordTree(points []Coord) *CoordTree {
	var sorted [2][]*flaggedCoord
	sorted[0] = make([]*flaggedCoord, len(points))
	for i, p := range points {
		sorted[0][i] = &flaggedCoord{Coord: p, Array: p.Array()}
	}
	for i := 1; i < 2; i++ {
		sorted[i] = append([]*flaggedCoord{}, sorted[0]...)
	}
	for i := 0; i < 2; i++ {
		sortMe := sorted[i]
		sort.Slice(sortMe, func(j, k int) bool {
			arr1 := sortMe[j].Array
			arr2 := sortMe[k].Array
			return arr1[i] < arr2[i]
		})
	}
	return newCoordTreeSorted(sorted, 0)
}

// Leaf returns true if this tree contains 1 or fewer
// points.
func (c *CoordTree) Leaf() bool {
	return c == nil || (c.LessThan == nil && c.GreaterEqual == nil)
}

// Empty returns true if c contains no points.
func (c *CoordTree) Empty() bool {
	return c == nil
}

// Contains checks if any point in the tree is exactly
// equal to p.
func (c *CoordTree) Contains(p Coord) bool {
	if c == nil {
		return false
	}
	if c.Coord == p {
		return true
	}
	splitValue := c.Coord.Array()[c.SplitAxis]
	pValue := p.Array()[c.SplitAxis]
	if pValue < splitValue {
		return c.LessThan.Contains(p)
	} else {
		return c.GreaterEqual.Contains(p)
	}
}

// NearestNeighbor gets the closest coordinate to p in the
// tree.
//
// This will panic() if c is empty.
func (c *CoordTree) NearestNeighbor(p Coord) Coord {
	if c == nil {
		panic("cannot lookup neighbor in empty tree")
	}
	res := Coord{}
	bound := math.Inf(1)
	c.nearestNeighbor(p, &res, &bound)
	return res
}

func (c *CoordTree) nearestNeighbor(p Coord, res *Coord, bound *float64) {
	if c == nil {
		return
	}
	dist := p.SquaredDist(c.Coord)
	if dist < *bound {
		*bound = dist
		*res = c.Coord
	}
	planeDist := c.Coord.Array()[c.SplitAxis] - p.Array()[c.SplitAxis]
	if planeDist > 0 {
		c.LessThan.nearestNeighbor(p, res, bound)
	} else {
		c.GreaterEqual.nearestNeighbor(p, res, bound)
	}
	// Attempt to eliminate out-of-bounds half spaces.
	if planeDist > 0 && planeDist*planeDist < *bound {
		c.GreaterEqual.nearestNeighbor(p, res, bound)
	} else if planeDist <= 0 && planeDist*planeDist < *bound {
		c.LessThan.nearestNeighbor(p, res, bound)
	}
}

// KNN gets the closest K coordinates to p in the tree.
// The results are sorted by ascending distance.
//
// If there are fewer than K coordinates in the tree, then
// fewer than K coordinates are returned.
func (c *CoordTree) KNN(k int, p Coord) []Coord {
	if k == 0 {
		return nil
	}
	res := &knnResults{Max: k}
	c.knn(p, res)
	return res.Coords
}

func (c *CoordTree) knn(p Coord, res *knnResults) {
	if c == nil {
		return
	}
	dist := p.SquaredDist(c.Coord)
	res.Insert(c.Coord, dist)
	planeDist := c.Coord.Array()[c.SplitAxis] - p.Array()[c.SplitAxis]
	if planeDist > 0 {
		c.LessThan.knn(p, res)
	} else {
		c.GreaterEqual.knn(p, res)
	}
	// Attempt to eliminate out-of-bounds half spaces.
	if planeDist > 0 && planeDist*planeDist < res.MaxDist() {
		c.GreaterEqual.knn(p, res)
	} else if planeDist <= 0 && planeDist*planeDist < res.MaxDist() {
		c.LessThan.knn(p, res)
	}
}

// SphereCollision checks if the sphere centered at point
// p with radius r contains any points in the tree.
func (c *CoordTree) SphereCollision(p Coord, r float64) bool {
	return c.sphereCollision(p, r*r)
}

func (c *CoordTree) sphereCollision(p Coord, rSquared float64) bool {
	if c == nil {
		return false
	}
	dist := p.SquaredDist(c.Coord)
	if dist <= rSquared {
		return true
	}
	planeDist := c.Coord.Array()[c.SplitAxis] - p.Array()[c.SplitAxis]
	if planeDist > 0 {
		if c.LessThan.sphereCollision(p, rSquared) {
			return true
		}
	} else {
		if c.GreaterEqual.sphereCollision(p, rSquared) {
			return true
		}
	}
	if planeDist > 0 && planeDist*planeDist <= rSquared {
		return c.GreaterEqual.sphereCollision(p, rSquared)
	} else if planeDist <= 0 && planeDist*planeDist <= rSquared {
		return c.LessThan.sphereCollision(p, rSquared)
	} else {
		return false
	}
}

// Slice combines the points back into a slice.
//
// The order will be from the first (less than) leaf to
// the final (greater than) leaf, with intermediate nodes
// interspersed throughout the middle.
func (c *CoordTree) Slice() []Coord {
	if c == nil {
		return nil
	}
	value := c.LessThan.Slice()
	value = append(value, c.Coord)
	value = append(value, c.GreaterEqual.Slice()...)
	return value
}

func newCoordTreeSorted(coords [2][]*flaggedCoord, axis int) *CoordTree {
	if len(coords[0]) == 0 {
		return nil
	} else if len(coords[0]) == 1 {
		return &CoordTree{
			Coord: coords[0][0].Coord,
		}
	}

	splitCoord := coords[axis][len(coords[axis])/2]
	splitValue := splitCoord.Array[axis]
	for _, c := range coords[axis] {
		if c.Array[axis] < splitValue {
			c.Flag = false
		} else {
			c.Flag = true
		}
	}

	// Maintain sorted left and right branches.
	left := [2][]*flaggedCoord{}
	right := [2][]*flaggedCoord{}
	for i := 0; i < 2; i++ {
		for _, c := range coords[i] {
			if c == splitCoord {
				continue
			}
			if c.Flag {
				right[i] = append(right[i], c)
			} else {
				left[i] = append(left[i], c)
			}
		}
	}

	nextAxis := (axis + 1) % 2
	return &CoordTree{
		Coord:        splitCoord.Coord,
		SplitAxis:    axis,
		LessThan:     newCoordTreeSorted(left, nextAxis),
		GreaterEqual: newCoordTreeSorted(right, nextAxis),
	}
}

type flaggedCoord struct {
	Coord Coord
	Array [2]float64
	Flag  bool
}

type knnResults struct {
	Max    int
	Coords []Coord
	Dists  []float64
}

func (s *knnResults) MaxDist() float64 {
	if len(s.Dists) < s.Max {
		return math.Inf(1)
	}
	return s.Dists[s.Max-1]
}

func (s *knnResults) Insert(c Coord, d float64) {
	if d >= s.MaxDist() {
		return
	}
	index := sort.SearchFloat64s(s.Dists, d)
	if len(s.Dists) < s.Max {
		s.Dists = append(s.Dists, 0)
		s.Coords = append(s.Coords, Coord{})
	}
	copy(s.Dists[index+1:], s.Dists[index:])
	copy(s.Coords[index+1:], s.Coords[index:])
	s.Coords[index] = c
	s.Dists[index] = d
}
