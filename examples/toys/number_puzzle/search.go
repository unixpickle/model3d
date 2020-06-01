package main

// SearchPlacement finds a transformed version of each
// digit such that all of the digits fit on the board at
// once.
func SearchPlacement(digits []Digit, boardSize int) []Digit {
	state := newBoardState(boardSize)
	return searchRecursively(digits, state)
}

func searchRecursively(digits []Digit, state *boardState) []Digit {
	if len(digits) == 0 {
		return []Digit{}
	}
	digit := digits[0]
	remaining := digits[1:]
	for i := 0; i < 4; i++ {
		for x := 0; x <= state.Size(); x++ {
			for y := 0; y <= state.Size(); y++ {
				translated := digit.Translate(Location{y, x})
				if state.CanAdd(translated) {
					state.Add(translated)
					if res := searchRecursively(remaining, state); res != nil {
						return append([]Digit{translated}, res...)
					}
					state.Remove(translated)
				}
			}
		}

		digit = digit.Rotate()
	}

	return nil
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
