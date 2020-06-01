package main

// Location is a (row, column) pair on the board, starting
// at the top left.
//
// It is zero indexed.
type Location [2]int

// A Segment connects two adjacent locations.
//
// Locations are adjacent if they don't differ in one
// dimension, and only differ by 1 in the other.
type Segment [2]Location

// A Digit is a drawing of a character using a set of line
// segments.
type Digit []Segment

// NewDigitContinuous creates a digit by connecting line
// segments between every consecutive pair of locations.
func NewDigitContinuous(points []Location) Digit {
	var res Digit
	for i := 1; i < len(points); i++ {
		res = append(res, Segment{points[i-1], points[i]})
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
