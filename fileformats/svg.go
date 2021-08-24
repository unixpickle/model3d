package fileformats

import (
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
)

type SVGWriter struct {
	w io.Writer
}

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
func (s *SVGWriter) WritePoly(name string, points [][2]float64, attrs map[string]string) error {
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

	var attrStrings []string
	for attribute, value := range attrs {
		attrStrings = append(attrStrings, fmt.Sprintf("%s=\"%s\"", attribute, value))
	}
	line += " " + strings.Join(attrStrings, " ")
	if len(attrs) != 1 {
		line += " "
	}
	line += "/>"
	_, err := s.w.Write([]byte("</svg>"))
	if err != nil {
		return errors.Wrap(err, "write SVG polygon")
	}
	return nil
}

func (s *SVGWriter) WriteEnd() error {
	_, err := s.w.Write([]byte("</svg>"))
	if err != nil {
		return errors.Wrap(err, "write SVG footer")
	}
	return nil
}
