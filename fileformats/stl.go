package fileformats

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"strconv"
	"strings"

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
	r             io.Reader
	br            *bufio.Reader
	isBinary      bool
	doneNonBinary bool
	numTris       uint32
	readTris      uint32
}

// NewSTLReader creates an STL reader by reading the header
// of an STL file.
func NewSTLReader(r io.Reader) (*STLReader, error) {
	firstBytes := make([]byte, 5)
	if _, err := io.ReadFull(r, firstBytes); err != nil {
		return nil, errors.Wrap(err, "read STL header")
	}
	if bytes.Equal(firstBytes, []byte("solid")) {
		br := bufio.NewReader(r)
		_, err := br.ReadString('\n')
		if err != nil {
			return nil, errors.Wrap(err, "read STL header")
		}
		return &STLReader{br: br}, nil
	} else {
		header := make([]byte, 80-5)
		if _, err := io.ReadFull(r, header); err != nil {
			return nil, errors.Wrap(err, "read STL header")
		}
		var numTris uint32
		if err := binary.Read(r, binary.LittleEndian, &numTris); err != nil {
			return nil, errors.Wrap(err, "read STL header")
		}
		return &STLReader{
			r:        r,
			isBinary: true,
			numTris:  numTris,
		}, nil
	}
}

// IsBinary returns true if this file is encoded in the
// binary STL format, rather than the ASCII format.
func (s *STLReader) IsBinary() bool {
	return s.isBinary
}

// NumTriangles gets the total number of triangles in the
// file as reported by the header, if this is a binary
// file. If it is an ASCII file, this always returns 0.
func (s *STLReader) NumTriangles() uint32 {
	return s.numTris
}

// ReadTriangle reads the next triangle from the file, or
// returns an error if no more triangles can be read.
//
// The error io.EOF is returned when the file ended and
// there are no more triangles to be read. When the file
// ends abruptly and incorrectly, io.ErrUnexpectedEOF is
// returned instead.
func (s *STLReader) ReadTriangle() (normal [3]float32, vertices [3][3]float32, err error) {
	if s.isBinary {
		normal, vertices, err = s.readBinary()
	} else {
		normal, vertices, err = s.readASCII()
	}
	if err != nil && !errors.Is(err, io.EOF) {
		err = errors.Wrap(err, "read STL triangle")
	}
	return
}

func (s *STLReader) readASCII() (normal [3]float32, vertices [3][3]float32, err error) {
	if s.doneNonBinary {
		err = io.EOF
		return
	}

	var vertexIndex int
	for {
		nextLine, err := s.br.ReadString('\n')
		nextLine = strings.TrimSpace(nextLine)
		if errors.Is(err, io.EOF) {
			if strings.HasPrefix(nextLine, "endsolid") {
				s.doneNonBinary = true
				return normal, vertices, io.EOF
			} else {
				return normal, vertices, io.ErrUnexpectedEOF
			}
		} else if err != nil {
			return normal, vertices, err
		}
		tokens := strings.Fields(nextLine)
		if len(tokens) == 0 {
			continue
		}
		if tokens[0] == "endsolid" {
			s.doneNonBinary = true
			return normal, vertices, io.EOF
		} else if tokens[0] == "endfacet" {
			break
		} else if tokens[0] == "facet" {
			if len(tokens) != 5 {
				return normal, vertices, errors.New("unexpected facet line: '" + nextLine + "'")
			}
			normal, err = parseSTLVector(tokens)
			if err != nil {
				return normal, vertices, err
			}
		} else if tokens[0] == "vertex" {
			if len(tokens) != 4 {
				return normal, vertices, errors.New("unexpected vertex line: '" + nextLine + "'")
			} else if vertexIndex == 3 {
				return normal, vertices, errors.New("more than three vertices in a facet")
			}
			vertices[vertexIndex], err = parseSTLVector(tokens)
			if err != nil {
				return normal, vertices, err
			}
			vertexIndex++
		}
	}
	if vertexIndex != 3 {
		return normal, vertices, errors.New("unexpected number of vertices before endfacet")
	}
	return normal, vertices, nil
}

func parseSTLVector(line []string) ([3]float32, error) {
	var vertex [3]float32
	for i, token := range line[len(line)-3:] {
		parsed, err := strconv.ParseFloat(token, 32)
		if err != nil {
			return [3]float32{}, err
		}
		vertex[i] = float32(parsed)
	}
	return vertex, nil
}

func (s *STLReader) readBinary() (normal [3]float32, vertices [3][3]float32, err error) {
	if s.readTris == s.numTris {
		err = io.EOF
		return
	}
	var data [4*4*3 + 2]byte
	if _, err = io.ReadFull(s.r, data[:]); err != nil {
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
