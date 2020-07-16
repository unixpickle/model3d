package model2d

import (
	"fmt"
	"strings"
)

func EncodeSVG(m *Mesh) []byte {
	min := m.Min()
	max := m.Max()

	var result strings.Builder
	result.WriteString(`<?xml version="1.0" encoding="utf-8" ?>`)
	result.WriteString(
		fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" version="1.1" viewBox="%f %f %f %f">`,
			min.X, min.Y, max.X-min.X, max.Y-min.Y))

	findPolylines(m, func(points []Coord) {
		pointStrs := make([]string, len(points))
		for i, c := range points {
			pointStrs[i] = fmt.Sprintf("%f,%f", c.X, c.Y)
		}
		if pointStrs[0] == pointStrs[len(pointStrs)-1] {
			pointStrs = pointStrs[1:]
			result.WriteString(`<polygon points="`)
		} else {
			result.WriteString(`<polyline points="`)
		}
		result.WriteString(strings.Join(pointStrs, " "))
		result.WriteString(`" stroke="black" fill="none" />`)
	})

	result.WriteString("</svg>")
	return []byte(result.String())
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
	m1 := NewMesh()
	m1.AddMesh(m)
	for len(m1.segments) > 0 {
		f(findNextPolyline(m1))
	}
}

func findNextPolyline(m *Mesh) []Coord {
	var seg *Segment
	for s := range m.segments {
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
