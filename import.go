package model3d

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// ReadOFF decodes a file in the object file format.
// See http://segeval.cs.princeton.edu/public/off_format.html.
func ReadOFF(r io.Reader) (*Mesh, error) {
	mesh, err := readOFF(r)
	if err != nil {
		return nil, errors.Wrap(err, "read OFF")
	}
	return mesh, err
}

func readOFF(r io.Reader) (*Mesh, error) {
	reader := bufio.NewReader(r)
	line1, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if line1 != "OFF\n" {
		return nil, errors.New("line 1: expected 'OFF' as first line")
	}
	line2, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
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
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, errors.Wrapf(err, "line %d", i+3)
		}
		parts := strings.Fields(line)
		if len(parts) != 3 {
			return nil, fmt.Errorf("line %d: unexpected number of tokens", i+3)
		}
		var numbers [3]float64
		for i, part := range parts {
			num, err := strconv.ParseFloat(part, 64)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid vector component", i+3)
			}
			numbers[i] = num
		}
		vertices[i] = newCoord3DArray(numbers)
	}

	m := NewMesh()
	for i := 0; i < numFaces; i++ {
		lineIdx := i + numVerts + 3
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
		for _, triangle := range TriangulateFace(poly) {
			m.Add(triangle)
		}
	}

	return m, nil
}
