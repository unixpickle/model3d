package toolbox3d

import (
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

// SliceSolid returns a 2D cross-section of a 3D solid
// along the given axis at the given offset.
//
// See model3d.CrossSectionSolid for a more general API.
//
// For example, if axis is AxisY, and axisValue is 3, then
// the resulting solid is true at (x', y') if the 3D solid
// is true at (x', 3, y').
func SliceSolid(solid model3d.Solid, axis Axis, axisValue float64) model2d.Solid {
	return model3d.CrossSectionSolid(solid, int(axis), axisValue)
}
