package fileformats

import (
	"encoding/binary"
	"io"

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
