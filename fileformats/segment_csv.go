package fileformats

import (
	"encoding/csv"
	"io"
	"strconv"

	"github.com/pkg/errors"
)

// A SegmentCSVReader reads a CSV file encoding a 2D mesh
// where each row is one segment.
type SegmentCSVReader struct {
	r *csv.Reader
}

// NewSegmentCSVReader creates a SegmentCSVReader that
// wraps an underlying io.Reader.
func NewSegmentCSVReader(r io.Reader) *SegmentCSVReader {
	csvReader := csv.NewReader(r)
	csvReader.ReuseRecord = true
	csvReader.FieldsPerRecord = 4
	return &SegmentCSVReader{r: csvReader}
}

// Read reads a segment from the file.
//
// Returns io.EOF if the file is complete.
func (s *SegmentCSVReader) Read() ([4]float64, error) {
	var res [4]float64

	record, err := s.r.Read()
	if err != nil {
		if err != io.EOF {
			err = errors.Wrap(err, "read segment CSV row")
		}
		return res, err
	}
	for i, x := range record {
		res[i], err = strconv.ParseFloat(x, 64)
		if err != nil {
			return res, errors.Wrap(err, "read segment CSV row")
		}
	}
	return res, nil
}

// A SegmentCSVWriter writes a CSV file encoding a 2D mesh
// where each row is one segment.
type SegmentCSVWriter struct {
	w io.Writer
}

// NewSegmentCSVWriter creates a SegmentCSVWriter that
// wraps an underlying io.Writer.
func NewSegmentCSVWriter(w io.Writer) *SegmentCSVWriter {
	return &SegmentCSVWriter{w: w}
}

// Write writes a segment to the file.
//
// The segment is stored as {x1, y1, x2, y2}.
func (s *SegmentCSVWriter) Write(seg [4]float64) error {
	line := ""
	for i, x := range seg {
		if i > 0 {
			line += ","
		}
		line += strconv.FormatFloat(x, 'G', -1, 64)
	}
	_, err := s.w.Write([]byte(line + "\n"))
	if err != nil {
		err = errors.Wrap(err, "write segment CSV row")
	}
	return err
}
