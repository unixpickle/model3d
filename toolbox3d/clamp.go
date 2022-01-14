package toolbox3d

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
)

// ClampAxis creates a solid which does not extend beyond
// the range [min, max] along the given axis.
func ClampAxis(s model3d.Solid, axis int, min, max float64) model3d.Solid {
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
