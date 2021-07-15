package fileformats

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/unixpickle/essentials"
)

// An OFFReader reads an OFF file.
//
// For info on the OFF format, see
// http://segeval.cs.princeton.edu/public/off_format.html.
type OFFReader struct {
	r           *bufio.Reader
	numVerts    int
	numFaces    int
	headerLines int
	curFace     int

	vertices [][3]float64
}

// NewOFFReader reads the header from an OFF file and
// returns the new reader, if successful.
func NewOFFReader(r io.Reader) (o *OFFReader, err error) {
	defer essentials.AddCtxTo("open OFF file", &err)

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
	return &OFFReader{
		r:           reader,
		numVerts:    numVerts,
		numFaces:    numFaces,
		headerLines: headerLines,
	}, nil
}

// NumFaces returns the total number of faces.
func (o *OFFReader) NumFaces() int {
	return o.numFaces
}

// ReadFace reads the next face. If vertices have not been
// read, they will be loaded first.
//
// If no more faces exist to be read, io.EOF is returned
// as the error.
func (o *OFFReader) ReadFace() (faces [][3]float64, err error) {
	defer func() {
		if err != io.EOF {
			err = essentials.AddCtx("read OFF face", err)
		}
	}()

	fmt.Println(o.numFaces, o.curFace)

	if o.curFace == o.numFaces {
		return nil, io.EOF
	} else if o.vertices == nil {
		if err := o.readVertices(); err != nil {
			return nil, err
		}
	}

	lineIdx := o.curFace + o.numVerts + o.headerLines + 1
	line, err := o.r.ReadString('\n')
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
	poly := make([][3]float64, numComponents)
	for i, componentStr := range parts[1:] {
		idx, err := strconv.Atoi(componentStr)
		if err != nil || idx < 0 || idx >= len(o.vertices) {
			return nil, fmt.Errorf("line %d: invalid vertex index", lineIdx)
		}
		poly[i] = o.vertices[idx]
	}
	o.curFace++
	return poly, nil
}

// readVertices loads the vertices from the file.
func (o *OFFReader) readVertices() (err error) {
	defer essentials.AddCtxTo("read OFF vertices", &err)
	vertices := make([][3]float64, o.numVerts)
	for i := 0; i < o.numVerts; i++ {
		lineIdx := i + o.headerLines + 1
		line, err := o.r.ReadString('\n')
		if err != nil {
			return errors.Wrapf(err, "line %d", lineIdx)
		}
		parts := strings.Fields(line)
		if len(parts) != 3 {
			return fmt.Errorf("line %d: unexpected number of tokens", lineIdx)
		}
		var numbers [3]float64
		for i, part := range parts {
			num, err := strconv.ParseFloat(part, 64)
			if err != nil {
				return fmt.Errorf("line %d: invalid vector component", lineIdx)
			}
			numbers[i] = num
		}
		vertices[i] = numbers
	}
	o.vertices = vertices
	return nil
}
