package toolbox3d

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
)

// ClampAxis creates a solid which does not extend beyond
// the range [min, max] along the given axis.
func ClampAxis(s model3d.Solid, axis Axis, min, max float64) model3d.Solid {
	newMin := s.Min().Array()
	newMax := s.Max().Array()
	newMin[axis] = math.Max(min, newMin[axis])
	newMax[axis] = math.Min(max, newMax[axis])
	if newMin[axis] > newMax[axis] {
		// The solid has been over-constrained.
		//
		// We must still satisfy model3d.BoundsValid().
		// Also, APIs like JoinedSolid don't know about empty
		// solids, so they will still expand their bounding box
		// to include the min/max points we choose.
		center := model3d.NewCoord3DArray(newMin).Mid(model3d.NewCoord3DArray(newMax))
		return model3d.CheckedFuncSolid(center, center, func(c model3d.Coord3D) bool {
			return false
		})
	}
	return model3d.CheckedFuncSolid(
		model3d.NewCoord3DArray(newMin),
		model3d.NewCoord3DArray(newMax),
		func(c model3d.Coord3D) bool {
			return s.Contains(c)
		},
	)
}

// ClampAxisMax clamps the solid below a given maximum
// value along a given axis.
func ClampAxisMax(s model3d.Solid, axis Axis, max float64) model3d.Solid {
	return ClampAxis(s, axis, math.Inf(-1), max)
}

// ClampAxisMin clamps the solid above a given minimum
// value along a given axis.
func ClampAxisMin(s model3d.Solid, axis Axis, min float64) model3d.Solid {
	return ClampAxis(s, axis, min, math.Inf(1))
}

// ClampXMax is like ClampAxisMax for the X axis.
func ClampXMax(s model3d.Solid, max float64) model3d.Solid {
	return ClampAxisMax(s, AxisX, max)
}

// ClampXMin is like ClampAxisMin for the X axis.
func ClampXMin(s model3d.Solid, max float64) model3d.Solid {
	return ClampAxisMin(s, AxisX, max)
}

// ClampYMax is like ClampAxisMax for the Y axis.
func ClampYMax(s model3d.Solid, max float64) model3d.Solid {
	return ClampAxisMax(s, AxisY, max)
}

// ClampYMin is like ClampAxisMin for the Y axis.
func ClampYMin(s model3d.Solid, max float64) model3d.Solid {
	return ClampAxisMin(s, AxisY, max)
}

// ClampZMax is like ClampAxisMax for the Z axis.
func ClampZMax(s model3d.Solid, max float64) model3d.Solid {
	return ClampAxisMax(s, AxisZ, max)
}

// ClampZMin is like ClampAxisMin for the Z axis.
func ClampZMin(s model3d.Solid, max float64) model3d.Solid {
	return ClampAxisMin(s, AxisZ, max)
}
