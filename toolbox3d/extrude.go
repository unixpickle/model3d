package toolbox3d

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

// InsetFunc defines an SDF inset/outset for the Extrude() API.
type InsetFunc interface {
	Inset(minZ, maxZ, z float64) float64
	MinInset() float64
}

// FilletInsetFunc is an InsetFunc for rounded fillets.
// This can be used for the Extrude() API.
type FilletInsetFunc struct {
	// Fillet radius at the minZ part of the shape.
	BottomRadius float64

	// Fillet radius at the maxZ part of the shape.
	TopRadius float64

	// If true, extend the solid outwards instead of inwards.
	// This is useful for cutouts.
	Outwards bool
}

func (f *FilletInsetFunc) MinInset() float64 {
	if f.Outwards {
		return -(math.Max(f.BottomRadius, f.TopRadius))
	} else {
		return 0
	}
}

func (f *FilletInsetFunc) Inset(minZ, maxZ, z float64) float64 {
	if z-minZ < f.BottomRadius {
		return f.insetAtZFrac(f.BottomRadius, (z-minZ)/f.BottomRadius)
	} else if maxZ-z < f.TopRadius {
		return f.insetAtZFrac(f.TopRadius, (maxZ-z)/f.TopRadius)
	} else {
		return 0
	}
}

func (f *FilletInsetFunc) insetAtZFrac(r, frac float64) float64 {
	if f.Outwards {
		x := math.Max(0, math.Min(1, frac)) - 1
		return r * (math.Sqrt(1-x*x) - 1)
	} else {
		x := math.Max(0, math.Min(1, frac)) - 1
		return r * (1 - math.Sqrt(1-x*x))
	}
}

// Extrude turns a 2D shape into a 3D shape by extending it along the Z axis.
//
// An inset function indicates how much the shape should be inset/outset at
// each z value, allowing fillets and chamfers.
//
// If you do not need an inset, model3d.ProfileSolid is likely more suitable.
func Extrude(shape model2d.SDF, minZ, maxZ float64, inset InsetFunc) model3d.Solid {
	min := shape.Min()
	max := shape.Max()
	minInset := inset.MinInset()
	min = min.Add(model2d.Ones(minInset))
	max = max.Sub(model2d.Ones(minInset))
	if min.X > max.X || min.Y > max.Y || minZ > maxZ || math.IsNaN(minInset) {
		panic("bounds of extruded SDF are invalid")
	}
	return model3d.CheckedFuncSolid(
		model3d.XYZ(min.X, min.Y, minZ),
		model3d.XYZ(max.X, max.Y, maxZ),
		func(c model3d.Coord3D) bool {
			sdfValue := shape.SDF(c.XY())
			return sdfValue > inset.Inset(minZ, maxZ, c.Z)
		},
	)
}
