package fileformats

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/pkg/errors"
)

// An STLWriter writes a triangle mesh in the STL format.
type STLWriter struct {
	w        io.Writer
	trisLeft uint32

	buffer [12]float32
}

// NewSTLWriter creates an STLWriter and writes a header,
// which requires knowledge of the total number of
// triangles being written.
func NewSTLWriter(w io.Writer, numTris uint32) (*STLWriter, error) {
	if _, err := w.Write(make([]byte, 80)); err != nil {
		return nil, errors.Wrap(err, "write STL header")
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(numTris)); err != nil {
		return nil, errors.Wrap(err, "write STL header")
	}
	return &STLWriter{w: w, trisLeft: numTris}, nil
}

// WriteTriangle writes a triangle to the file.
//
// This should be called exactly the number of times passed
// to NewSTLWriter.
func (s *STLWriter) WriteTriangle(normal [3]float32, faces [3][3]float32) error {
	if s.trisLeft == 0 {
		return errors.New("write STL triangle: too many triangles written")
	}
	s.trisLeft -= 1
	copy(s.buffer[0:3], normal[:])
	copy(s.buffer[3:6], faces[0][:])
	copy(s.buffer[6:9], faces[1][:])
	copy(s.buffer[9:12], faces[2][:])

	if err := binary.Write(s.w, binary.LittleEndian, s.buffer[:]); err != nil {
		return errors.Wrap(err, "write STL triangle")
	}
	if _, err := s.w.Write([]byte{0, 0}); err != nil {
		return errors.Wrap(err, "write STL triangle")
	}
	return nil
}

// An STLReader reads STL files.
type STLReader struct {
	r       io.Reader
	numTris uint32
}

// NewSTLReader creates an STL reader by reading the header
// of an STL file.
func NewSTLReader(r io.Reader) (*STLReader, error) {
	header := make([]byte, 80)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, errors.Wrap(err, "read STL header")
	}
	var numTris uint32
	if err := binary.Read(r, binary.LittleEndian, &numTris); err != nil {
		return nil, errors.Wrap(err, "read STL header")
	}
	return &STLReader{
		r:       r,
		numTris: numTris,
	}, nil
}

// NumTriangles gets the total number of triangles in the
// file as reported by the header.
func (s *STLReader) NumTriangles() uint32 {
	return s.numTris
}

// ReadTriangle reads the next triangle from the file.
func (s *STLReader) ReadTriangle() (normal [3]float32, vertices [3][3]float32, err error) {
	var data [4*4*3 + 2]byte
	if _, err = io.ReadFull(s.r, data[:]); err != nil {
		err = errors.Wrap(err, "read STL triangle")
		return
	}
	for i := 0; i < 3; i++ {
		normal[i] = math.Float32frombits(binary.LittleEndian.Uint32(data[i*4 : (i+1)*4]))
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			vertices[i][j] = math.Float32frombits(binary.LittleEndian.Uint32(data[(i*3+j+3)*4 : (i*3+j+4)*4]))
		}
	}
	return
}
