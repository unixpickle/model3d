package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

// A HeightMap maps a 2D grid of points to Z values.
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
// columns.
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

// AddSphere adds a hemisphere to the height map.
// where the hemisphere is on the Z plane.
func (h *HeightMap) AddSphere(center model2d.Coord, radius float64) {
	minRow, minCol := h.coordToIndex(center.Sub(model2d.XY(radius, radius)))
	maxRow, maxCol := h.coordToIndex(center.Add(model2d.XY(radius, radius)))
	r2 := radius * radius
	for row := minRow; row <= maxRow+1; row++ {
		for col := minCol; col <= maxCol+1; col++ {
			c := h.indexToCoord(row, col)
			d2 := (c.X-center.X)*(c.X-center.X) + (c.Y-center.Y)*(c.Y-center.Y)
			if d2 < r2 {
				h.updateAt(row, col, r2-d2)
			}
		}
	}
}

// AddHeightMap writes the union of h and h1 to h.
//
// Both h and h1 must be laid out exactly the same.
func (h *HeightMap) AddHeightMap(h1 *HeightMap) {
	if h.Rows != h1.Rows || h.Cols != h1.Cols || h.Min != h1.Min || h.Max != h1.Max {
		panic("incompatible height maps")
	}
	for i, x := range h1.Data {
		h.Data[i] = math.Max(h.Data[i], x)
	}
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

func (h *HeightMap) updateAt(row, col int, height float64) {
	if row < 0 || col < 0 || row >= h.Rows || col >= h.Cols {
		return
	}
	idx := row*h.Cols + col
	if h.Data[idx] < height {
		h.Data[idx] = height
	}
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

// A HeightMapSolid is a 3D solid representing the volume
// under a height map and above the Z plane.
type HeightMapSolid struct {
	HeightMap *HeightMap
	MaxHeight float64
}

func NewHeightMapSolid(hm *HeightMap) *HeightMapSolid {
	return &HeightMapSolid{
		HeightMap: hm,
		MaxHeight: hm.MaxHeight(),
	}
}

func (h *HeightMapSolid) Min() model3d.Coord3D {
	return model3d.XY(h.HeightMap.Min.X, h.HeightMap.Min.Y)
}

func (h *HeightMapSolid) Max() model3d.Coord3D {
	return model3d.XYZ(h.HeightMap.Max.X, h.HeightMap.Max.Y, h.MaxHeight)
}

func (h *HeightMapSolid) Contains(c model3d.Coord3D) bool {
	return model3d.InBounds(h, c) && h.HeightMap.HigherAt(c.XY(), c.Z)
}
