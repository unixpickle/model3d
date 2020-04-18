package model2d

import (
	"fmt"
	"strings"
)

func EncodeSVG(m *Mesh) []byte {
	segments := map[*Segment]bool{}
	m.Iterate(func(s *Segment) {
		segments[s] = true
	})

	min := m.Min()
	max := m.Max()

	var result strings.Builder
	result.WriteString(`<?xml version="1.0" encoding="utf-8" ?>`)
	result.WriteString(
		fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" version="1.1" viewBox="%f %f %f %f">`,
			min.X, min.Y, max.X-min.X, max.Y-min.Y))
	for len(segments) > 0 {
		var seg *Segment
		for s := range segments {
			seg = s
			break
		}
		delete(segments, seg)
		pointStrs := []string{fmt.Sprintf("%f,%f", seg[0].X, seg[0].Y)}
	PathLoop:
		for {
			pointStrs = append(pointStrs, fmt.Sprintf("%f,%f", seg[1].X, seg[1].Y))
			for _, s := range m.Find(seg[1]) {
				if s == seg || s[0] != seg[1] {
					continue
				}
				if segments[s] {
					delete(segments, s)
					seg = s
					continue PathLoop
				}
			}
			break
		}
		if pointStrs[0] == pointStrs[len(pointStrs)-1] {
			pointStrs = pointStrs[1:]
		}
		result.WriteString(`<polygon points="`)
		result.WriteString(strings.Join(pointStrs, " "))
		result.WriteString(`" stroke="black" fill="none" />`)
	}
	result.WriteString("</svg>")
	return []byte(result.String())
}
