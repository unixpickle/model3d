package model3d

import (
	"bufio"
	"io"

	"github.com/pkg/errors"
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
