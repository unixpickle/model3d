package toolbox3d

import "github.com/unixpickle/model3d/model3d"

func singularVertexFamilies(m *model3d.Mesh, v model3d.Coord3D) [][]*model3d.Triangle {
	var families [][]*model3d.Triangle
	tris := m.Find(v)
	for len(tris) > 0 {
		var family []*model3d.Triangle
		family, tris = singularVertexNextFamily(m, tris)
		families = append(families, family)
	}
	return families
}

func singularVertexNextFamily(m *model3d.Mesh, tris []*model3d.Triangle) (family,
	leftover []*model3d.Triangle) {
	// See mesh.SingularVertices() for an explanation of
	// this algorithm.

	queue := make([]int, len(tris))
	queue[0] = 1
	changed := true
	numVisited := 1
	for changed {
		changed = false
		for i, status := range queue {
			if status != 1 {
				continue
			}
			t := tris[i]
			for j, t1 := range tris {
				if queue[j] == 0 && t.SharesEdge(t1) {
					queue[j] = 1
					numVisited++
					changed = true
				}
			}
			queue[i] = 2
		}
	}
	if numVisited == len(tris) {
		return tris, nil
	} else {
		for i, status := range queue {
			if status == 0 {
				leftover = append(leftover, tris[i])
			} else {
				family = append(family, tris[i])
			}
		}
		return
	}
}
