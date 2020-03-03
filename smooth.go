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
	// for meshes created with SolidToMesh().
	//
	// If this is 0, then points will always be pulled
	// towards their origin by a factor of
	// ConstraintWeight.
	ConstraintDistance float64

	// ConstraintWeight is the weight of the distance
	// constraint term, ||x-x0||^2.
	// If this is 0, no constraint is applied.
	ConstraintWeight float64
}

// Smooth applies gradient descent to smooth the mesh.
func (m *MeshSmoother) Smooth(mesh *Mesh) *Mesh {
	im := newIndexMesh(mesh)
	origins := append([]Coord3D{}, im.Coords...)
	newCoords := append([]Coord3D{}, im.Coords...)
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
		for i := range im.Triangles {
			indexTri := im.Triangles[i]
			t := im.Triangle(i)
			for i, grad := range t.AreaGradient() {
				j := indexTri[i]
				newCoords[j] = newCoords[j].Add(grad.Scale(-m.StepSize))
			}
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
	capacity := len(m.triangles) * 3
	if m.vertexToTriangle != nil {
		capacity = len(m.vertexToTriangle)
	}

	res := &indexMesh{
		Coords:    make([]Coord3D, 0, capacity),
		Triangles: make([][3]int, 0, len(m.triangles)),
	}
	coordToIdx := make(map[Coord3D]int, capacity)

	m.Iterate(func(t *Triangle) {
		var triangle [3]int
		for i, c := range t {
			idx, ok := coordToIdx[c]
			if !ok {
				idx = len(res.Coords)
				coordToIdx[c] = idx
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