package toolbox3d

import (
	"math"
	"math/rand"
	"runtime"

	"github.com/unixpickle/essentials"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

// A HeightMap maps a 2D grid of points to non-negative Z
// values.
//
// The HeightMap can be updated by adding hemispheres and
// other HeightMaps.
// These operations are union operators, in that they
// never reduce the height at any given grid point.
//
// The HeightMap automatically performs interpolation for
// reads to provide the appearance of a continuous curve.
type HeightMap struct {
	// 2D boundaries of the grid.
	Min model2d.Coord
	Max model2d.Coord

	// Spacing of the grid elements.
	Delta float64

	// Row-major data storing the squared heights at every
	// grid element.
	Rows int
	Cols int
	Data []float64
}

// NewHeightMap fills a rectangular 2D region with a
// height map that starts out at zero height.
//
// The maxSize argument limits the number of rows and
// columns, and will be the greater of the two dimensions
// in the data grid.
func NewHeightMap(min, max model2d.Coord, maxSize int) *HeightMap {
	size := max.Sub(min)
	delta := math.Max(size.X, size.Y) / float64(maxSize-1)
	numRows := int(math.Round(size.Y/delta) + 1)
	numCols := int(math.Round(size.X/delta) + 1)
	return &HeightMap{
		Min:   min,
		Max:   max,
		Delta: delta,
		Rows:  numRows,
		Cols:  numCols,
		Data:  make([]float64, numRows*numCols),
	}
}

// Copy creates a deep copy of h.
func (h *HeightMap) Copy() *HeightMap {
	res := *h
	res.Data = append([]float64{}, h.Data...)
	return &res
}

// MaxHeight gets the maximum height at any cell in the
// height map.
func (h *HeightMap) MaxHeight() float64 {
	var maxHeight float64
	for _, h := range h.Data {
		if h > maxHeight {
			maxHeight = h
		}
	}
	return math.Sqrt(maxHeight)
}

// AddSphere adds a hemisphere to the height map, updating
// any cells that were lower than the corresponding point
// on the hemisphere.
//
// Returns true if the sphere changed the height map in
// any way, or false if the sphere was already covered.
func (h *HeightMap) AddSphere(center model2d.Coord, radius float64) bool {
	minRow, minCol := h.coordToIndex(center.Sub(model2d.XY(radius, radius)))
	maxRow, maxCol := h.coordToIndex(center.Add(model2d.XY(radius, radius)))
	r2 := radius * radius
	covered := false
	for row := minRow; row <= maxRow+1; row++ {
		for col := minCol; col <= maxCol+1; col++ {
			c := h.indexToCoord(row, col)
			d2 := (c.X-center.X)*(c.X-center.X) + (c.Y-center.Y)*(c.Y-center.Y)
			if d2 < r2 {
				if h.updateAt(row, col, r2-d2) {
					covered = true
				}
			}
		}
	}
	return covered
}

// AddSphereFill fills a circle with spheres of a bounded
// radius.
//
// The sphereRadius argument determines the maximum radius
// for a sphere.
// The radius argument determines the radius of the circle
// to fill with spheres.
//
// Returns true if the spheres changed the height map in
// any way, or false if the spheres were covered.
func (h *HeightMap) AddSphereFill(center model2d.Coord, radius, sphereRadius float64) bool {
	if sphereRadius > radius {
		return h.AddSphere(center, radius)
	}
	minRow, minCol := h.coordToIndex(center.Sub(model2d.XY(radius, radius)))
	maxRow, maxCol := h.coordToIndex(center.Add(model2d.XY(radius, radius)))
	covered := false
	for row := minRow; row <= maxRow+1; row++ {
		for col := minCol; col <= maxCol+1; col++ {
			c := h.indexToCoord(row, col)
			dist := c.Dist(center)
			if radius-dist >= sphereRadius {
				if h.updateAt(row, col, sphereRadius*sphereRadius) {
					covered = true
				}
			} else {
				outset := dist - (radius - sphereRadius)
				if outset < sphereRadius {
					remaining := sphereRadius*sphereRadius - outset*outset
					if h.updateAt(row, col, remaining) {
						covered = true
					}
				}
			}
		}
	}
	return covered
}

// AddSpheresSDF fills a 2D signed distance function with
// spheres that touch the edges of the SDF. This creates a
// smooth, 3D version of the 2D model.
//
// The numSpheres argument specifies the number of spheres
// to sample inside the shape.
//
// The eps argument is a small value used to determine the
// medial axis.
// Smaller values are more sensitive to jagged edges of
// the collider.
// See model2d.ProjectMedialAxis().
//
// The maxRadius argument, if non-zero, is used to limit
// the height of the resulting object.
// See HeightMap.AddSphereFill().
func (h *HeightMap) AddSpheresSDF(p model2d.PointSDF, numSpheres int, eps, maxRadius float64) {
	essentials.StatefulConcurrentMap(runtime.GOMAXPROCS(0), numSpheres, func() func(int) {
		rng := rand.New(rand.NewSource(rand.Int63()))
		min := p.Min()
		max := p.Max()
		return func(_ int) {
			c := model2d.XY(rng.Float64(), rng.Float64())
			c = min.Add(c.Mul(max.Sub(min)))
			proj := model2d.ProjectMedialAxis(p, c, 0, eps)
			radius := p.SDF(proj)
			if maxRadius != 0 {
				h.AddSphereFill(proj, radius, maxRadius)
			} else {
				h.AddSphere(proj, radius)
			}
		}
	})
}

func (h *HeightMap) updateAt(row, col int, height float64) bool {
	if row < 0 || col < 0 || row >= h.Rows || col >= h.Cols {
		return false
	}
	idx := row*h.Cols + col
	if h.Data[idx] < height {
		h.Data[idx] = height
		return true
	}
	return false
}

// AddHeightMap writes the union of h and h1 to h.
//
// This is optimized for the case when h and h1 are laid
// out exactly the same, with the same grid spacing and
// boundaries.
//
// One use case for this API is to combine multiple height
// maps that were generated on different Goroutines.
//
// Returns true if h1 modified h, or false otherwise.
func (h *HeightMap) AddHeightMap(h1 *HeightMap) bool {
	if h.Rows == h1.Rows && h.Cols == h1.Cols && h.Min == h1.Min && h.Max == h1.Max {
		var changed bool
		for i, x := range h1.Data {
			if x > h.Data[i] {
				changed = true
				h.Data[i] = x
			}
		}
		return changed
	}
	var changed bool
	var idx int
	for y := 0; y < h.Rows; y++ {
		for x := 0; x < h.Cols; x++ {
			c := h.indexToCoord(y, x)
			height := h1.HeightSquaredAt(c)
			if h.Data[idx] < height {
				h.Data[idx] = height
				changed = true
			}
			idx++
		}
	}
	return changed
}

// HigherAt checks if the height map is higher than a
// given height at the given 2D coordinate.
// Returns true if the height map is higher.
//
// The coordinate may be out of bounds.
func (h *HeightMap) HigherAt(c model2d.Coord, height float64) bool {
	return h.HeightSquaredAt(c) > height*height
}

// HeightSquaredAt gets the interpolated square of the
// height at any coordinate.
//
// The coordinate may be out of bounds.
func (h *HeightMap) HeightSquaredAt(c model2d.Coord) float64 {
	rowMin, colMin := h.coordToIndex(c)
	c1 := h.indexToCoord(rowMin, colMin)

	colFrac := (c.X - c1.X) / h.Delta
	rowFrac := (c.Y - c1.Y) / h.Delta

	h11 := h.getAt(rowMin, colMin)
	h21 := h.getAt(rowMin+1, colMin)
	h12 := h.getAt(rowMin, colMin+1)
	h22 := h.getAt(rowMin+1, colMin+1)

	return (1-rowFrac)*(1-colFrac)*h11 +
		rowFrac*(1-colFrac)*h21 +
		(1-rowFrac)*colFrac*h12 +
		rowFrac*colFrac*h22
}

// SetHeightSquaredAt sets the squared height at the index
// closest to the given coordinate.
func (h *HeightMap) SetHeightSquaredAt(c model2d.Coord, hs float64) {
	if hs < 0 {
		panic("squared value must be non-negative")
	}
	row, col := h.coordToIndex(c)
	c1 := h.indexToCoord(row, col)

	col += int(math.Round((c.X - c1.X) / h.Delta))
	row += int(math.Round((c.Y - c1.Y) / h.Delta))

	if row < 0 || col < 0 || row >= h.Rows || col >= h.Cols {
		return
	}
	h.Data[row*h.Cols+col] = hs
}

// Mesh generates a solid mesh containing the volume under
// the height map but above the Z axis.
func (h *HeightMap) Mesh() *model3d.Mesh {
	// Create a result mesh without a vertex-to-face cache.
	mesh := model3d.NewMesh()
	mesh.AddMesh(h.surfaceMesh())

	edges := findUnsharedEdges(mesh)

	// Create base triangles.
	mesh.Iterate(func(t *model3d.Triangle) {
		mesh.Add(&model3d.Triangle{
			model3d.XY(t[0].X, t[0].Y),
			model3d.XY(t[2].X, t[2].Y),
			model3d.XY(t[1].X, t[1].Y),
		})
	})

	// Connect edges to base.
	for edge := range edges {
		p1, p2 := edge[0], edge[1]
		mesh.AddQuad(
			p2,
			p1,
			model3d.XY(p1.X, p1.Y),
			model3d.XY(p2.X, p2.Y),
		)
	}

	return mesh
}

// Mesh generates a mesh for the surface by reflecting it
// across the Z axis. This is like Mesh(), but with a
// symmetrical base rather than a flat one.
func (h *HeightMap) MeshBidir() *model3d.Mesh {
	mesh := model3d.NewMesh()
	mesh.AddMesh(h.surfaceMesh())

	edges := findUnsharedEdges(mesh)

	// Create base triangles.
	mesh.Iterate(func(t *model3d.Triangle) {
		mesh.Add(&model3d.Triangle{
			model3d.XYZ(t[0].X, t[0].Y, -t[0].Z),
			model3d.XYZ(t[2].X, t[2].Y, -t[2].Z),
			model3d.XYZ(t[1].X, t[1].Y, -t[1].Z),
		})
	})

	for edge := range edges {
		p1, p2 := edge[0], edge[1]
		mesh.AddQuad(
			p2,
			p1,
			model3d.XYZ(p1.X, p1.Y, -p1.Z),
			model3d.XYZ(p2.X, p2.Y, -p2.Z),
		)
	}

	return mesh
}

func (h *HeightMap) surfaceMesh() *model3d.Mesh {
	// By default, we keep all zero points slightly above
	// z=0 to prevent singularities.
	minZ2 := math.Pow(h.Delta*1e-5, 2)
	for _, d := range h.Data {
		if d != 0 && d < minZ2 {
			minZ2 = d
		}
	}
	minZ := math.Sqrt(minZ2)

	mesh := model3d.NewMesh()
	for row := -1; row < h.Rows; row++ {
		for col := -1; col < h.Cols; col++ {
			sqHeights := [4]float64{
				h.getAt(row, col),
				h.getAt(row, col+1),
				h.getAt(row+1, col),
				h.getAt(row+1, col+1),
			}
			coords := [4]model2d.Coord{
				h.indexToCoord(row, col),
				h.indexToCoord(row, col+1),
				h.indexToCoord(row+1, col),
				h.indexToCoord(row+1, col+1),
			}
			surface := [4]model3d.Coord3D{}

			for i, sqHeight := range sqHeights {
				surface[i] = model3d.XYZ(
					coords[i].X,
					coords[i].Y,
					math.Sqrt(sqHeight),
				)
			}
			surfaceTris := triangulateQuad(surface)
			for _, t := range surfaceTris {
				if t[0].Z != 0 || t[1].Z != 0 || t[2].Z != 0 {
					for i, c := range t {
						if c.Z == 0 {
							c.Z = minZ
							t[i] = c
						}
					}
					mesh.Add(t)
				}
			}
		}
	}

	separateSingularVertices(mesh)
	return mesh
}

func (h *HeightMap) getAt(row, col int) float64 {
	if row < 0 || col < 0 || row >= h.Rows || col >= h.Cols {
		return 0
	}
	return h.Data[row*h.Cols+col]
}

func (h *HeightMap) indexToCoord(row, col int) model2d.Coord {
	res := h.Min
	res.X += float64(col) * h.Delta
	res.Y += float64(row) * h.Delta
	return res
}

func (h *HeightMap) coordToIndex(c model2d.Coord) (row, col int) {
	row = int((c.Y - h.Min.Y) / h.Delta)
	col = int((c.X - h.Min.X) / h.Delta)
	return
}

type heightMapSolid struct {
	heightMap *HeightMap
	maxHeight float64
	minHeight float64
}

// HeightMapToSolid creates a 3D solid representing the
// volume under a height map and above the Z plane.
func HeightMapToSolid(hm *HeightMap) model3d.Solid {
	return &heightMapSolid{
		heightMap: hm,
		maxHeight: hm.MaxHeight(),
		minHeight: 0,
	}
}

// HeightMapToSolidBidir is like HeightMapToSolid, but it
// mirrors the solid across the Z plane to make it
// symmetric.
func HeightMapToSolidBidir(hm *HeightMap) model3d.Solid {
	return &heightMapSolid{
		heightMap: hm,
		maxHeight: hm.MaxHeight(),
		minHeight: -hm.MaxHeight(),
	}
}

func (h *heightMapSolid) Min() model3d.Coord3D {
	return model3d.XYZ(h.heightMap.Min.X, h.heightMap.Min.Y, h.minHeight)
}

func (h *heightMapSolid) Max() model3d.Coord3D {
	return model3d.XYZ(h.heightMap.Max.X, h.heightMap.Max.Y, h.maxHeight)
}

func (h *heightMapSolid) Contains(c model3d.Coord3D) bool {
	return model3d.InBounds(h, c) && h.heightMap.HigherAt(c.XY(), math.Abs(c.Z))
}

func separateSingularVertices(m *model3d.Mesh) {
	for _, v := range m.SingularVertices() {
		for _, family := range singularVertexFamilies(m, v) {
			min, max := family[0].Min(), family[0].Max()
			for _, t := range family {
				min = min.Min(t.Min())
				max = max.Max(t.Max())
			}
			center := min.Mid(max)
			for _, t := range family {
				m.Remove(t)
				for i, c := range t {
					if c == v {
						t[i].X = (c.X * 0.99) + (center.X * 0.01)
						t[i].Y = (c.Y * 0.99) + (center.Y * 0.01)
					}
				}
				m.Add(t)
			}
		}
	}
}

func findUnsharedEdges(m *model3d.Mesh) map[[2]model3d.Coord3D]bool {
	edges := map[[2]model3d.Coord3D]bool{}
	m.Iterate(func(t *model3d.Triangle) {
		for i := 0; i < 3; i++ {
			p1 := t[i]
			p2 := t[(i+1)%3]
			edge := [2]model3d.Coord3D{p1, p2}
			otherEdge := [2]model3d.Coord3D{p2, p1}
			if edges[otherEdge] {
				delete(edges, otherEdge)
			} else {
				edges[edge] = true
			}
		}
	})
	return edges
}

func triangulateQuad(surface [4]model3d.Coord3D) [2]*model3d.Triangle {
	// In the future, we may be able to use heuristics here to
	// eliminate flat triangles when possible.
	return [2]*model3d.Triangle{
		{surface[0], surface[1], surface[2]},
		{surface[2], surface[1], surface[3]},
	}
}
