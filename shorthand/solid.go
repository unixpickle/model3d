package shorthand

import (
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

type Solid2 = model2d.Solid
type Solid3 = model3d.Solid

func MakeSolid2(min, max C2, f func(c C2) bool) Solid2 {
	return model2d.CheckedFuncSolid(min, max, f)
}

func MakeSolid3(min, max C3, f func(c C3) bool) Solid3 {
	return model3d.CheckedFuncSolid(min, max, f)
}

func Sub2(s1, s2 Solid2) *model2d.SubtractedSolid {
	return model2d.Subtract(s1, s2)
}

func Sub3(s1, s2 Solid3) *model3d.SubtractedSolid {
	return model3d.Subtract(s1, s2)
}

func Join2(s ...Solid2) model2d.JoinedSolid {
	return model2d.JoinedSolid(s)
}

func Join3(s ...Solid3) model3d.JoinedSolid {
	return model3d.JoinedSolid(s)
}

func Stack(s ...Solid3) model3d.StackedSolid {
  return model3d.StackedSolid(s)
}
