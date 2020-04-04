package model3d

// MarchingCubes turns a Solid into a surface mesh using a
// corrected marching cubes algorithm.
func MarchingCubes(s Solid, delta float64) *Mesh {
	table := mcLookupTable()

	spacer := newSquareSpacer(s, delta)
	cache := newSolidCache(s, spacer)

	mesh := NewMesh()

	spacer.IterateSquares(func(x, y, z int) {
		var intersections mcIntersections
		mask := mcIntersections(1)
		for i := 0; i < 2; i++ {
			for j := 0; j < 2; j++ {
				for k := 0; k < 2; k++ {
					x1 := x + k
					y1 := y + j
					z1 := z + i
					if cache.CornerValue(x1, y1, z1) {
						if x1 == 0 || x1 == len(spacer.Xs)-1 ||
							y1 == 0 || y1 == len(spacer.Ys)-1 ||
							z1 == 0 || z1 == len(spacer.Zs)-1 {
							panic("solid is true outside of bounds")
						}
						intersections |= mask
					}
					mask <<= 1
				}
			}
		}

		triangles := table[intersections]
		if len(triangles) > 0 {
			min := spacer.CornerCoord(x, y, z)
			max := spacer.CornerCoord(x+1, y+1, z+1)
			corners := mcCornerCoordinates(min, max)
			for _, t := range triangles {
				mesh.Add(t.Triangle(corners))
			}
		}
	})

	return mesh
}

// mcCorner is a corner index on a cube used for marching
// cubes.
//
// Ordered as:
//
//     (0, 0, 0), (1, 0, 0), (0, 1, 0), (1, 1, 0),
//     (0, 0, 1), (1, 0, 1), (0, 1, 1), (1, 1, 1)
//
// Here is a visualization of the cube indices:
//
//         6 + -----------------------+ 7
//          /|                       /|
//         / |                      / |
//        /  |                     /  |
//     4 +------------------------+ 5 |
//       |   |                    |   |
//       |   |                    |   |
//       |   |                    |   |
//       |   | 2                  |   | 3
//       |   +--------------------|---+
//       |  /                     |  /
//       | /                      | /
//       |/                       |/
//       +------------------------+
//      0                           1
//
type mcCorner uint8

// mcCornerCoordinates gets the coordinates of all eight
// corners for a cube.
func mcCornerCoordinates(min, max Coord3D) [8]Coord3D {
	return [8]Coord3D{
		min,
		{X: max.X, Y: min.Y, Z: min.Z},
		{X: min.X, Y: max.Y, Z: min.Z},
		{X: max.X, Y: max.Y, Z: min.Z},

		{X: min.X, Y: min.Y, Z: max.Z},
		{X: max.X, Y: min.Y, Z: max.Z},
		{X: min.X, Y: max.Y, Z: max.Z},
		max,
	}
}

// mcRotation represents a cube rotation for marching
// cubes.
//
// For corner c, rotation[c] is the new corner at that
// location.
type mcRotation [8]mcCorner

// allMcRotations gets all 24 possible rotations for a
// cube in marching cubes.
func allMcRotations() []mcRotation {
	// Create a generating basis.
	zRotation := mcRotation{2, 0, 3, 1, 6, 4, 7, 5}
	xRotation := mcRotation{2, 3, 6, 7, 0, 1, 4, 5}

	queue := []mcRotation{{0, 1, 2, 3, 4, 5, 6, 7}}
	resMap := map[mcRotation]bool{queue[0]: true}
	for len(queue) > 0 {
		next := queue[0]
		queue = queue[1:]
		resMap[next] = true
		for _, op := range []mcRotation{zRotation, xRotation} {
			rotated := op.Compose(next)
			if !resMap[rotated] {
				resMap[rotated] = true
				queue = append(queue, rotated)
			}
		}
	}

	var result []mcRotation
	for rotation := range resMap {
		result = append(result, rotation)
	}
	return result
}

// Compose combines two rotations.
func (m mcRotation) Compose(m1 mcRotation) mcRotation {
	var res mcRotation
	for i := range res {
		res[i] = m[m1[i]]
	}
	return res
}

// ApplyCorner applies the rotation to a corner.
func (m mcRotation) ApplyCorner(c mcCorner) mcCorner {
	return m[c]
}

// ApplyTriangle applies the rotation to a triangle.
func (m mcRotation) ApplyTriangle(t mcTriangle) mcTriangle {
	var res mcTriangle
	for i, c := range t {
		res[i] = m.ApplyCorner(c)
	}
	return res
}

// ApplyIntersections applies the rotation to an
// mcIntersections.
func (m mcRotation) ApplyIntersections(i mcIntersections) mcIntersections {
	var res mcIntersections
	for c := mcCorner(0); c < 8; c++ {
		if i.Inside(c) {
			res |= 1 << m.ApplyCorner(c)
		}
	}
	return res
}

// mcTriangle is a triangle constructed out of midpoints
// of edges of a cube.
// There are 6 corners because each pair of two represents
// an edge.
//
// The triangle is ordered in counter-clockwise order when
// looked upon from the outside.
type mcTriangle [6]mcCorner

// Triangle creates a real triangle out of the mcTriangle,
// given the corner coordinates.
func (m mcTriangle) Triangle(corners [8]Coord3D) *Triangle {
	return &Triangle{
		corners[m[0]].Mid(corners[m[1]]),
		corners[m[2]].Mid(corners[m[3]]),
		corners[m[4]].Mid(corners[m[5]]),
	}
}

// mcIntersections represents which corners on a cube are
// inside of a solid.
// Each corner is a bit, and 1 means inside.
type mcIntersections uint8

// newMcIntersections creates an mcIntersections using the
// corners that are inside the solid.
func newMcIntersections(trueCorners ...mcCorner) mcIntersections {
	if len(trueCorners) > 8 {
		panic("expected at most 8 corners")
	}
	var res mcIntersections
	for _, c := range trueCorners {
		res |= (1 << c)
	}
	return res
}

// Inside checks if a corner c is true.
func (m mcIntersections) Inside(c mcCorner) bool {
	return (m & (1 << c)) != 0
}

// mcLookupTable creates a full lookup table that maps
// each mcIntersection to a set of triangles.
func mcLookupTable() [256][]mcTriangle {
	rotations := allMcRotations()
	result := map[mcIntersections][]mcTriangle{}

	for baseInts, baseTris := range baseTriangleTable {
		for _, rot := range rotations {
			newInts := rot.ApplyIntersections(baseInts)
			if _, ok := result[newInts]; !ok {
				newTris := []mcTriangle{}
				for _, t := range baseTris {
					newTris = append(newTris, rot.ApplyTriangle(t))
				}
				result[newInts] = newTris
			}
		}
	}

	var resultArray [256][]mcTriangle
	for key, value := range result {
		resultArray[key] = value
	}
	return resultArray
}

// baseTriangleTable encodes the marching cubes lookup
// table (up to rotations) from:
// "A survey of the marching cubes algorithm" (2006).
// https://cg.informatik.uni-freiburg.de/intern/seminar/surfaceReconstruction_survey%20of%20marching%20cubes.pdf
var baseTriangleTable = map[mcIntersections][]mcTriangle{
	// Case 0-5
	newMcIntersections(): []mcTriangle{},
	newMcIntersections(0): []mcTriangle{
		{0, 1, 0, 2, 0, 4},
	},
	newMcIntersections(0, 1): []mcTriangle{
		{0, 4, 1, 5, 0, 2},
		{1, 5, 1, 3, 0, 2},
	},
	newMcIntersections(0, 5): []mcTriangle{
		{0, 1, 0, 2, 0, 4},
		{5, 7, 1, 5, 4, 5},
	},
	newMcIntersections(0, 7): []mcTriangle{
		{0, 1, 0, 2, 0, 4},
		{6, 7, 3, 7, 5, 7},
	},
	newMcIntersections(1, 2, 3): []mcTriangle{
		{0, 1, 1, 5, 0, 2},
		{0, 2, 1, 5, 2, 6},
		{2, 6, 1, 5, 3, 7},
	},

	// Case 6-11
	newMcIntersections(0, 1, 7): []mcTriangle{
		// Case 2.
		{0, 4, 1, 5, 0, 2},
		{1, 5, 1, 3, 0, 2},
		// End of case 4
		{6, 7, 3, 7, 5, 7},
	},
	newMcIntersections(1, 4, 7): []mcTriangle{
		{4, 6, 4, 5, 0, 4},
		{1, 5, 1, 3, 0, 1},
		// End of case 4.
		{6, 7, 3, 7, 5, 7},
	},
	newMcIntersections(0, 1, 2, 3): []mcTriangle{
		{0, 4, 1, 5, 3, 7},
		{0, 4, 3, 7, 2, 6},
	},
	newMcIntersections(0, 2, 3, 6): []mcTriangle{
		{0, 1, 4, 6, 0, 4},
		{0, 1, 6, 7, 4, 6},
		{0, 1, 1, 3, 6, 7},
		{1, 3, 3, 7, 6, 7},
	},
	newMcIntersections(1, 2, 5, 6): []mcTriangle{
		{0, 2, 2, 3, 6, 7},
		{0, 2, 6, 7, 4, 6},
		{0, 1, 4, 5, 5, 7},
		{5, 7, 1, 3, 0, 1},
	},
	newMcIntersections(0, 2, 3, 7): []mcTriangle{
		{0, 4, 0, 1, 2, 6},
		{0, 1, 5, 7, 2, 6},
		{2, 6, 5, 7, 6, 7},
		{0, 1, 1, 3, 5, 7},
	},

	// Case 12-17
	newMcIntersections(1, 2, 3, 4): []mcTriangle{
		{0, 1, 1, 5, 0, 2},
		{0, 2, 1, 5, 2, 6},
		{2, 6, 1, 5, 3, 7},
		{4, 5, 0, 4, 4, 6},
	},
	newMcIntersections(1, 2, 4, 7): []mcTriangle{
		{0, 1, 1, 5, 1, 3},
		{0, 2, 2, 3, 2, 6},
		{4, 5, 0, 4, 4, 6},
		{5, 7, 6, 7, 3, 7},
	},
	newMcIntersections(1, 2, 3, 6): []mcTriangle{
		{0, 2, 0, 1, 4, 6},
		{0, 1, 3, 7, 4, 6},
		{0, 1, 1, 5, 3, 7},
		{4, 6, 3, 7, 6, 7},
	},
	newMcIntersections(0, 2, 3, 5, 6): []mcTriangle{
		// Case 9
		{0, 1, 4, 6, 0, 4},
		{0, 1, 6, 7, 4, 6},
		{0, 1, 1, 3, 6, 7},
		{1, 3, 3, 7, 6, 7},
		// End of case 3
		{5, 7, 1, 5, 4, 5},
	},
	newMcIntersections(2, 3, 4, 5, 6): []mcTriangle{
		{5, 7, 1, 5, 0, 4},
		{0, 4, 6, 7, 5, 7},
		{0, 2, 6, 7, 0, 4},
		{0, 2, 3, 7, 6, 7},
		{0, 2, 1, 3, 3, 7},
	},
	newMcIntersections(0, 4, 5, 6, 7): []mcTriangle{
		{1, 5, 0, 1, 0, 2},
		{0, 2, 2, 6, 1, 5},
		{1, 5, 2, 6, 3, 7},
	},

	// Case 18-22
	newMcIntersections(1, 2, 3, 4, 5, 6): []mcTriangle{
		// Inverse of case 4.
		{0, 2, 0, 1, 0, 4},
		{3, 7, 6, 7, 5, 7},
	},
	newMcIntersections(1, 2, 3, 4, 6, 7): []mcTriangle{
		{0, 2, 4, 5, 0, 4},
		{0, 2, 5, 7, 4, 5},
		{0, 2, 1, 5, 5, 7},
		{0, 1, 1, 5, 0, 2},
	},
	newMcIntersections(2, 3, 4, 5, 6, 7): []mcTriangle{
		// Inverse of case 2.
		{1, 5, 0, 4, 0, 2},
		{1, 3, 1, 5, 0, 2},
	},
	newMcIntersections(1, 2, 3, 4, 5, 6, 7): []mcTriangle{
		// Inverse of case 1.
		{0, 2, 0, 1, 0, 4},
	},
	newMcIntersections(0, 1, 2, 3, 4, 5, 6, 7): []mcTriangle{},
}
