package model2d

import (
	"bytes"
	"fmt"

	"github.com/unixpickle/model3d/fileformats"
)

// EncodeCSV encodes the mesh as a CSV file.
func EncodeCSV(m *Mesh) []byte {
	buffer := bytes.NewBuffer(nil)
	writer := fileformats.NewSegmentCSVWriter(buffer)
	m.Iterate(func(s *Segment) {
		writer.Write([4]float64{s[0].X, s[0].Y, s[1].X, s[1].Y})
	})
	return buffer.Bytes()
}

// EncodeSVG encodes the mesh as an SVG file.
func EncodeSVG(m *Mesh) []byte {
	return EncodeCustomSVG([]*Mesh{m}, []string{"black"}, []float64{1.0}, nil)
}

// EncodePathSVG encodes the mesh as a single filled path
// in an SVG file.
func EncodePathSVG(m *Mesh) []byte {
	return EncodeCustomPathSVG([]*Mesh{m}, []string{"black"}, []string{"none"}, []float64{0.0}, nil)
}

// EncodeCustomSVG encodes multiple meshes, each with a
// different color and line thickness.
//
// If bounds is not nil, it is used to determine the
// resulting bounds of the SVG.
// Otherwise, the union of all meshes is used.
func EncodeCustomSVG(meshes []*Mesh, colors []string, thicknesses []float64, bounds Bounder) []byte {
	if len(meshes) != len(colors) {
		panic("incorrect number of colors")
	}
	if len(meshes) != len(thicknesses) {
		panic("incorrect number of thicknesses")
	}

	var min, max Coord
	if bounds != nil {
		min, max = bounds.Min(), bounds.Max()
	} else {
		min, max = BoundsUnion(meshes)
	}

	var result bytes.Buffer
	writer, err := fileformats.NewSVGWriter(&result, [4]float64{
		min.X, min.Y, max.X - min.X, max.Y - min.Y,
	})
	if err != nil {
		panic(err)
	}

	for i, m := range meshes {
		color := colors[i]
		thickness := fmt.Sprintf("%f", thicknesses[i])
		findPolylines(m, func(points []Coord) {
			pointArrs := make([][2]float64, len(points))
			for i, x := range points {
				pointArrs[i] = x.Array()
			}
			err = writer.WritePoly(pointArrs, map[string]string{
				"fill":         "none",
				"stroke-width": thickness,
				"stroke":       color,
			})
			if err != nil {
				panic(err)
			}
		})
	}

	if err := writer.WriteEnd(); err != nil {
		panic(err)
	}
	return result.Bytes()
}

// EncodeCustomPathSVG encodes a collection of meshes each as
// path elements with individually specified parameters.
//
// If bounds is not nil, it is used to determine the
// resulting bounds of the SVG. Otherwise, the bounds of
// the mesh are used.
func EncodeCustomPathSVG(meshes []*Mesh, fillColors []string, strokeColors []string,
	thicknesses []float64, bounds Bounder) []byte {
	if len(meshes) != len(fillColors) {
		panic("incorrect number of fill colors")
	}
	if len(meshes) != len(strokeColors) {
		panic("incorrect number of stroke colors")
	}
	if len(meshes) != len(thicknesses) {
		panic("incorrect number of thicknesses")
	}

	var min, max Coord
	if bounds != nil {
		min, max = bounds.Min(), bounds.Max()
	} else {
		min, max = BoundsUnion(meshes)
	}

	var result bytes.Buffer
	writer, err := fileformats.NewSVGWriter(&result, [4]float64{
		min.X, min.Y, max.X - min.X, max.Y - min.Y,
	})
	if err != nil {
		panic(err)
	}

	for i, m := range meshes {
		fillColor := fillColors[i]
		strokeColor := strokeColors[i]
		thickness := fmt.Sprintf("%f", thicknesses[i])
		var allPaths [][][2]float64
		findPolylines(m, func(points []Coord) {
			pointArrs := make([][2]float64, len(points))
			for i, x := range points {
				pointArrs[i] = x.Array()
			}
			allPaths = append(allPaths, pointArrs)
		})
		err = writer.WritePolyPath(allPaths, map[string]string{
			"fill":         fillColor,
			"stroke-width": thickness,
			"stroke":       strokeColor,
		})
		if err != nil {
			panic(err)
		}
	}

	if err := writer.WriteEnd(); err != nil {
		panic(err)
	}
	return result.Bytes()
}

// findPolylines finds sequences of connected segments and
// calls f for each one.
//
// The f function is called with all of the points in each
// sequence, such that segments connect consecutive
// points.
//
// If the figure is closed, or is open but properly
// connected (with no vertices used more than twice), then
// f is only called once.
func findPolylines(m *Mesh, f func(points []Coord)) {
	m1 := m.Copy()
	for len(m1.faces) > 0 {
		f(findNextPolyline(m1))
	}
}

func findNextPolyline(m *Mesh) []Coord {
	var seg *Segment
	for s := range m.faces {
		seg = s
		break
	}
	m.Remove(seg)

	before := findPolylineFromPoint(m, seg[0])
	after := findPolylineFromPoint(m, seg[1])
	allCoords := make([]Coord, len(before)+len(after))
	for i, c := range before {
		allCoords[len(before)-(i+1)] = c
	}
	copy(allCoords[len(before):], after)

	return allCoords
}

func findPolylineFromPoint(m *Mesh, c Coord) []Coord {
	result := []Coord{c}
	for {
		other := m.Find(c)
		if len(other) == 0 {
			return result
		}
		next := other[0]
		m.Remove(next)
		if next[0] == c {
			c = next[1]
		} else {
			c = next[0]
		}
		result = append(result, c)
	}
}
