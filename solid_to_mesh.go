package model3d

import (
	"github.com/unixpickle/essentials"
)

// SolidToMesh approximates the solid s as a triangle mesh
// by blurring the result of a RectScanner.
//
// The delta argument specifies the initial spacing
// between sampled cubes, and subdivisions indicates the
// maximum number of times these cubes can be cut in half.
//
// The blurFrac argument specifies how much each vertex is
// moved towards its neighbors, between 0 and 1.
// The blurIters argument specifies how many times the
// resulting mesh is blurred before being returned.
func SolidToMesh(s Solid, delta float64, subdivisions int, blurFrac float64, blurIters int) *Mesh {
	if delta == 0 {
		panic("invalid delta argument")
	}
	var mesh *Mesh
	if subdivisions == 0 {
		mesh = directSolidToMesh(s, delta)
	} else {
		scanner := NewRectScanner(s, delta)
		for i := 0; i < subdivisions; i++ {
			scanner.Subdivide()
		}
		mesh = scanner.Mesh()
	}
	if blurIters == 0 {
		return mesh
	}
	blurRates := make([]float64, blurIters)
	for i := range blurRates {
		blurRates[i] = blurFrac
	}
	mesh = mesh.Blur(blurRates...)
	return mesh
}

func directSolidToMesh(s Solid, delta float64) *Mesh {
	spacer := newSquareSpacer(s, delta)
	cache := newSolidCache(s, spacer)

	return faceQuadsToMesh(func(f func([4]Coord3D)) {
		spacer.IterateSquares(func(x, y, z int) {
			min := spacer.CornerCoord(x, y, z)
			max := spacer.CornerCoord(x+1, y+1, z+1)

			numInterior := cache.NumInteriorCorners(x, y, z)
			if numInterior == 0 {
				return
			} else if numInterior == 8 {
				if x == 0 || x == len(spacer.Xs)-2 || y == 0 || y == len(spacer.Ys)-2 ||
					z == 0 || z == len(spacer.Zs)-2 {
					panic("solid is true outside of bounds")
				}
				return
			}

			isSideBorder := func(axis int, positive bool) bool {
				x1, y1, z1 := x, y, z
				delta := 1
				if !positive {
					delta = -1
				}
				if axis == 0 {
					x1 += delta
					if x1 < 0 || x1+1 >= len(spacer.Xs) {
						return true
					}
				} else if axis == 1 {
					y1 += delta
					if y1 < 0 || y1+1 >= len(spacer.Ys) {
						return true
					}
				} else {
					z1 += delta
					if z1 < 0 || z1+1 >= len(spacer.Zs) {
						return true
					}
				}
				return cache.NumInteriorCorners(x1, y1, z1) == 0
			}
			boxToFaceQuads(min, max, isSideBorder, f)
		})
	})
}

// A RectScanner maps out the edges of a solid using
// axis-aligned cubes.
type RectScanner struct {
	border map[*rectPiece]bool
	solid  Solid
}

// NewRectScanner creates a RectScanner by uniformly
// scanning the solid with a spacing of delta units.
func NewRectScanner(s Solid, delta float64) *RectScanner {
	spacer := newSquareSpacer(s, delta)
	cache := newSolidCache(s, spacer)

	pieces := map[int]*rectPiece{}
	res := &RectScanner{
		border: map[*rectPiece]bool{},
		solid:  s,
	}

	// First, create all border pieces so that we can
	// create all the empty and locked pieces next to them
	// without creating unneeded ones.
	spacer.IterateSquares(func(x, y, z int) {
		piece := &rectPiece{
			Min: spacer.CornerCoord(x, y, z),
			Max: spacer.CornerCoord(x+1, y+1, z+1),

			NumInteriorCorners: cache.NumInteriorCorners(x, y, z),
		}
		if piece.NumInteriorCorners != 0 && piece.NumInteriorCorners != 8 {
			piece.Neighbors = []*rectPiece{}
			pieces[spacer.SquareIndex(x, y, z)] = piece
			res.border[piece] = true
		} else if piece.NumInteriorCorners == 8 {
			if x == 0 || x == len(spacer.Xs)-2 || y == 0 || y == len(spacer.Ys)-2 ||
				z == 0 || z == len(spacer.Zs)-2 {
				panic("solid is true outside of bounds")
			}
		}
	})

	// Create all neighbors of the border pieces while
	// discarding pieces with no border neighbors.
	// This can save considerable amounts of memory.
	spacer.IterateSquares(func(x, y, z int) {
		var piece *rectPiece
		if p, ok := pieces[spacer.SquareIndex(x, y, z)]; ok {
			piece = p
		} else {
			piece = &rectPiece{
				Min: spacer.CornerCoord(x, y, z),
				Max: spacer.CornerCoord(x+1, y+1, z+1),

				NumInteriorCorners: cache.NumInteriorCorners(x, y, z),
			}
			if piece.NumInteriorCorners == 0 {
				piece.Deleted = true
			} else if piece.NumInteriorCorners == 8 {
				piece.Locked = true
			}
		}
		addNeighbor := func(x, y, z int) {
			if p1, ok := pieces[spacer.SquareIndex(x, y, z)]; ok {
				p1.AddNeighbor(piece)
			}
		}
		if x > 0 {
			addNeighbor(x-1, y, z)
		}
		if x+2 < len(spacer.Xs) {
			addNeighbor(x+1, y, z)
		}
		if y > 0 {
			addNeighbor(x, y-1, z)
		}
		if y+2 < len(spacer.Ys) {
			addNeighbor(x, y+1, z)
		}
		if z > 0 {
			addNeighbor(x, y, z-1)
		}
		if z+2 < len(spacer.Zs) {
			addNeighbor(x, y, z+1)
		}
	})

	return res
}

// Subdivide doubles the resolution along the border of
// the solid.
func (r *RectScanner) Subdivide() {
	pieces := make([]*rectPiece, 0, len(r.border))
	for p := range r.border {
		pieces = append(pieces, p)
	}
	for _, p := range pieces {
		r.splitBorder(p)
	}
}

// BorderRects calls f with every rectangle on the outside
// of the border.
//
// Each rectangle is passed in counter-clockwise order, so
// using the right-hand rule will yield normals facing the
// inside of the solid.
func (r *RectScanner) BorderRects(f func(points [4]Coord3D)) {
	for p := range r.border {
		boxToFaceQuads(p.Min, p.Max, p.IsSideBorder, f)
	}
}

// Mesh creates a mesh for the border.
func (r *RectScanner) Mesh() *Mesh {
	return faceQuadsToMesh(r.BorderRects)
}

func (r *RectScanner) splitBorder(rp *rectPiece) {
	delete(r.border, rp)
	for _, n := range rp.Neighbors {
		n.RemoveNeighbor(rp)
	}

	var newPieces []*rectPiece

	mid := rp.Min.Mid(rp.Max)
	for xIdx := 0; xIdx < 2; xIdx++ {
		minX := rp.Min.X
		maxX := rp.Max.X
		if xIdx == 0 {
			maxX = mid.X
		} else {
			minX = mid.X
		}
		for yIdx := 0; yIdx < 2; yIdx++ {
			minY := rp.Min.Y
			maxY := rp.Max.Y
			if yIdx == 0 {
				maxY = mid.Y
			} else {
				minY = mid.Y
			}
			for zIdx := 0; zIdx < 2; zIdx++ {
				minZ := rp.Min.Z
				maxZ := rp.Max.Z
				if zIdx == 0 {
					maxZ = mid.Z
				} else {
					minZ = mid.Z
				}

				newPiece := &rectPiece{
					Min:       Coord3D{X: minX, Y: minY, Z: minZ},
					Max:       Coord3D{X: maxX, Y: maxY, Z: maxZ},
					Neighbors: []*rectPiece{},
				}
				newPiece.CountInteriorCorners(r.solid)
				newPiece.UpdateNeighbors(rp.Neighbors)
				rp.AddNeighbor(newPiece)
				newPieces = append(newPieces, newPiece)
			}
		}
	}

	for _, p := range newPieces {
		if p.NumInteriorCorners == 0 {
			if p.TouchingLocked() {
				r.border[p] = true
			} else {
				p.Neighbors = nil
				p.Deleted = true
			}
		} else if p.NumInteriorCorners == 8 {
			if p.TouchingDeleted() {
				r.border[p] = true
			} else {
				p.Neighbors = nil
				p.Locked = true
			}
		} else {
			r.border[p] = true
		}
	}
}

type rectPiece struct {
	Min Coord3D
	Max Coord3D

	// A set of adjacent pieces.
	//
	// May be nil for locked or deleted pieces.
	Neighbors []*rectPiece

	// The number of corners inside the solid.
	NumInteriorCorners int

	// If true, this piece is definitely inside the solid
	// and is not allowed to be on the border.
	// It will not be subdivided any more, and no pieces
	// touching it may be deleted.
	Locked bool

	// If true, this piece is definitely outside the
	// solid.
	// Therefore, no pieces touching it may be locked.
	Deleted bool
}

func (r *rectPiece) CheckNeighbor(r1 *rectPiece) bool {
	for i := 0; i < 3; i++ {
		i1 := (i + 1) % 3
		i2 := (i + 2) % 3
		if r.Min.Array()[i1] >= r1.Max.Array()[i1] ||
			r.Min.Array()[i2] >= r1.Max.Array()[i2] ||
			r.Max.Array()[i1] <= r1.Min.Array()[i1] ||
			r.Max.Array()[i2] <= r1.Min.Array()[i2] {
			continue
		}
		if r.Min.Array()[i] == r1.Max.Array()[i] {
			return true
		} else if r.Max.Array()[i] == r1.Min.Array()[i] {
			return true
		}
	}
	return false
}

func (r *rectPiece) CountInteriorCorners(s Solid) {
	for _, x := range []float64{r.Min.X, r.Max.X} {
		for _, y := range []float64{r.Min.Y, r.Max.Y} {
			for _, z := range []float64{r.Min.Z, r.Max.Z} {
				if s.Contains(Coord3D{X: x, Y: y, Z: z}) {
					r.NumInteriorCorners++
				}
			}
		}
	}
}

func (r *rectPiece) UpdateNeighbors(possible []*rectPiece) {
	for _, n := range possible {
		if n.CheckNeighbor(r) {
			if r.Neighbors != nil {
				r.AddNeighbor(n)
			}
			if n.Neighbors != nil {
				n.AddNeighbor(r)
			}
		}
	}
}

func (r *rectPiece) AddNeighbor(r1 *rectPiece) {
	r.Neighbors = append(r.Neighbors, r1)
}

func (r *rectPiece) RemoveNeighbor(r1 *rectPiece) {
	for i, n := range r.Neighbors {
		if n == r1 {
			last := len(r.Neighbors) - 1
			r.Neighbors[i] = r.Neighbors[last]
			r.Neighbors[last] = nil
			r.Neighbors = r.Neighbors[:last]
			return
		}
	}
}

func (r *rectPiece) TouchingLocked() bool {
	for _, n := range r.Neighbors {
		if n.Locked {
			return true
		}
	}
	return false
}

func (r *rectPiece) TouchingDeleted() bool {
	for _, n := range r.Neighbors {
		if n.Deleted {
			return true
		}
	}
	return false
}

func (r *rectPiece) IsSideBorder(axis int, max bool) bool {
	for _, n := range r.Neighbors {
		if n.Deleted {
			continue
		}
		if max {
			if n.Min.Array()[axis] == r.Max.Array()[axis] {
				return false
			}
		} else {
			if n.Max.Array()[axis] == r.Min.Array()[axis] {
				return false
			}
		}
	}
	return true
}

type squareSpacer struct {
	Xs []float64
	Ys []float64
	Zs []float64
}

func newSquareSpacer(s Solid, delta float64) *squareSpacer {
	var xs, ys, zs []float64
	min := s.Min()
	max := s.Max()
	for x := min.X - delta; x <= max.X+delta; x += delta {
		xs = append(xs, x)
	}
	for y := min.Y - delta; y <= max.Y+delta; y += delta {
		ys = append(ys, y)
	}
	for z := min.Z - delta; z <= max.Z+delta; z += delta {
		zs = append(zs, z)
	}
	return &squareSpacer{Xs: xs, Ys: ys, Zs: zs}
}

func (s *squareSpacer) IterateSquares(f func(x, y, z int)) {
	for z := 0; z < len(s.Zs)-1; z++ {
		for y := 0; y < len(s.Ys)-1; y++ {
			for x := 0; x < len(s.Xs)-1; x++ {
				f(x, y, z)
			}
		}
	}
}

func (s *squareSpacer) NumSquares() int {
	return (len(s.Xs) - 1) * (len(s.Ys) - 1) * (len(s.Zs) - 1)
}

func (s *squareSpacer) SquareIndex(x, y, z int) int {
	return x + y*(len(s.Xs)-1) + z*(len(s.Xs)-1)*(len(s.Ys)-1)
}

func (s *squareSpacer) CornerCoord(x, y, z int) Coord3D {
	return Coord3D{X: s.Xs[x], Y: s.Ys[y], Z: s.Zs[z]}
}

func (s *squareSpacer) IterateCorners(f func(x, y, z int)) {
	for z := range s.Zs {
		for y := range s.Ys {
			for x := range s.Xs {
				f(x, y, z)
			}
		}
	}
}

func (s *squareSpacer) NumCorners() int {
	return len(s.Xs) * len(s.Ys) * len(s.Zs)
}

func (s *squareSpacer) CornerIndex(x, y, z int) int {
	return x + y*len(s.Xs) + z*len(s.Xs)*len(s.Ys)
}

func (s *squareSpacer) IndexToCorner(idx int) (int, int, int) {
	x := idx % len(s.Xs)
	idx /= len(s.Xs)
	y := idx % len(s.Ys)
	z := idx / len(s.Ys)
	return x, y, z
}

type solidCache struct {
	spacer *squareSpacer
	solid  Solid

	startZ  int
	strideZ int
	cachedZ int
	values  []bool
}

func newSolidCache(solid Solid, spacer *squareSpacer) *solidCache {
	cachedZ := essentials.MinInt(len(spacer.Zs), 10)
	strideZ := len(spacer.Xs) * len(spacer.Ys)
	cache := &solidCache{
		spacer:  spacer,
		solid:   solid,
		strideZ: strideZ,
		cachedZ: cachedZ,
		values:  make([]bool, strideZ*cachedZ),
	}
	cache.fillTailValues(cachedZ)
	return cache
}

func (s *solidCache) NumInteriorCorners(x, y, z int) int {
	var res int
	for k := z; k < z+2; k++ {
		for j := y; j < y+2; j++ {
			for i := x; i < x+2; i++ {
				if s.cornerValue(i, j, k) {
					res++
				}
			}
		}
	}
	return res
}

func (s *solidCache) cornerValue(x, y, z int) bool {
	if z >= s.startZ && z < s.startZ+s.cachedZ {
		return s.values[s.spacer.CornerIndex(x, y, z-s.startZ)]
	}

	newStart := essentials.MinInt(essentials.MaxInt(0, z-s.cachedZ/2),
		len(s.spacer.Zs)-s.cachedZ)
	shift := newStart - s.startZ
	s.startZ = newStart
	if shift < 0 || shift >= s.cachedZ {
		// Start the cache all over again.
		s.fillTailValues(s.cachedZ)
	} else {
		copy(s.values, s.values[shift*s.strideZ:])
		s.fillTailValues(shift)
	}

	return s.cornerValue(x, y, z)
}

func (s *solidCache) fillTailValues(numTail int) {
	idx := (s.cachedZ - numTail) * s.strideZ
	for _, z := range s.spacer.Zs[s.startZ+s.cachedZ-numTail:][:numTail] {
		for _, y := range s.spacer.Ys {
			for _, x := range s.spacer.Xs {
				s.values[idx] = s.solid.Contains(Coord3D{X: x, Y: y, Z: z})
				idx++
			}
		}
	}
}

func boxToFaceQuads(min, max Coord3D, isSideBorder func(axis int, positive bool) bool,
	f func([4]Coord3D)) {
	// Left and right sides.
	if isSideBorder(0, false) {
		p1, p2, p3 := min, min, min
		p1.Y = max.Y
		p2.Y = max.Y
		p2.Z = max.Z
		p3.Z = max.Z
		f([4]Coord3D{min, p1, p2, p3})
	}
	if isSideBorder(0, true) {
		p1, p2, p3 := max, max, max
		p1.Z = min.Z
		p2.Z = min.Z
		p2.Y = min.Y
		p3.Y = min.Y
		f([4]Coord3D{max, p1, p2, p3})
	}

	// Top and bottom sides.
	if isSideBorder(1, false) {
		p1, p2, p3 := min, min, min
		p1.Z = max.Z
		p2.Z = max.Z
		p2.X = max.X
		p3.X = max.X
		f([4]Coord3D{min, p1, p2, p3})
	}
	if isSideBorder(1, true) {
		p1, p2, p3 := max, max, max
		p1.X = min.X
		p2.X = min.X
		p2.Z = min.Z
		p3.Z = min.Z
		f([4]Coord3D{max, p1, p2, p3})
	}

	// Front and back sides.
	if isSideBorder(2, false) {
		p1, p2, p3 := min, min, min
		p1.X = max.X
		p2.X = max.X
		p2.Y = max.Y
		p3.Y = max.Y
		f([4]Coord3D{min, p1, p2, p3})
	}
	if isSideBorder(2, true) {
		p1, p2, p3 := max, max, max
		p1.Y = min.Y
		p2.Y = min.Y
		p2.X = min.X
		p3.X = min.X
		f([4]Coord3D{max, p1, p2, p3})
	}
}

func faceQuadsToMesh(rects func(func(points [4]Coord3D))) *Mesh {
	p := newPtrMesh()
	cmap := newPtrCoordMap()
	segmentCounts := map[ptrSegment]int{}
	rects(func(points [4]Coord3D) {
		var ptrs [4]*ptrCoord
		for i, c := range points {
			ptrs[i] = cmap.Coord(c)
		}
		p.Add(newPtrTriangle(ptrs[0], ptrs[2], ptrs[1]))
		p.Add(newPtrTriangle(ptrs[0], ptrs[3], ptrs[2]))
		for i, p1 := range ptrs {
			s := newPtrSegment(p1, ptrs[(i+1)%4])
			segmentCounts[s]++
		}
	})
	fixSingularEdges(p, segmentCounts)
	fixSingularVertices(p)
	return p.Mesh()
}

// fixSingularEdges fixes edges of two touching diagonal
// edge boxes, since these edges belong to four faces at
// once (which is not allowed).
//
// The fix is done by splitting the edge apart and pulling
// the two middle vertices apart, producing singular
// vertices but no singular edges.
//
// Conceptually, singular edges really ought not touch,
// since there is only a singularity because the touching
// vertices are not in the solid.
func fixSingularEdges(p *ptrMesh, counts map[ptrSegment]int) {
	for seg, count := range counts {
		if count == 2 {
			continue
		} else if count == 4 {
			fixSingularEdge(p, seg)
		} else {
			panic("unexpected edge situation")
		}
	}
}

func fixSingularEdge(p *ptrMesh, seg ptrSegment) {
	tris := seg.Triangles()
	if len(tris) != 4 {
		panic("not a singular edge")
	}

	// Find a pair of triangles on the same cube.
	var t1, t2 *ptrTriangle
	t1 = tris[0]
	t1Normal := t1.Triangle().Normal()
	t1Other := seg.Other(t1).Coord3D
	minDot := 0.0
	for _, t := range tris[1:] {
		dir := seg.Other(t).Coord3D.Sub(t1Other)
		dot := dir.Dot(t1Normal)
		if dot < minDot {
			minDot = dot
			t2 = t
		}
	}

	// Find the other cube's triangle pair.
	var t3, t4 *ptrTriangle
	for _, t := range tris[1:] {
		if t != t2 {
			if t3 == nil {
				t3 = t
			} else {
				t4 = t
			}
		}
	}

	fixSingularEdgePair(p, seg, t1, t2)
	fixSingularEdgePair(p, seg, t3, t4)
}

func fixSingularEdgePair(p *ptrMesh, seg ptrSegment, t1, t2 *ptrTriangle) {
	p1 := seg.Other(t1).Coord3D
	p2 := seg.Other(t2).Coord3D

	// Move the segment's midpoint away from the singular
	// edge to make the edges not touch.
	mp := seg.Mid().Scale(0.9).Add(p1.Mid(p2).Scale(0.1))

	// This point is definitely not already in the mesh,
	// since it's off the grid.
	mpPtr := &ptrCoord{Coord3D: mp}

	fixSingularEdgeTriangle(p, seg, mpPtr, t1)
	fixSingularEdgeTriangle(p, seg, mpPtr, t2)
}

func fixSingularEdgeTriangle(p *ptrMesh, seg ptrSegment, mid *ptrCoord, t *ptrTriangle) {
	other := seg.Other(t)
	tNormal := t.Triangle().Normal()
	p.Remove(t)
	t.RemoveCoords()

	t1 := newPtrTriangle(other, seg[0], mid)
	t2 := newPtrTriangle(other, seg[1], mid)

	// Flip the triangles so that, if mp == mid, the
	// orientation/normal would be the same.
	mp := seg.Mid()
	rawT1 := &Triangle{other.Coord3D, seg[0].Coord3D, mp}
	rawT2 := &Triangle{other.Coord3D, seg[1].Coord3D, mp}
	if rawT1.Normal().Dot(tNormal) < 0 {
		t1.Coords[0], t1.Coords[1] = t1.Coords[1], t1.Coords[0]
	}
	if rawT2.Normal().Dot(tNormal) < 0 {
		t2.Coords[0], t2.Coords[1] = t2.Coords[1], t2.Coords[0]
	}

	p.Add(t1)
	p.Add(t2)
}

// fixSingularVertices fixes singular vertices by
// duplicating them and then moving the duplicates
// slightly away from each other.
func fixSingularVertices(p *ptrMesh) {
	vertices := map[*ptrCoord]struct{}{}
	p.Iterate(func(t *ptrTriangle) {
		for _, c := range t.Coords {
			vertices[c] = struct{}{}
		}
	})
	for v := range vertices {
		clusters := v.Clusters()
		if len(clusters) == 1 {
			continue
		}
		for _, cluster := range clusters {
			// Move the vertex closer to the mean of this
			// cluster. Might not work in the general case,
			// but appears to work for the cube-based grids
			// we generate here.
			mean := Coord3D{}
			count := 0.0
			for _, t := range cluster {
				for _, p := range t.Coords {
					count++
					mean = mean.Add(p.Coord3D)
				}
			}
			mean = mean.Scale(1 / count)
			v1 := v.Scale(0.99).Add(mean.Scale(0.01))
			newCoord := &ptrCoord{Coord3D: v1}

			for _, t := range cluster {
				v.RemoveTriangle(t)
				newCoord.Triangles = append(newCoord.Triangles, t)
				for i, c := range t.Coords {
					if c == v {
						t.Coords[i] = newCoord
					}
				}
			}
		}
	}
}
