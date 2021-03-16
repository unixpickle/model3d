package model2d

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// DecodeCSV decodes the CSV format from EncodeCSV().
func DecodeCSV(data []byte) ([]*Segment, error) {
	lines := strings.Split(string(data), "\n")
	res := []*Segment{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) != 4 {
			return nil, errors.New("read csv: invalid number of columns")
		}
		var values [4]float64
		for i, part := range parts {
			value, err := strconv.ParseFloat(part, 64)
			if err != nil {
				return nil, errors.Wrap(err, "read csv")
			}
			values[i] = value
		}
		res = append(res, &Segment{
			XY(values[0], values[1]),
			XY(values[2], values[3]),
		})
	}
	return res, nil
}
