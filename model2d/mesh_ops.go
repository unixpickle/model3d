package model2d

// Blur moves each vertex closer to the average of its
// neighbors.
//
// The rate argument controls how much the vertices move.
// If it is 1, then the vertices become the average of
// their neighbors. If it is 0, then the vertices remain
// where they are.
func (m *Mesh) Blur(rate float64) *Mesh {
	return m.MapCoords(func(c Coord) Coord {
		count := 0.0
		sum := Coord{}
		for _, s := range m.Find(c) {
			for _, c1 := range s {
				if c1 != c {
					sum = sum.Add(c1)
					count++
				}
			}
		}
		return c.Scale(1 - rate).Add(sum.Scale(rate / count))
	})
}