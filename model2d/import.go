package model2d

import (
	"bytes"
	"io"

	"github.com/unixpickle/model3d/fileformats"
)

// DecodeCSV decodes the CSV format from EncodeCSV().
func DecodeCSV(data []byte) ([]*Segment, error) {
	r := fileformats.NewSegmentCSVReader(bytes.NewReader(data))
	res := []*Segment{}
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		res = append(res, &Segment{XY(row[0], row[1]), XY(row[2], row[3])})
	}
	return res, nil
}
