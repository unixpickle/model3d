package toolbox3d

import "github.com/unixpickle/model3d"

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

// ApplySolid squeezes a solid.
func (a *AxisSqueeze) ApplySolid(s model3d.Solid) model3d.Solid {
	return &axisSqueezeSolid{
		min:     a.Apply(s.Min()),
		max:     a.Apply(s.Max()),
		solid:   s,
		inverse: a.Inverse(),
	}
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

type axisSqueezeSolid struct {
	min     model3d.Coord3D
	max     model3d.Coord3D
	solid   model3d.Solid
	inverse model3d.Transform
}

func (a *axisSqueezeSolid) Min() model3d.Coord3D {
	return a.min
}

func (a *axisSqueezeSolid) Max() model3d.Coord3D {
	return a.max
}

func (a *axisSqueezeSolid) Contains(c model3d.Coord3D) bool {
	return a.solid.Contains(a.inverse.Apply(c))
}
