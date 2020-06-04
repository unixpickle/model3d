package main

import "github.com/unixpickle/essentials"

// Location is a (row, column) pair on the board, starting
// at the top left.
//
// It is zero indexed.
type Location [2]int

// Reflect reflects l around the point s, using the
// formula s+(s-l).
func (l Location) Reflect(s Location) Location {
	return Location{s[0] + (s[0] - l[0]), s[1] + (s[1] - l[1])}
}

// A Segment connects two adjacent locations.
//
// Locations are adjacent if they don't differ in one
// dimension, and only differ by 1 in the other.
//
// Segments are ordered deterministically, so long as they
// are created with NewSegment.
type Segment [2]Location

// NewSegment creates a deterministically ordered segment.
func NewSegment(l1, l2 Location) Segment {
	if l1[0] < l2[0] || (l1[0] == l2[0] && l1[1] < l2[1]) {
		return Segment{l1, l2}
	} else {
		return Segment{l2, l1}
	}
}

// A Digit is a drawing of a character using a set of line
// segments.
type Digit []Segment

// NewDigitContinuous creates a digit by connecting line
// segments between every consecutive pair of locations.
func NewDigitContinuous(points []Location) Digit {
	var res Digit
	for i := 1; i < len(points); i++ {
		res = append(res, NewSegment(points[i-1], points[i]))
	}
	return res
}

// AllDigits creates the numerical digits.
func AllDigits() []Digit {
	return []Digit{
		// 1 through 5.
		NewDigitContinuous([]Location{{0, 0}, {1, 0}, {2, 0}}),
		NewDigitContinuous([]Location{{0, 0}, {0, 1}, {1, 1}, {1, 0}, {2, 0}, {2, 1}}),
		append(
			NewDigitContinuous([]Location{{0, 0}, {0, 1}, {1, 1}, {1, 0}}),
			NewDigitContinuous([]Location{{1, 1}, {2, 1}, {2, 0}})...,
		),
		append(
			NewDigitContinuous([]Location{{0, 0}, {1, 0}, {1, 1}, {0, 1}}),
			NewDigitContinuous([]Location{{1, 1}, {2, 1}})...,
		),
		NewDigitContinuous([]Location{{0, 1}, {0, 0}, {1, 0}, {1, 1}, {2, 1}, {2, 0}}),

		// 6-9
		NewDigitContinuous([]Location{{0, 1}, {0, 0}, {1, 0}, {1, 1}, {2, 1}, {2, 0}, {1, 0}}),
		NewDigitContinuous([]Location{{0, 0}, {0, 1}, {1, 1}, {2, 1}}),
		append(
			NewDigitContinuous([]Location{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}}),
			NewDigitContinuous([]Location{{1, 0}, {2, 0}, {2, 1}, {1, 1}})...,
		),
		NewDigitContinuous([]Location{{2, 0}, {2, 1}, {1, 1}, {0, 1}, {0, 0}, {1, 0}, {1, 1}}),
	}
}

// Copy copies the memory of d.
func (d Digit) Copy() Digit {
	return append(Digit{}, d...)
}

// Rotate rotates the digit 90 degrees clockwise and keeps
// it tucked in the top-left corner.
func (d Digit) Rotate() {
	for i, s := range d {
		s1 := s
		for i, l := range s1 {
			s1[i] = Location{-l[1], l[0]}
		}
		// Order it properly.
		d[i] = NewSegment(s1[0], s1[1])
	}
	min := d.Min()
	d.Translate(Location{-min[0], -min[1]})
}

// Translate adds l to the locations in the digit.
func (d Digit) Translate(l Location) {
	for i, s := range d {
		s1 := s
		for i, l1 := range s1 {
			for j, c1 := range l1 {
				s1[i][j] = c1 + l[j]
			}
		}
		// Order it properly.
		d[i] = NewSegment(s1[0], s1[1])
	}
}

// Min gets the minimum location in the digit.
func (d Digit) Min() Location {
	res := d[0][0]
	for _, s := range d {
		for _, l := range s {
			for i, c := range l {
				res[i] = essentials.MinInt(res[i], c)
			}
		}
	}
	return res
}

// Max gets the maximum location in the digit.
func (d Digit) Max() Location {
	res := d[0][0]
	for _, s := range d {
		for _, l := range s {
			for i, c := range l {
				res[i] = essentials.MaxInt(res[i], c)
			}
		}
	}
	return res
}
