package toolbox3d

import "github.com/unixpickle/model3d/model3d"

// LineJoin creates a Solid containing all points within a
// distance d of any line segments in a list.
func LineJoin(r float64, lines ...model3d.Segment) model3d.Solid {
	m := model3d.NewMesh()
	for _, l := range lines {
		m.Add(&model3d.Triangle{l[0], l[1], l[1]})
	}
	collider := model3d.MeshToCollider(m)
	return model3d.NewColliderSolidHollow(collider, r)
}
