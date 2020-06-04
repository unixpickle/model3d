package main

// SearchPlacement finds a transformed version of each
// digit such that all of the digits fit on the board at
// once.
func SearchPlacement(fixed, digits []Digit, boardSize int) []Digit {
	freeSegs := (boardSize + 1) * boardSize * 2
	for _, ds := range [][]Digit{fixed, digits} {
		for _, d := range ds {
			freeSegs -= len(d)
		}
	}
	if freeSegs < 0 {
		panic("too many segments for board size")
	}

	state := newBoardState(boardSize)
	for _, d := range fixed {
		state.Add(d)
	}
	digitCopy := make([]Digit, len(digits))
	for i, d := range digits {
		digitCopy[i] = d.Copy()
	}
	if searchRecursively(digitCopy, state) {
		return digitCopy
	}
	return nil
}

func searchRecursively(digits []Digit, state *boardState) bool {
	if len(digits) == 0 {
		return true
	}
	digit := digits[0]
	remaining := digits[1:]
	for i := 0; i < 4; i++ {
		for x := 0; x <= state.Size(); x++ {
			for y := 0; y <= state.Size(); y++ {
				if state.CanAdd(digit) {
					state.Add(digit)
					if searchRecursively(remaining, state) {
						return true
					}
					state.Remove(digit)
				}
				digit.Translate(Location{0, 1})
			}
			digit.Translate(Location{1, -(state.Size() + 1)})
		}
		// Automatically brings it back to the corner.
		digit.Rotate()
	}

	return false
}

type boardState struct {
	filled map[Segment]bool
	size   int
}

func newBoardState(size int) *boardState {
	return &boardState{
		filled: map[Segment]bool{},
		size:   size,
	}
}

func (b *boardState) Size() int {
	return b.size
}

func (b *boardState) CanAdd(d Digit) bool {
	for _, s := range d {
		for _, l := range s {
			for _, c := range l {
				if c < 0 || c > b.size {
					return false
				}
			}
		}
		if b.filled[s] {
			return false
		}
	}

	return true
}

func (b *boardState) Add(d Digit) {
	for _, s := range d {
		b.filled[s] = true
	}
}

func (b *boardState) Remove(d Digit) {
	for _, s := range d {
		b.filled[s] = false
	}
}
