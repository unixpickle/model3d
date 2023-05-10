package model3d

import (
	"bufio"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/fileformats"
)

// ReadSTL decodes a file in the STL file format.
func ReadSTL(r io.Reader) ([]*Triangle, error) {
	tris, err := readSTL(r)
	if err != nil {
		return nil, errors.Wrap(err, "read STL")
	}
	return tris, nil
}

func readSTL(r io.Reader) ([]*Triangle, error) {
	br := bufio.NewReader(r)
	reader, err := fileformats.NewSTLReader(br)
	if err != nil {
		return nil, err
	}
	tris := make([]*Triangle, 0, int(reader.NumTriangles()))
	for {
		_, vertices, err := reader.ReadTriangle()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		tri := &Triangle{}
		for j, vert := range vertices {
			tri[j] = XYZ(float64(vert[0]), float64(vert[1]), float64(vert[2]))
		}
		tris = append(tris, tri)
	}
	return tris, nil
}

// ReadOFF decodes a file in the object file format.
// See http://segeval.cs.princeton.edu/public/off_format.html.
func ReadOFF(r io.Reader) ([]*Triangle, error) {
	o, err := fileformats.NewOFFReader(r)
	if err != nil {
		return nil, err
	}
	triangles := make([]*Triangle, 0, o.NumFaces())
	for i := 0; i < o.NumFaces(); i++ {
		face, err := o.ReadFace()
		if err != nil {
			return nil, err
		}
		poly := make([]Coord3D, len(face))
		for i, x := range face {
			poly[i] = NewCoord3DArray(x)
		}
		triangles = append(triangles, TriangulateFace(poly)...)
	}
	return triangles, nil
}

// ReadColorPLY decodes a PLY file with vertex colors.
func ReadColorPLY(r io.Reader) ([]*Triangle, *CoordMap[[3]uint8], error) {
	tris, colors, err := readColorPLY(r)
	return tris, colors, essentials.AddCtx("read color PLY", err)
}

func readColorPLY(r io.Reader) ([]*Triangle, *CoordMap[[3]uint8], error) {
	reader, err := fileformats.NewPLYReader(r)
	if err != nil {
		return nil, nil, err
	}

	var foundFaces, foundVertices bool
	for _, element := range reader.Header().Elements {
		if element.Name == "vertex" {
			if !element.IsStandardVertex() {
				return nil, nil,
					fmt.Errorf("unexpected vertex element: %s", element.Encode())
			}
			foundVertices = true
		} else if element.Name == "face" {
			if !element.IsStandardFace() {
				return nil, nil, fmt.Errorf("unexpected vertex element: %s", element.Encode())
			}
			foundFaces = true
		}
	}
	if !foundFaces {
		return nil, nil, errors.New("missing 'face' element")
	}
	if !foundVertices {
		return nil, nil, errors.New("missing 'vertex' element")
	}

	var vertices []Coord3D
	var triangles [][3]int
	colors := NewCoordMap[[3]uint8]()

	for {
		values, element, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if element.Name == "face" {
			val := values[0].(fileformats.PLYValueList)
			if val.Length.(fileformats.PLYValueUint8).Value != 3 {
				return nil, nil,
					fmt.Errorf("expected triangles but got face of count: %d", val.Length)
			}
			triangles = append(triangles, [3]int{
				int(val.Values[0].(fileformats.PLYValueInt32).Value),
				int(val.Values[1].(fileformats.PLYValueInt32).Value),
				int(val.Values[2].(fileformats.PLYValueInt32).Value),
			})
		} else {
			var r, g, b uint8
			var x, y, z float64
			for i, value := range values {
				switch element.Properties[i].Name {
				case "red":
					r = value.(fileformats.PLYValueUint8).Value
				case "green":
					g = value.(fileformats.PLYValueUint8).Value
				case "blue":
					b = value.(fileformats.PLYValueUint8).Value
				case "x":
					x = float64(value.(fileformats.PLYValueFloat32).Value)
				case "y":
					y = float64(value.(fileformats.PLYValueFloat32).Value)
				case "z":
					z = float64(value.(fileformats.PLYValueFloat32).Value)
				}
			}
			vertex := XYZ(x, y, z)
			vertices = append(vertices, vertex)
			colors.Store(vertex, [3]uint8{r, g, b})
		}
	}

	tris := make([]*Triangle, len(triangles))
	for i, t := range triangles {
		for _, v := range t {
			if v >= len(vertices) {
				return nil, nil, errors.New("vertex out of bounds")
			}
		}
		tris[i] = &Triangle{vertices[t[0]], vertices[t[1]], vertices[t[2]]}
	}

	return tris, colors, nil
}
