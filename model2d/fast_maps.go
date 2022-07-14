// Generated from templates/fast_maps.template

package model2d

type Adder interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | ~complex64 | ~complex128
}

// CoordMap implements a map-like interface for
// mapping Coord to T.
//
// This can be more efficient than using a map directly,
// since it uses a special hash function for coordinates.
// The speed-up is variable, but was ~2x as of mid-2021.
type CoordMap[T any] struct {
	slowMap map[Coord]T
	fastMap map[uint64]cellForCoordMap[T]
}

// NewCoordMap creates an empty map.
func NewCoordMap[T any]() *CoordMap[T] {
	return &CoordMap[T]{fastMap: map[uint64]cellForCoordMap[T]{}}
}

// Len gets the number of elements in the map.
func (m *CoordMap[T]) Len() int {
	if m.fastMap != nil {
		return len(m.fastMap)
	} else {
		return len(m.slowMap)
	}
}

// Value is like Load(), but without a second return
// value.
func (m *CoordMap[T]) Value(key Coord) T {
	res, _ := m.Load(key)
	return res
}

// Load gets the value for the given key.
//
// If no value is present, the first return argument is a
// zero value, and the second is false. Otherwise, the
// second return value is true.
func (m *CoordMap[T]) Load(key Coord) (T, bool) {
	if m.fastMap != nil {
		cell, ok := m.fastMap[hashForCoordMap(key)]
		if !ok || cell.Key != key {
			return zeroForCoordMap[T](), false
		}
		return cell.Value, true
	} else {
		x, y := m.slowMap[key]
		return x, y
	}
}

// Delete removes the key from the map if it exists, and
// does nothing otherwise.
func (m *CoordMap[T]) Delete(key Coord) {
	if m.fastMap != nil {
		hash := hashForCoordMap(key)
		if cell, ok := m.fastMap[hash]; ok && cell.Key == key {
			delete(m.fastMap, hash)
		}
	} else {
		delete(m.slowMap, key)
	}
}

// Store assigns the value to the given key, overwriting
// the previous value for the key if necessary.
func (m *CoordMap[T]) Store(key Coord, value T) {
	if m.fastMap != nil {
		hash := hashForCoordMap(key)
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			m.slowMap[key] = value
		} else {
			m.fastMap[hash] = cellForCoordMap[T]{Key: key, Value: value}
		}
	} else {
		m.slowMap[key] = value
	}
}

// KeyRange is like Range, but only iterates over
// keys, not values.
func (m *CoordMap[T]) KeyRange(f func(key Coord) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key) {
				return
			}
		}
	} else {
		for k := range m.slowMap {
			if !f(k) {
				return
			}
		}
	}
}

// ValueRange is like Range, but only iterates over
// values only.
func (m *CoordMap[T]) ValueRange(f func(value T) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Value) {
				return
			}
		}
	} else {
		for _, v := range m.slowMap {
			if !f(v) {
				return
			}
		}
	}
}

// Range iterates over the map, calling f successively for
// each value until it returns false, or all entries are
// enumerated.
//
// It is not safe to modify the map with Store or Delete
// during enumeration.
func (m *CoordMap[T]) Range(f func(key Coord, value T) bool) {
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

func (m *CoordMap[T]) fastToSlow() {
	m.slowMap = map[Coord]T{}
	for _, cell := range m.fastMap {
		m.slowMap[cell.Key] = cell.Value
	}
	m.fastMap = nil
}

type cellForCoordMap[T any] struct {
	Key   Coord
	Value T
}

func hashForCoordMap(c Coord) uint64 {
	return c.fastHash64()
}

func zeroForCoordMap[T any]() T {
	var e T
	return e
}

// CoordToSlice implements a map-like interface for
// mapping Coord to []T.
//
// This can be more efficient than using a map directly,
// since it uses a special hash function for coordinates.
// The speed-up is variable, but was ~2x as of mid-2021.
type CoordToSlice[T any] struct {
	slowMap map[Coord][]T
	fastMap map[uint64]cellForCoordToSlice[T]
}

// NewCoordToSlice creates an empty map.
func NewCoordToSlice[T any]() *CoordToSlice[T] {
	return &CoordToSlice[T]{fastMap: map[uint64]cellForCoordToSlice[T]{}}
}

// Len gets the number of elements in the map.
func (m *CoordToSlice[T]) Len() int {
	if m.fastMap != nil {
		return len(m.fastMap)
	} else {
		return len(m.slowMap)
	}
}

// Value is like Load(), but without a second return
// value.
func (m *CoordToSlice[T]) Value(key Coord) []T {
	res, _ := m.Load(key)
	return res
}

// Load gets the value for the given key.
//
// If no value is present, the first return argument is a
// zero value, and the second is false. Otherwise, the
// second return value is true.
func (m *CoordToSlice[T]) Load(key Coord) ([]T, bool) {
	if m.fastMap != nil {
		cell, ok := m.fastMap[hashForCoordToSlice(key)]
		if !ok || cell.Key != key {
			return zeroForCoordToSlice[T](), false
		}
		return cell.Value, true
	} else {
		x, y := m.slowMap[key]
		return x, y
	}
}

// Delete removes the key from the map if it exists, and
// does nothing otherwise.
func (m *CoordToSlice[T]) Delete(key Coord) {
	if m.fastMap != nil {
		hash := hashForCoordToSlice(key)
		if cell, ok := m.fastMap[hash]; ok && cell.Key == key {
			delete(m.fastMap, hash)
		}
	} else {
		delete(m.slowMap, key)
	}
}

// Store assigns the value to the given key, overwriting
// the previous value for the key if necessary.
func (m *CoordToSlice[T]) Store(key Coord, value []T) {
	if m.fastMap != nil {
		hash := hashForCoordToSlice(key)
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			m.slowMap[key] = value
		} else {
			m.fastMap[hash] = cellForCoordToSlice[T]{Key: key, Value: value}
		}
	} else {
		m.slowMap[key] = value
	}
}

// Append appends x to the value stored for the given key
// and returns the new value.
func (m *CoordToSlice[T]) Append(key Coord, x T) []T {
	if m.fastMap != nil {
		hash := hashForCoordToSlice(key)
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			return m.Append(key, x)
		} else {
			value := append(cell.Value, x)
			m.fastMap[hash] = cellForCoordToSlice[T]{Key: key, Value: value}
			return value
		}
	} else {
		value := append(m.slowMap[key], x)
		m.slowMap[key] = value
		return value
	}
}

// KeyRange is like Range, but only iterates over
// keys, not values.
func (m *CoordToSlice[T]) KeyRange(f func(key Coord) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key) {
				return
			}
		}
	} else {
		for k := range m.slowMap {
			if !f(k) {
				return
			}
		}
	}
}

// ValueRange is like Range, but only iterates over
// values only.
func (m *CoordToSlice[T]) ValueRange(f func(value []T) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Value) {
				return
			}
		}
	} else {
		for _, v := range m.slowMap {
			if !f(v) {
				return
			}
		}
	}
}

// Range iterates over the map, calling f successively for
// each value until it returns false, or all entries are
// enumerated.
//
// It is not safe to modify the map with Store or Delete
// during enumeration.
func (m *CoordToSlice[T]) Range(f func(key Coord, value []T) bool) {
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

func (m *CoordToSlice[T]) fastToSlow() {
	m.slowMap = map[Coord][]T{}
	for _, cell := range m.fastMap {
		m.slowMap[cell.Key] = cell.Value
	}
	m.fastMap = nil
}

type cellForCoordToSlice[T any] struct {
	Key   Coord
	Value []T
}

func hashForCoordToSlice(c Coord) uint64 {
	return c.fastHash64()
}

func zeroForCoordToSlice[T any]() []T {
	var e []T
	return e
}

// CoordToNumber implements a map-like interface for
// mapping Coord to T.
//
// This can be more efficient than using a map directly,
// since it uses a special hash function for coordinates.
// The speed-up is variable, but was ~2x as of mid-2021.
type CoordToNumber[T Adder] struct {
	slowMap map[Coord]T
	fastMap map[uint64]cellForCoordToNumber[T]
}

// NewCoordToNumber creates an empty map.
func NewCoordToNumber[T Adder]() *CoordToNumber[T] {
	return &CoordToNumber[T]{fastMap: map[uint64]cellForCoordToNumber[T]{}}
}

// Len gets the number of elements in the map.
func (m *CoordToNumber[T]) Len() int {
	if m.fastMap != nil {
		return len(m.fastMap)
	} else {
		return len(m.slowMap)
	}
}

// Value is like Load(), but without a second return
// value.
func (m *CoordToNumber[T]) Value(key Coord) T {
	res, _ := m.Load(key)
	return res
}

// Load gets the value for the given key.
//
// If no value is present, the first return argument is a
// zero value, and the second is false. Otherwise, the
// second return value is true.
func (m *CoordToNumber[T]) Load(key Coord) (T, bool) {
	if m.fastMap != nil {
		cell, ok := m.fastMap[hashForCoordToNumber(key)]
		if !ok || cell.Key != key {
			return zeroForCoordToNumber[T](), false
		}
		return cell.Value, true
	} else {
		x, y := m.slowMap[key]
		return x, y
	}
}

// Delete removes the key from the map if it exists, and
// does nothing otherwise.
func (m *CoordToNumber[T]) Delete(key Coord) {
	if m.fastMap != nil {
		hash := hashForCoordToNumber(key)
		if cell, ok := m.fastMap[hash]; ok && cell.Key == key {
			delete(m.fastMap, hash)
		}
	} else {
		delete(m.slowMap, key)
	}
}

// Store assigns the value to the given key, overwriting
// the previous value for the key if necessary.
func (m *CoordToNumber[T]) Store(key Coord, value T) {
	if m.fastMap != nil {
		hash := hashForCoordToNumber(key)
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			m.slowMap[key] = value
		} else {
			m.fastMap[hash] = cellForCoordToNumber[T]{Key: key, Value: value}
		}
	} else {
		m.slowMap[key] = value
	}
}

// Add adds x to the value stored for the given key and
// returns the new value.
func (m *CoordToNumber[T]) Add(key Coord, x T) T {
	if m.fastMap != nil {
		hash := hashForCoordToNumber(key)
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			return m.Add(key, x)
		} else {
			m.fastMap[hash] = cellForCoordToNumber[T]{Key: key, Value: cell.Value + x}
			return cell.Value + x
		}
	} else {
		value := m.slowMap[key] + x
		m.slowMap[key] = value
		return value
	}
}

// KeyRange is like Range, but only iterates over
// keys, not values.
func (m *CoordToNumber[T]) KeyRange(f func(key Coord) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key) {
				return
			}
		}
	} else {
		for k := range m.slowMap {
			if !f(k) {
				return
			}
		}
	}
}

// ValueRange is like Range, but only iterates over
// values only.
func (m *CoordToNumber[T]) ValueRange(f func(value T) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Value) {
				return
			}
		}
	} else {
		for _, v := range m.slowMap {
			if !f(v) {
				return
			}
		}
	}
}

// Range iterates over the map, calling f successively for
// each value until it returns false, or all entries are
// enumerated.
//
// It is not safe to modify the map with Store or Delete
// during enumeration.
func (m *CoordToNumber[T]) Range(f func(key Coord, value T) bool) {
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

func (m *CoordToNumber[T]) fastToSlow() {
	m.slowMap = map[Coord]T{}
	for _, cell := range m.fastMap {
		m.slowMap[cell.Key] = cell.Value
	}
	m.fastMap = nil
}

type cellForCoordToNumber[T Adder] struct {
	Key   Coord
	Value T
}

func hashForCoordToNumber(c Coord) uint64 {
	return c.fastHash64()
}

func zeroForCoordToNumber[T any]() T {
	var e T
	return e
}

// EdgeMap implements a map-like interface for
// mapping [2]Coord to T.
//
// This can be more efficient than using a map directly,
// since it uses a special hash function for coordinates.
// The speed-up is variable, but was ~2x as of mid-2021.
type EdgeMap[T any] struct {
	slowMap map[[2]Coord]T
	fastMap map[uint64]cellForEdgeMap[T]
}

// NewEdgeMap creates an empty map.
func NewEdgeMap[T any]() *EdgeMap[T] {
	return &EdgeMap[T]{fastMap: map[uint64]cellForEdgeMap[T]{}}
}

// Len gets the number of elements in the map.
func (m *EdgeMap[T]) Len() int {
	if m.fastMap != nil {
		return len(m.fastMap)
	} else {
		return len(m.slowMap)
	}
}

// Value is like Load(), but without a second return
// value.
func (m *EdgeMap[T]) Value(key [2]Coord) T {
	res, _ := m.Load(key)
	return res
}

// Load gets the value for the given key.
//
// If no value is present, the first return argument is a
// zero value, and the second is false. Otherwise, the
// second return value is true.
func (m *EdgeMap[T]) Load(key [2]Coord) (T, bool) {
	if m.fastMap != nil {
		cell, ok := m.fastMap[hashForEdgeMap(key)]
		if !ok || cell.Key != key {
			return zeroForEdgeMap[T](), false
		}
		return cell.Value, true
	} else {
		x, y := m.slowMap[key]
		return x, y
	}
}

// Delete removes the key from the map if it exists, and
// does nothing otherwise.
func (m *EdgeMap[T]) Delete(key [2]Coord) {
	if m.fastMap != nil {
		hash := hashForEdgeMap(key)
		if cell, ok := m.fastMap[hash]; ok && cell.Key == key {
			delete(m.fastMap, hash)
		}
	} else {
		delete(m.slowMap, key)
	}
}

// Store assigns the value to the given key, overwriting
// the previous value for the key if necessary.
func (m *EdgeMap[T]) Store(key [2]Coord, value T) {
	if m.fastMap != nil {
		hash := hashForEdgeMap(key)
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			m.slowMap[key] = value
		} else {
			m.fastMap[hash] = cellForEdgeMap[T]{Key: key, Value: value}
		}
	} else {
		m.slowMap[key] = value
	}
}

// KeyRange is like Range, but only iterates over
// keys, not values.
func (m *EdgeMap[T]) KeyRange(f func(key [2]Coord) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key) {
				return
			}
		}
	} else {
		for k := range m.slowMap {
			if !f(k) {
				return
			}
		}
	}
}

// ValueRange is like Range, but only iterates over
// values only.
func (m *EdgeMap[T]) ValueRange(f func(value T) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Value) {
				return
			}
		}
	} else {
		for _, v := range m.slowMap {
			if !f(v) {
				return
			}
		}
	}
}

// Range iterates over the map, calling f successively for
// each value until it returns false, or all entries are
// enumerated.
//
// It is not safe to modify the map with Store or Delete
// during enumeration.
func (m *EdgeMap[T]) Range(f func(key [2]Coord, value T) bool) {
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

func (m *EdgeMap[T]) fastToSlow() {
	m.slowMap = map[[2]Coord]T{}
	for _, cell := range m.fastMap {
		m.slowMap[cell.Key] = cell.Value
	}
	m.fastMap = nil
}

type cellForEdgeMap[T any] struct {
	Key   [2]Coord
	Value T
}

func hashForEdgeMap(c [2]Coord) uint64 {
	h1 := c[0].fastHash()
	h2 := c[1].fastHash()
	return uint64(h1) | (uint64(h2) << 32)
}

func zeroForEdgeMap[T any]() T {
	var e T
	return e
}

// EdgeToSlice implements a map-like interface for
// mapping [2]Coord to []T.
//
// This can be more efficient than using a map directly,
// since it uses a special hash function for coordinates.
// The speed-up is variable, but was ~2x as of mid-2021.
type EdgeToSlice[T any] struct {
	slowMap map[[2]Coord][]T
	fastMap map[uint64]cellForEdgeToSlice[T]
}

// NewEdgeToSlice creates an empty map.
func NewEdgeToSlice[T any]() *EdgeToSlice[T] {
	return &EdgeToSlice[T]{fastMap: map[uint64]cellForEdgeToSlice[T]{}}
}

// Len gets the number of elements in the map.
func (m *EdgeToSlice[T]) Len() int {
	if m.fastMap != nil {
		return len(m.fastMap)
	} else {
		return len(m.slowMap)
	}
}

// Value is like Load(), but without a second return
// value.
func (m *EdgeToSlice[T]) Value(key [2]Coord) []T {
	res, _ := m.Load(key)
	return res
}

// Load gets the value for the given key.
//
// If no value is present, the first return argument is a
// zero value, and the second is false. Otherwise, the
// second return value is true.
func (m *EdgeToSlice[T]) Load(key [2]Coord) ([]T, bool) {
	if m.fastMap != nil {
		cell, ok := m.fastMap[hashForEdgeToSlice(key)]
		if !ok || cell.Key != key {
			return zeroForEdgeToSlice[T](), false
		}
		return cell.Value, true
	} else {
		x, y := m.slowMap[key]
		return x, y
	}
}

// Delete removes the key from the map if it exists, and
// does nothing otherwise.
func (m *EdgeToSlice[T]) Delete(key [2]Coord) {
	if m.fastMap != nil {
		hash := hashForEdgeToSlice(key)
		if cell, ok := m.fastMap[hash]; ok && cell.Key == key {
			delete(m.fastMap, hash)
		}
	} else {
		delete(m.slowMap, key)
	}
}

// Store assigns the value to the given key, overwriting
// the previous value for the key if necessary.
func (m *EdgeToSlice[T]) Store(key [2]Coord, value []T) {
	if m.fastMap != nil {
		hash := hashForEdgeToSlice(key)
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			m.slowMap[key] = value
		} else {
			m.fastMap[hash] = cellForEdgeToSlice[T]{Key: key, Value: value}
		}
	} else {
		m.slowMap[key] = value
	}
}

// Append appends x to the value stored for the given key
// and returns the new value.
func (m *EdgeToSlice[T]) Append(key [2]Coord, x T) []T {
	if m.fastMap != nil {
		hash := hashForEdgeToSlice(key)
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			return m.Append(key, x)
		} else {
			value := append(cell.Value, x)
			m.fastMap[hash] = cellForEdgeToSlice[T]{Key: key, Value: value}
			return value
		}
	} else {
		value := append(m.slowMap[key], x)
		m.slowMap[key] = value
		return value
	}
}

// KeyRange is like Range, but only iterates over
// keys, not values.
func (m *EdgeToSlice[T]) KeyRange(f func(key [2]Coord) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key) {
				return
			}
		}
	} else {
		for k := range m.slowMap {
			if !f(k) {
				return
			}
		}
	}
}

// ValueRange is like Range, but only iterates over
// values only.
func (m *EdgeToSlice[T]) ValueRange(f func(value []T) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Value) {
				return
			}
		}
	} else {
		for _, v := range m.slowMap {
			if !f(v) {
				return
			}
		}
	}
}

// Range iterates over the map, calling f successively for
// each value until it returns false, or all entries are
// enumerated.
//
// It is not safe to modify the map with Store or Delete
// during enumeration.
func (m *EdgeToSlice[T]) Range(f func(key [2]Coord, value []T) bool) {
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

func (m *EdgeToSlice[T]) fastToSlow() {
	m.slowMap = map[[2]Coord][]T{}
	for _, cell := range m.fastMap {
		m.slowMap[cell.Key] = cell.Value
	}
	m.fastMap = nil
}

type cellForEdgeToSlice[T any] struct {
	Key   [2]Coord
	Value []T
}

func hashForEdgeToSlice(c [2]Coord) uint64 {
	h1 := c[0].fastHash()
	h2 := c[1].fastHash()
	return uint64(h1) | (uint64(h2) << 32)
}

func zeroForEdgeToSlice[T any]() []T {
	var e []T
	return e
}

// EdgeToNumber implements a map-like interface for
// mapping [2]Coord to T.
//
// This can be more efficient than using a map directly,
// since it uses a special hash function for coordinates.
// The speed-up is variable, but was ~2x as of mid-2021.
type EdgeToNumber[T Adder] struct {
	slowMap map[[2]Coord]T
	fastMap map[uint64]cellForEdgeToNumber[T]
}

// NewEdgeToNumber creates an empty map.
func NewEdgeToNumber[T Adder]() *EdgeToNumber[T] {
	return &EdgeToNumber[T]{fastMap: map[uint64]cellForEdgeToNumber[T]{}}
}

// Len gets the number of elements in the map.
func (m *EdgeToNumber[T]) Len() int {
	if m.fastMap != nil {
		return len(m.fastMap)
	} else {
		return len(m.slowMap)
	}
}

// Value is like Load(), but without a second return
// value.
func (m *EdgeToNumber[T]) Value(key [2]Coord) T {
	res, _ := m.Load(key)
	return res
}

// Load gets the value for the given key.
//
// If no value is present, the first return argument is a
// zero value, and the second is false. Otherwise, the
// second return value is true.
func (m *EdgeToNumber[T]) Load(key [2]Coord) (T, bool) {
	if m.fastMap != nil {
		cell, ok := m.fastMap[hashForEdgeToNumber(key)]
		if !ok || cell.Key != key {
			return zeroForEdgeToNumber[T](), false
		}
		return cell.Value, true
	} else {
		x, y := m.slowMap[key]
		return x, y
	}
}

// Delete removes the key from the map if it exists, and
// does nothing otherwise.
func (m *EdgeToNumber[T]) Delete(key [2]Coord) {
	if m.fastMap != nil {
		hash := hashForEdgeToNumber(key)
		if cell, ok := m.fastMap[hash]; ok && cell.Key == key {
			delete(m.fastMap, hash)
		}
	} else {
		delete(m.slowMap, key)
	}
}

// Store assigns the value to the given key, overwriting
// the previous value for the key if necessary.
func (m *EdgeToNumber[T]) Store(key [2]Coord, value T) {
	if m.fastMap != nil {
		hash := hashForEdgeToNumber(key)
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			m.slowMap[key] = value
		} else {
			m.fastMap[hash] = cellForEdgeToNumber[T]{Key: key, Value: value}
		}
	} else {
		m.slowMap[key] = value
	}
}

// Add adds x to the value stored for the given key and
// returns the new value.
func (m *EdgeToNumber[T]) Add(key [2]Coord, x T) T {
	if m.fastMap != nil {
		hash := hashForEdgeToNumber(key)
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			return m.Add(key, x)
		} else {
			m.fastMap[hash] = cellForEdgeToNumber[T]{Key: key, Value: cell.Value + x}
			return cell.Value + x
		}
	} else {
		value := m.slowMap[key] + x
		m.slowMap[key] = value
		return value
	}
}

// KeyRange is like Range, but only iterates over
// keys, not values.
func (m *EdgeToNumber[T]) KeyRange(f func(key [2]Coord) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key) {
				return
			}
		}
	} else {
		for k := range m.slowMap {
			if !f(k) {
				return
			}
		}
	}
}

// ValueRange is like Range, but only iterates over
// values only.
func (m *EdgeToNumber[T]) ValueRange(f func(value T) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Value) {
				return
			}
		}
	} else {
		for _, v := range m.slowMap {
			if !f(v) {
				return
			}
		}
	}
}

// Range iterates over the map, calling f successively for
// each value until it returns false, or all entries are
// enumerated.
//
// It is not safe to modify the map with Store or Delete
// during enumeration.
func (m *EdgeToNumber[T]) Range(f func(key [2]Coord, value T) bool) {
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

func (m *EdgeToNumber[T]) fastToSlow() {
	m.slowMap = map[[2]Coord]T{}
	for _, cell := range m.fastMap {
		m.slowMap[cell.Key] = cell.Value
	}
	m.fastMap = nil
}

type cellForEdgeToNumber[T Adder] struct {
	Key   [2]Coord
	Value T
}

func hashForEdgeToNumber(c [2]Coord) uint64 {
	h1 := c[0].fastHash()
	h2 := c[1].fastHash()
	return uint64(h1) | (uint64(h2) << 32)
}

func zeroForEdgeToNumber[T any]() T {
	var e T
	return e
}
