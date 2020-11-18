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

// SmartSqueeze creates a transformation to squeeze a
// model along a certain axis without affecting certain
// regions that don't lend themselves to squeezing.
type SmartSqueeze struct {
	// Axis is the axis to squeeze along.
	Axis Axis

	// SqueezeRatio is the axis squeeze ratio to use.
	SqueezeRatio float64

	// PinchRange specifies how much space should be added
	// before and after a pinch to avoid singularities.
	// Should be small, but larger than the marching cubes
	// epsilon.
	PinchRange float64

	// PinchPower is the power for pinches.
	PinchPower float64

	// Ranges along the axis that cannot be squeezed.
	// Typically ranges should be added via AddUnsqueezable().
	Unsqueezable [][2]float64

	// Values of the axis at which a pinch should be used
	// to flatten plateaus.
	// Typically pinches will be added via AddPinch().
	Pinches []float64
}

// AddUnsqueezable adds a range of axis values in which no
// squeezing should be performed.
func (s *SmartSqueeze) AddUnsqueezable(min, max float64) {
	s.Unsqueezable = append(s.Unsqueezable, [2]float64{min, max})
}

// AddPinch adds an axis value at which the coordinates
// should be squeezed to flatten plateaus.
func (s *SmartSqueeze) AddPinch(val float64) {
	s.Pinches = append(s.Pinches, val)
}

// MachingCubesSearch uses the smart squeeze to convert a
// solid into a mesh efficiently.
//
// In particular, the model is transformed, meshified, and
// then the inverse transformation is applied.
//
// For usage information, see model3d.MarchingCubesSearch.
func (s *SmartSqueeze) MarchingCubesSearch(solid model3d.Solid, delta float64,
	iters int) *model3d.Mesh {
	return model3d.MarchingCubesConj(solid, delta, iters, s.Transform(solid))
}

// Transform creates a transformation for the squeezes and
// pinches, given the bounds of a model.
func (s *SmartSqueeze) Transform(b model3d.Bounder) model3d.Transform {
	min, max := b.Min().Array()[s.Axis], b.Max().Array()[s.Axis]

	if len(s.Pinches) > 0 && s.PinchRange <= 0 {
		panic("pinch range must be greater than zero")
	}

	squeezes := model3d.JoinedTransform{}
	value := min
	for value < max {
		isSqueezable, next := s.checkSqueezed(value)
		next = math.Min(next, max) // cap infinite or very large values.
		if isSqueezable {
			if s.SqueezeRatio <= 0 {
				panic("squeeze ratio must be greater than zero")
			}
			squeezes = append(squeezes, &AxisSqueeze{
				Axis:  s.Axis,
				Min:   value,
				Max:   next,
				Ratio: s.SqueezeRatio,
			})
		}
		value = next
	}
	for _, p := range s.Pinches {
		if s.PinchPower <= 0 {
			panic("pinch power must be greater than zero")
		}
		squeezes = append(squeezes, &AxisPinch{
			Axis:  s.Axis,
			Min:   p - s.PinchRange,
			Max:   p + s.PinchRange,
			Power: s.PinchPower,
		})
	}

	reversed := make(model3d.JoinedTransform, 0, len(squeezes))
	for i := len(squeezes) - 1; i >= 0; i-- {
		reversed = append(reversed, squeezes[i])
	}
	return reversed
}

func (s *SmartSqueeze) checkSqueezed(axisValue float64) (isSqueezable bool, next float64) {
	next = math.Inf(1)

	for _, us := range s.Unsqueezable {
		if axisValue >= us[0] && axisValue < us[1] {
			isSqueezable = false
			next = us[1]
			return
		} else if us[0] > axisValue && us[0] < next {
			next = us[0]
		}
	}

	for _, p := range s.Pinches {
		pStart := p - s.PinchRange
		pEnd := p + s.PinchRange
		if axisValue >= pStart && axisValue < pEnd {
			isSqueezable = false
			next = pEnd
			return
		} else if pStart > axisValue && pStart < next {
			next = pStart
		}
	}

	isSqueezable = true
	return
}
