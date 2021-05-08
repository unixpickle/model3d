// Generated from templates/coord_map.template

package model2d

// CoordMap implements a map-like interface for
// mapping Coord to interface{}.
//
// This can be more efficient than using a map directly,
// since it uses a special hash function for coordinates.
// The speed-up is variable, but was ~2x as of mid-2021.
type CoordMap struct {
	slowMap map[Coord]interface{}
	fastMap map[uint64]cellForCoordMap
}

// NewCoordMap creates an empty map.
func NewCoordMap() *CoordMap {
	return &CoordMap{fastMap: map[uint64]cellForCoordMap{}}
}

// Len gets the number of elements in the map.
func (m *CoordMap) Len() int {
	if m.fastMap != nil {
		return len(m.fastMap)
	} else {
		return len(m.slowMap)
	}
}

// Load gets the value for the given key.
//
// If no value is present, the first return argument is a
// zero value, and the second is false. Otherwise, the
// second return value is true.
func (m *CoordMap) Load(key Coord) (interface{}, bool) {
	if m.fastMap != nil {
		cell, ok := m.fastMap[key.fastHash64()]
		if !ok || cell.Key != key {
			return nil, false
		}
		return cell.Value, true
	} else {
		x, y := m.slowMap[key]
		return x, y
	}
}

// Delete removes the key from the map if it exists, and
// does nothing otherwise.
func (m *CoordMap) Delete(key Coord) {
	if m.fastMap != nil {
		hash := key.fastHash64()
		if cell, ok := m.fastMap[hash]; ok && cell.Key == key {
			delete(m.fastMap, hash)
		}
	} else {
		delete(m.slowMap, key)
	}
}

// Store assigns the value to the given key, overwriting
// the previous value for the key if necessary.
func (m *CoordMap) Store(key Coord, value interface{}) {
	if m.fastMap != nil {
		hash := key.fastHash64()
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			m.slowMap[key] = value
		} else {
			m.fastMap[hash] = cellForCoordMap{Key: key, Value: value}
		}
	} else {
		m.slowMap[key] = value
	}
}

// Range iterates over the map, calling f successively for
// each value until it returns false, or all entries are
// enumerated.
//
// It is not safe to modify the map with Store or Delete
// during enumeration.
func (m *CoordMap) Range(f func(key Coord, value interface{}) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key, cell.Value) {
				return
			}
		}
	} else {
		for k, v := range m.slowMap {
			if !f(k, v) {
				return
			}
		}
	}
}

func (m *CoordMap) fastToSlow() {
	m.slowMap = map[Coord]interface{}{}
	for _, cell := range m.fastMap {
		m.slowMap[cell.Key] = cell.Value
	}
	m.fastMap = nil
}

type cellForCoordMap struct {
	Key   Coord
	Value interface{}
}

// CoordToFaces implements a map-like interface for
// mapping Coord to []*Segment.
//
// This can be more efficient than using a map directly,
// since it uses a special hash function for coordinates.
// The speed-up is variable, but was ~2x as of mid-2021.
type CoordToFaces struct {
	slowMap map[Coord][]*Segment
	fastMap map[uint64]cellForCoordToFaces
}

// NewCoordToFaces creates an empty map.
func NewCoordToFaces() *CoordToFaces {
	return &CoordToFaces{fastMap: map[uint64]cellForCoordToFaces{}}
}

// Len gets the number of elements in the map.
func (m *CoordToFaces) Len() int {
	if m.fastMap != nil {
		return len(m.fastMap)
	} else {
		return len(m.slowMap)
	}
}

// Load gets the value for the given key.
//
// If no value is present, the first return argument is a
// zero value, and the second is false. Otherwise, the
// second return value is true.
func (m *CoordToFaces) Load(key Coord) ([]*Segment, bool) {
	if m.fastMap != nil {
		cell, ok := m.fastMap[key.fastHash64()]
		if !ok || cell.Key != key {
			return nil, false
		}
		return cell.Value, true
	} else {
		x, y := m.slowMap[key]
		return x, y
	}
}

// Delete removes the key from the map if it exists, and
// does nothing otherwise.
func (m *CoordToFaces) Delete(key Coord) {
	if m.fastMap != nil {
		hash := key.fastHash64()
		if cell, ok := m.fastMap[hash]; ok && cell.Key == key {
			delete(m.fastMap, hash)
		}
	} else {
		delete(m.slowMap, key)
	}
}

// Store assigns the value to the given key, overwriting
// the previous value for the key if necessary.
func (m *CoordToFaces) Store(key Coord, value []*Segment) {
	if m.fastMap != nil {
		hash := key.fastHash64()
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			m.slowMap[key] = value
		} else {
			m.fastMap[hash] = cellForCoordToFaces{Key: key, Value: value}
		}
	} else {
		m.slowMap[key] = value
	}
}

// Range iterates over the map, calling f successively for
// each value until it returns false, or all entries are
// enumerated.
//
// It is not safe to modify the map with Store or Delete
// during enumeration.
func (m *CoordToFaces) Range(f func(key Coord, value []*Segment) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key, cell.Value) {
				return
			}
		}
	} else {
		for k, v := range m.slowMap {
			if !f(k, v) {
				return
			}
		}
	}
}

func (m *CoordToFaces) fastToSlow() {
	m.slowMap = map[Coord][]*Segment{}
	for _, cell := range m.fastMap {
		m.slowMap[cell.Key] = cell.Value
	}
	m.fastMap = nil
}

type cellForCoordToFaces struct {
	Key   Coord
	Value []*Segment
}
