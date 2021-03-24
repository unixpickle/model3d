package toolbox3d

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
)

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

// L1LineJoin is like LineJoin, but uses L1 distance.
// All of the segments must be axis aligned.
func L1LineJoin(r float64, lines ...model3d.Segment) model3d.Solid {
	res := model3d.JoinedSolid{}
	for _, seg := range lines {
		res = append(
			res,
			TriangularLine(r, seg[0], seg[1]),
			TriangularBall(r, seg[0]),
			TriangularBall(r, seg[1]),
		)
	}
	return res.Optimize()
}

// TriangularPolygon is similar to L1LineJoin, but only
// adds L1 balls to the connections between segments,
// rather than to all endpoints.
//
// If close is true, then the last point is connected to
// the first. Otherwise, the endpoints are "cut off", i.e.
// points past the endpoints will not be inside the solid,
// even if they are within thickness distance of the
// endpoint. To achieve "rounded" tips, use
// TriangularBall() at the endpoints.
func TriangularPolygon(thickness float64, close bool, p ...model3d.Coord3D) model3d.Solid {
	res := model3d.JoinedSolid{}
	for i := 0; i < len(p)-1; i++ {
		res = append(res, TriangularLine(thickness, p[i], p[i+1]))
		if i != 0 {
			res = append(res, TriangularBall(thickness, p[i]))
		}
	}
	if close {
		res = append(
			res,
			TriangularLine(thickness, p[len(p)-1], p[0]),
			TriangularBall(thickness, p[len(p)-1]),
			TriangularBall(thickness, p[0]),
		)
	}
	return res.Optimize()
}

// TriangularLine creates a solid that is true within a
// given L1 distance of an axis-aligned segment.
//
// The solid is false past the endpoints of the segment.
// To "smooth" the endpoints, use TriangularBall().
func TriangularLine(thickness float64, p1, p2 model3d.Coord3D) model3d.Solid {
	dir := p1.Sub(p2)
	length := dir.Norm()
	dir = dir.Normalize()

	// Basis vectors will be axis-aligned if dir is.
	b1, b2 := dir.OrthoBasis()

	ball := model3d.XYZ(thickness, thickness, thickness)
	return model3d.CheckedFuncSolid(
		p1.Min(p2).Sub(ball),
		p1.Max(p2).Add(ball),
		func(c model3d.Coord3D) bool {
			subtracted := c.Sub(p2)
			dot := subtracted.Dot(dir)
			if dot < 0 || dot > length {
				return false
			}
			dot1, dot2 := b1.Dot(subtracted), b2.Dot(subtracted)
			return math.Abs(dot1)+math.Abs(dot2) < thickness
		},
	)
}

// TriangularBall creates a solid that is true within a
// given L1 distance from a point p.
func TriangularBall(thickness float64, p model3d.Coord3D) model3d.Solid {
	ball := model3d.XYZ(thickness, thickness, thickness)
	return model3d.CheckedFuncSolid(
		p.Sub(ball),
		p.Add(ball),
		func(c model3d.Coord3D) bool {
			diff := c.Sub(p)
			return math.Abs(diff.X)+math.Abs(diff.Y)+math.Abs(diff.Z) < thickness
		},
	)
}
