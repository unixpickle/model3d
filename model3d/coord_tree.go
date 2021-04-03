// Generated from templates/coord_tree.template

package model3d

import (
	"math"
	"sort"
)

// A CoordTree is a k-d tree over Coord3Ds.
//
// A nil *CoordTree represents an empty tree.
type CoordTree struct {
	Coord Coord3D

	// SplitAxis is the dimension to split on for branches.
	SplitAxis int

	// At least one of these is non-nil for branches.
	LessThan     *CoordTree
	GreaterEqual *CoordTree
}

func NewCoordTree(points []Coord3D) *CoordTree {
	var sorted [3][]*flaggedCoord
	sorted[0] = make([]*flaggedCoord, len(points))
	for i, p := range points {
		sorted[0][i] = &flaggedCoord{Coord: p, Array: p.Array()}
	}
	for i := 1; i < 3; i++ {
		sorted[i] = append([]*flaggedCoord{}, sorted[0]...)
	}
	for i := 0; i < 3; i++ {
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
func (c *CoordTree) Contains(p Coord3D) bool {
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
func (c *CoordTree) NearestNeighbor(p Coord3D) Coord3D {
	if c == nil {
		panic("cannot lookup neighbor in empty tree")
	}
	res := Coord3D{}
	bound := math.Inf(1)
	c.nearestNeighbor(p, &res, &bound)
	return res
}

func (c *CoordTree) nearestNeighbor(p Coord3D, res *Coord3D, bound *float64) {
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

// SphereCollision checks if the sphere centered at point
// p with radius r contains any points in the tree.
func (c *CoordTree) SphereCollision(p Coord3D, r float64) bool {
	return c.sphereCollision(p, r*r)
}

func (c *CoordTree) sphereCollision(p Coord3D, rSquared float64) bool {
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
func (c *CoordTree) Slice() []Coord3D {
	if c == nil {
		return nil
	}
	value := c.LessThan.Slice()
	value = append(value, c.Coord)
	value = append(value, c.GreaterEqual.Slice()...)
	return value
}

func newCoordTreeSorted(coords [3][]*flaggedCoord, axis int) *CoordTree {
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
	left := [3][]*flaggedCoord{}
	right := [3][]*flaggedCoord{}
	for i := 0; i < 3; i++ {
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

	nextAxis := (axis + 1) % 3
	return &CoordTree{
		Coord:        splitCoord.Coord,
		SplitAxis:    axis,
		LessThan:     newCoordTreeSorted(left, nextAxis),
		GreaterEqual: newCoordTreeSorted(right, nextAxis),
	}
}

type flaggedCoord struct {
	Coord Coord3D
	Array [3]float64
	Flag  bool
}
