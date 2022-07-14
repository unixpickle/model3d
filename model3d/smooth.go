package model3d

// A MeshSmoother uses gradient descent to smooth out the
// surface of a mesh by minimizing surface area.
//
// The smoother can be constrained to discourage vertices
// from moving far from their origins, making the surface
// locally smooth without greatly modifying the volume.
type MeshSmoother struct {
	// StepSize controls how fast the mesh is updated.
	// A good value will depend on the mesh, but a good
	// default is 0.1.
	StepSize float64

	// Iterations controls the number of gradient steps
	// to take in order to smooth a mesh.
	// More values result in smoother meshes but take
	// more time.
	Iterations int

	// ConstraintDistance is the minimum distance after
	// which the origin constraint will take effect.
	// This allows points to move freely a little bit
	// without being constrained at all, which is good
	// for sharp meshes created from voxel grids.
	//
	// If this is 0, then points will always be pulled
	// towards their origin by a factor of
	// ConstraintWeight.
	ConstraintDistance float64

	// ConstraintWeight is the weight of the distance
	// constraint term, ||x-x0||^2.
	// If this is 0, no constraint is applied.
	ConstraintWeight float64

	// ConstraintFunc, if specified, is a totally custom
	// gradient term that constrains points to their
	// original positions.
	//
	// The return value of the function is added to the
	// gradient at every step.
	//
	// This is independent of ConstraintDistance and
	// ConstraintWeight, which can be used simultaneously.
	ConstraintFunc func(origin, newCoord Coord3D) Coord3D

	// HardConstraintFunc, if non-nil, is a function that
	// returns true for all of the initial points that
	// should not be modified at all.
	HardConstraintFunc func(origin Coord3D) bool
}

// Smooth applies gradient descent to smooth the mesh.
func (m *MeshSmoother) Smooth(mesh *Mesh) *Mesh {
	im := newIndexMesh(mesh)
	origins := append([]Coord3D{}, im.Coords...)
	newCoords := append([]Coord3D{}, im.Coords...)

	// List of coordinate indices to never change.
	var hardConstraints []int
	if m.HardConstraintFunc != nil {
		for i, c := range im.Coords {
			if m.HardConstraintFunc(c) {
				hardConstraints = append(hardConstraints, i)
			}
		}
	}

	for step := 0; step < m.Iterations; step++ {
		if m.ConstraintWeight != 0 {
			for i, c := range newCoords {
				d := origins[i].Sub(c)
				if m.ConstraintDistance > 0 {
					norm := d.Norm()
					if norm <= m.ConstraintDistance {
						continue
					}
					d = d.Scale((norm - m.ConstraintDistance) / norm)
				}
				newCoords[i] = c.Add(d.Scale(2 * m.ConstraintWeight * m.StepSize))
			}
		}
		if m.ConstraintFunc != nil {
			for i, c := range newCoords {
				grad := m.ConstraintFunc(origins[i], c)
				newCoords[i] = c.Add(grad.Scale(m.StepSize))
			}
		}
		for i := range im.Triangles {
			indexTri := im.Triangles[i]
			t := im.Triangle(i)
			for i, grad := range t.AreaGradient() {
				j := indexTri[i]
				newCoords[j] = newCoords[j].Add(grad.Scale(-m.StepSize))
			}
		}
		if hardConstraints != nil {
			for _, i := range hardConstraints {
				newCoords[i] = im.Coords[i]
			}
		}
		copy(im.Coords, newCoords)
	}

	return im.Mesh()
}

// VoxelSmoother uses hard-constraints on top of gradient
// descent to minimize the surface area of a mesh while
// keeping vertices within a square region of space.
//
// This is based on the surface nets algorithm.
//
// Also see MeshSmoother, which is similar but more
// general-purpose.
type VoxelSmoother struct {
	StepSize   float64
	Iterations int

	// MaxDistance is the maximum L_infinity distance a
	// vertex must move.
	MaxDistance float64
}

// Smooth applies gradient descent to smooth the mesh.
func (v *VoxelSmoother) Smooth(mesh *Mesh) *Mesh {
	im := newIndexMesh(mesh)
	origins := append([]Coord3D{}, im.Coords...)
	newCoords := append([]Coord3D{}, im.Coords...)
	for step := 0; step < v.Iterations; step++ {
		for i := range im.Triangles {
			indexTri := im.Triangles[i]
			t := im.Triangle(i)
			for i, grad := range t.AreaGradient() {
				j := indexTri[i]
				newCoords[j] = newCoords[j].Add(grad.Scale(-v.StepSize))
			}
		}
		for i, c := range newCoords {
			o := origins[i]
			constraint := XYZ(v.MaxDistance, v.MaxDistance, v.MaxDistance)
			c = c.Max(o.Sub(constraint))
			c = c.Min(o.Add(constraint))
			newCoords[i] = c
			im.Coords[i] = c
		}
		copy(im.Coords, newCoords)
	}

	return im.Mesh()
}

type indexMesh struct {
	Coords    []Coord3D
	Triangles [][3]int
}

func newIndexMesh(m *Mesh) *indexMesh {
	capacity := len(m.faces) * 3
	if v2t := m.getVertexToFaceOrNil(); v2t != nil {
		capacity = v2t.Len()
	}

	res := &indexMesh{
		Coords:    make([]Coord3D, 0, capacity),
		Triangles: make([][3]int, 0, len(m.faces)),
	}
	coordToIdx := NewCoordMap[int]()

	m.Iterate(func(t *Triangle) {
		var triangle [3]int
		for i, c := range t {
			idx, ok := coordToIdx.Load(c)
			if !ok {
				idx = len(res.Coords)
				coordToIdx.Store(c, idx)
				res.Coords = append(res.Coords, c)
			}
			triangle[i] = idx
		}
		res.Triangles = append(res.Triangles, triangle)
	})

	return res
}

func (i *indexMesh) Triangle(idx int) Triangle {
	var t Triangle
	for j, ci := range i.Triangles[idx] {
		t[j] = i.Coords[ci]
	}
	return t
}

func (i *indexMesh) Mesh() *Mesh {
	m := NewMesh()
	for j := range i.Triangles {
		t := i.Triangle(j)
		m.Add(&t)
	}
	return m
}
