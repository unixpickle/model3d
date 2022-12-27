package fileformats

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
)

// An SVGWriter encodes paths to SVG files.
type SVGWriter struct {
	w io.Writer
}

// NewSVGWriter creates writes an SVG header and returns a
// new SVGWriter.
//
// The viewbox argument specifies the bounding box of the
// image as [minX, minY, width, height].
func NewSVGWriter(w io.Writer, viewbox [4]float64) (*SVGWriter, error) {
	header := `<?xml version="1.0" encoding="utf-8" ?>` +
		fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" version="1.1" viewBox="%f %f %f %f">`,
			viewbox[0], viewbox[1], viewbox[2], viewbox[3])
	if _, err := w.Write([]byte(header)); err != nil {
		return nil, errors.Wrap(err, "write SVG header")
	}
	return &SVGWriter{w: w}, nil
}

// WritePoly writes a polygon or a polyline depending on
// whether the final point matches up with the first.
func (s *SVGWriter) WritePoly(points [][2]float64, attrs map[string]string) error {
	pointStrs := make([]string, len(points))
	for i, c := range points {
		pointStrs[i] = fmt.Sprintf("%f,%f", c[0], c[1])
	}
	var line string
	if pointStrs[0] == pointStrs[len(pointStrs)-1] {
		pointStrs = pointStrs[1:]
		line = `<polygon points="`
	} else {
		line = `<polyline points="`
	}
	line += strings.Join(pointStrs, " ")
	line += `"`

	var attrStrings []string
	for attribute, value := range attrs {
		var encodedString bytes.Buffer
		if err := xml.EscapeText(&encodedString, []byte(value)); err != nil {
			return errors.Wrap(err, "write SVG polygon")
		}
		attrStrings = append(attrStrings, fmt.Sprintf("%s=\"%s\"", attribute, string(encodedString.Bytes())))
	}
	line += " " + strings.Join(attrStrings, " ")
	if len(attrs) != 1 {
		line += " "
	}
	line += "/>"
	_, err := s.w.Write([]byte(line))
	if err != nil {
		return errors.Wrap(err, "write SVG polygon")
	}
	return nil
}

// WritePolyPath writes one or more polygons into a single
// path element. Each polygon is either closed or left open
// depending on whether the final point matches up with the
// first.
func (s *SVGWriter) WritePolyPath(paths [][][2]float64, attrs map[string]string) error {
	fullPath := []string{}
	for _, points := range paths {
		pointStrs := make([]string, len(points))
		for i, c := range points {
			start := 'L'
			if i == 0 {
				start = 'M'
			}
			pointStrs[i] = fmt.Sprintf("%c%f,%f", start, c[0], c[1])
		}
		if len(pointStrs) > 1 && pointStrs[0][1:] == pointStrs[len(pointStrs)-1][1:] {
			pointStrs[len(pointStrs)-1] = "z"
		}
		fullPath = append(fullPath, strings.Join(pointStrs, " "))
	}
	line := "<path d=\"" + strings.Join(fullPath, " ") + "\""

	var attrStrings []string
	for attribute, value := range attrs {
		var encodedString bytes.Buffer
		if err := xml.EscapeText(&encodedString, []byte(value)); err != nil {
			return errors.Wrap(err, "write SVG polygon path")
		}
		attrStrings = append(attrStrings, fmt.Sprintf("%s=\"%s\"", attribute, string(encodedString.Bytes())))
	}
	line += " " + strings.Join(attrStrings, " ")
	if len(attrs) != 1 {
		line += " "
	}
	line += "/>"
	_, err := s.w.Write([]byte(line))
	if err != nil {
		return errors.Wrap(err, "write SVG polygon path")
	}
	return nil
}

// WriteEnd writes any necessary footer information.
func (s *SVGWriter) WriteEnd() error {
	_, err := s.w.Write([]byte("</svg>"))
	if err != nil {
		return errors.Wrap(err, "write SVG footer")
	}
	return nil
}
