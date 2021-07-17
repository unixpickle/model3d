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

// A PLYWriter encodes a PLY writable stream.
//
// This may use buffering as it writes the file, but the
// full file will always be flushed by the time the last
// triangle is written.
type PLYWriter struct {
	w *bufio.Writer

	numCoords int
	numTris   int

	writtenCoords int
	writtenTris   int

	builder strings.Builder
}

// NewPLYWriter creates a new PLYWriter and writes the
// file header.
func NewPLYWriter(w io.Writer, numCoords, numTris int) (*PLYWriter, error) {
	var header strings.Builder
	header.WriteString("ply\nformat ascii 1.0\n")
	header.WriteString(fmt.Sprintf("element vertex %d\n", numCoords))
	header.WriteString("property float x\n")
	header.WriteString("property float y\n")
	header.WriteString("property float z\n")
	header.WriteString("property uchar red\n")
	header.WriteString("property uchar green\n")
	header.WriteString("property uchar blue\n")
	header.WriteString(fmt.Sprintf("element face %d\n", numTris))
	header.WriteString("property list uchar int vertex_index\n")
	header.WriteString("end_header\n")

	bw := bufio.NewWriter(w)
	if _, err := bw.WriteString(header.String()); err != nil {
		return nil, errors.Wrap(err, "write PLY")
	}

	if err := bw.Flush(); err != nil {
		return nil, err
	}

	return &PLYWriter{
		w:         bw,
		numCoords: numCoords,
		numTris:   numTris,
	}, nil
}

// WriteCoord writes the next coordinate to the file.
//
// This should be called exactly numCoords times.
func (p *PLYWriter) WriteCoord(c [3]float64, color [3]uint8) (err error) {
	defer essentials.AddCtxTo("write PLY", &err)
	if p.writtenTris > 0 || p.writtenCoords >= p.numCoords {
		return errors.New("cannot write another coordinate")
	}
	coordLine := fmt.Sprintf("%f %f %f %d %d %d\n", c[0], c[1], c[2],
		int(color[0]), int(color[1]), int(color[2]))
	_, err = p.w.WriteString(coordLine)
	p.writtenCoords++
	return
}

// WriteTriangle writes the next triangle to the file.
//
// This should be called exactly numTris times.
func (p *PLYWriter) WriteTriangle(coords [3]int) (err error) {
	defer essentials.AddCtxTo("write PLY", &err)

	if p.writtenCoords < p.numCoords {
		return errors.New("must write all coordinates before a triangle")
	} else if p.writtenTris >= p.numTris {
		return errors.New("too many triangles written")
	}

	p.builder.Reset()
	p.builder.WriteString("3")
	for _, idx := range coords {
		p.builder.WriteByte(' ')
		p.builder.WriteString(strconv.Itoa(idx))
	}
	p.builder.WriteByte('\n')
	if _, err := p.w.WriteString(p.builder.String()); err != nil {
		return err
	}

	p.writtenTris++
	if p.writtenTris == p.numTris {
		return p.w.Flush()
	}

	return nil
}
