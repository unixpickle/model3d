package toolbox3d

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
)

type Axis int

const (
	AxisX = 0
	AxisY = 1
	AxisZ = 2
)

// AxisSqueeze is a coordinate transformation which
// squeezes some section of space into a much smaller
// amount of space along some axis.
//
// AxisSqueeze can be used to efficiently produce meshes
// which are mostly uniform along some axis, for example
// a tall cylinder.
type AxisSqueeze struct {
	// The axis to compress.
	Axis Axis

	// Bounds on the axis to compress.
	Min float64
	Max float64

	// This is (new length / old length).
	// For example, if we use a squeeze ratio of 0.1,
	// then squeezing 2 inches will bring it down to
	// 0.2 inches.
	Ratio float64
}

// Apply squeezes the coordinate.
func (a *AxisSqueeze) Apply(c model3d.Coord3D) model3d.Coord3D {
	arr := c.Array()

	if arr[a.Axis] < a.Min {
		return c
	}

	if arr[a.Axis] > a.Max {
		arr[a.Axis] -= (a.Max - a.Min) * (1 - a.Ratio)
	} else {
		arr[a.Axis] -= (arr[a.Axis] - a.Min) * (1 - a.Ratio)
	}

	return model3d.NewCoord3DArray(arr)
}

// ApplyBounds squeezes the bounds.
func (a *AxisSqueeze) ApplyBounds(min, max model3d.Coord3D) (newMin, newMax model3d.Coord3D) {
	return a.Apply(min), a.Apply(max)
}

// Inverse creates an AxisSqueeze that undoes the squeeze
// performed by a.
func (a *AxisSqueeze) Inverse() model3d.Transform {
	return &AxisSqueeze{
		Axis:  a.Axis,
		Min:   a.Min,
		Max:   a.Min + (a.Max-a.Min)*a.Ratio,
		Ratio: 1 / a.Ratio,
	}
}

// AxisPinch is similar to AxisSqueeze, except that it
// does not affect space outside of the pinched region.
//
// Coordinates within the pinched region are pulled to the
// center of the region by following some polynomial.
//
// AxisPinch can be used to prevent jagged edges on the
// tops and bottoms of marching cubes solids, by pinching
// the uneven edges into a much flatter surface.
type AxisPinch struct {
	// The axis to pinch along.
	Axis Axis

	// Bounds on the axis to pinch.
	Min float64
	Max float64

	// Power controls how much pinching is performed.
	// A Power of 1 means no pinching.
	// Higher powers perform spatial pinching.
	// Lower powers un-pinch a region, moving coordinates
	// further from the center.
	// Reciprocal powers undo each other.
	Power float64
}

// Apply pinches the coordinate.
func (a *AxisPinch) Apply(c model3d.Coord3D) model3d.Coord3D {
	arr := c.Array()

	if arr[a.Axis] < a.Min || arr[a.Axis] > a.Max {
		return c
	}

	center := (a.Min + a.Max) / 2
	scale := (a.Max - a.Min) / 2
	t := (arr[a.Axis] - center) / scale
	negative := t < 0
	if negative {
		t = -t
	}
	t = math.Pow(t, a.Power)
	if negative {
		t = -t
	}
	arr[a.Axis] = t*scale + center

	return model3d.NewCoord3DArray(arr)
}

// ApplyBounds pinches the bounds.
func (a *AxisPinch) ApplyBounds(min, max model3d.Coord3D) (newMin, newMax model3d.Coord3D) {
	return a.Apply(min), a.Apply(max)
}

// Inverse creates an AxisPinch that undoes the pinch
// performed by a.
func (a *AxisPinch) Inverse() model3d.Transform {
	return &AxisPinch{
		Axis:  a.Axis,
		Min:   a.Min,
		Max:   a.Max,
		Power: 1 / a.Power,
	}
}
