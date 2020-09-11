package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
)

type HeightMap struct {
	Min   model2d.Coord
	Max   model2d.Coord
	Delta float64

	Rows int
	Cols int
	Data []float64
}

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

func (h *HeightMap) MaxHeight() float64 {
	var maxHeight float64
	for _, h := range h.Data {
		if h > maxHeight {
			maxHeight = h
		}
	}
	return math.Sqrt(maxHeight)
}

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

func (h *HeightMap) AddHeightMap(h1 *HeightMap) {
	if h.Rows != h1.Rows || h.Cols != h1.Cols || h.Min != h1.Min || h.Max != h1.Max {
		panic("incompatible height maps")
	}
	for i, x := range h1.Data {
		h.Data[i] = math.Max(h.Data[i], x)
	}
}

func (h *HeightMap) GetHeightSquared(c model2d.Coord) float64 {
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
