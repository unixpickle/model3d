package toolbox3d

import (
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

// SliceSolid returns a 2D cross-section of a 3D solid
// along the given axis at the given offset.
//
// For example, if axis is AxisY, and axisValue is 3, then
// the resulting solid is true at (x', y') if the 3D solid
// is true at (x', 3, y').
func SliceSolid(solid model3d.Solid, axis Axis, axisValue float64) model2d.Solid {
	var xIndex, yIndex int
	switch axis {
	case AxisX:
		xIndex, yIndex = 1, 2
	case AxisY:
		xIndex, yIndex = 0, 2
	case AxisZ:
		xIndex, yIndex = 0, 1
	}

	to2d := func(c3d model3d.Coord3D) model2d.Coord {
		arr := c3d.Array()
		return model2d.XY(arr[xIndex], arr[yIndex])
	}
	to3d := func(c2d model2d.Coord) model3d.Coord3D {
		var res [3]float64
		res[axis] = axisValue
		res[xIndex] = c2d.X
		res[yIndex] = c2d.Y
		return model3d.NewCoord3DArray(res)
	}

	min, max := solid.Min(), solid.Max()
	return model2d.CheckedFuncSolid(to2d(min), to2d(max), func(c model2d.Coord) bool {
		return solid.Contains(to3d(c))
	})
}
