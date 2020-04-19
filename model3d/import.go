package model3d

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/pkg/errors"
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
	header := make([]byte, 80)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}
	var numTris uint32
	if err := binary.Read(r, binary.LittleEndian, &numTris); err != nil {
		return nil, err
	}
	data := make([]byte, numTris*(4*4*3+2))
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}
	vertexData := make([]byte, 0, numTris*4*3*3)
	for i := 0; i < int(numTris); i++ {
		idx := i*(4*4*3+2) + 4*3
		vertexData = append(vertexData, data[idx:idx+4*3*3]...)
	}
	vertices := make([]float32, numTris*3*3)
	if err := binary.Read(bytes.NewReader(vertexData), binary.LittleEndian, vertices); err != nil {
		return nil, err
	}
	triangles := make([]*Triangle, numTris)
	for i := range triangles {
		verts := make([]float64, 9)
		for j := range verts {
			verts[j] = float64(vertices[i*3*3+j])
		}
		triangles[i] = &Triangle{
			Coord3D{verts[0], verts[1], verts[2]},
			Coord3D{verts[3], verts[4], verts[5]},
			Coord3D{verts[6], verts[7], verts[8]},
		}
	}
	return triangles, nil
}

// ReadOFF decodes a file in the object file format.
// See http://segeval.cs.princeton.edu/public/off_format.html.
func ReadOFF(r io.Reader) ([]*Triangle, error) {
	tris, err := readOFF(r)
	if err != nil {
		return nil, errors.Wrap(err, "read OFF")
	}
	return tris, nil
}

func readOFF(r io.Reader) ([]*Triangle, error) {
	reader := bufio.NewReader(r)
	headerLines := 1
	line1, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(line1, "OFF") {
		return nil, errors.New("line 1: expected 'OFF' as first line")
	}

	var line2 string
	if len(line1) > 4 {
		line2 = line1[3:]
	} else {
		line2, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		headerLines++
	}

	parts := strings.Fields(line2)
	if len(parts) != 3 {
		return nil, errors.New("line 2: unexpected number of tokens")
	}
	numVerts, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, errors.New("line 2: invalid vertex count")
	}
	numFaces, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, errors.New("line 2: invalid face count")
	}

	vertices := make([]Coord3D, numVerts)
	for i := 0; i < numVerts; i++ {
		lineIdx := i + headerLines + 1
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, errors.Wrapf(err, "line %d", lineIdx)
		}
		parts := strings.Fields(line)
		if len(parts) != 3 {
			return nil, fmt.Errorf("line %d: unexpected number of tokens", lineIdx)
		}
		var numbers [3]float64
		for i, part := range parts {
			num, err := strconv.ParseFloat(part, 64)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid vector component", lineIdx)
			}
			numbers[i] = num
		}
		vertices[i] = NewCoord3DArray(numbers)
	}

	triangles := make([]*Triangle, 0, numFaces)
	for i := 0; i < numFaces; i++ {
		lineIdx := i + numVerts + headerLines + 1
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, errors.Wrapf(err, "line %d", lineIdx)
		}
		parts := strings.Fields(line)
		if len(parts) == 0 {
			return nil, fmt.Errorf("line %d: no tokens", lineIdx)
		}
		numComponents, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}
		if numComponents+1 != len(parts) {
			return nil, fmt.Errorf("line %d: unexpected number of components", lineIdx)
		}
		poly := make([]Coord3D, numComponents)
		for i, componentStr := range parts[1:] {
			idx, err := strconv.Atoi(componentStr)
			if err != nil || idx < 0 || idx >= len(vertices) {
				return nil, fmt.Errorf("line %d: invalid vertex index", lineIdx)
			}
			poly[i] = vertices[idx]
		}
		triangles = append(triangles, TriangulateFace(poly)...)
	}
	return triangles, nil
}
