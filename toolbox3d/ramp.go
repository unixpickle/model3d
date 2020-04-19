package toolbox3d

import (
	"github.com/unixpickle/model3d/model3d"
)

// A Ramp wraps an existing solid and gradually increases
// the scale of the solid from 0% to 100% along a given
// axis.
//
// This makes it easier to make shapes like upside-down
// pyramids and cones for use in FDM printing without
// supports.
type Ramp struct {
	model3d.Solid

	// P1 is the tip of the ramp, where the scale is 0.
	// Any point further in the direction of P1 will have
	// a scale of zero.
	P1 model3d.Coord3D

	// P2 is the base of the ramp, where the scale is 1.
	// Any point further in the direction of P2 will have
	// a scale of one.
	P2 model3d.Coord3D
}

func (r *Ramp) Contains(c model3d.Coord3D) bool {
	axis := r.P2.Sub(r.P1)
	v := c.Sub(r.P1)
	scale := axis.Dot(v)
	if scale < 0 {
		return false
	}

	norm := axis.Norm()
	scale /= norm * norm

	if scale >= 1 {
		return r.Solid.Contains(c)
	}

	scaled := v.Sub(axis.Scale(scale)).Scale(1 / scale).Add(axis.Scale(scale)).Add(r.P1)
	return r.Solid.Contains(scaled)
}
