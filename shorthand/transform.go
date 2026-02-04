package shorthand

import (
	"fmt"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

type Transformer3 interface {
}

func Translate3[T Transformer3](x T, offset C3) T {
	return Transform3(&model3d.Translate{Offset: offset}, x)
}

func Rotate3[T Transformer3](x T, axis C3, angle float64) T {
	return Transform3(model3d.Rotation(axis, angle), x)
}

func Scale3[T Transformer3](x T, scale float64) T {
	return Transform3(&model3d.Scale{Scale: scale}, x)
}

func Transform3[T Transformer3](t model3d.DistTransform, x T) T {
	var y any = x
	var out any
	switch x := y.(type) {
	case model3d.Solid:
		out = model3d.TransformSolid(t, x)
	case *model3d.Mesh:
		out = x.Transform(t)
	case model3d.Collider:
		out = model3d.TransformCollider(t, x)
	case model3d.SDF:
		out = model3d.TransformSDF(t, x)
  case toolbox3d.CoordColorFunc:
    out = x.Transform(t)
	default:
		panic(fmt.Sprintf("unknown type passed to Transform3: %T", x))
	}
	return out.(T)
}

type Transformer2 interface {
}

func Translate2[T Transformer2](x T, offset C2) T {
	return Transform2(&model2d.Translate{Offset: offset}, x)
}

func Rotate2[T Transformer2](x T, angle float64) T {
	return Transform2(model2d.Rotation(angle), x)
}

func Scale2[T Transformer2](x T, scale float64) T {
	return Transform2(&model2d.Scale{Scale: scale}, x)
}

func Transform2[T Transformer2](t model2d.DistTransform, x T) T {
	var y any = x
	var out any
	switch x := y.(type) {
	case model2d.Solid:
		out = model2d.TransformSolid(t, x)
	case *model2d.Mesh:
		out = x.Transform(t)
	case model2d.Collider:
		out = model2d.TransformCollider(t, x)
	case model2d.SDF:
		out = model2d.TransformSDF(t, x)
	default:
		panic(fmt.Sprintf("unknown type passed to Transform2: %T", x))
	}
	return out.(T)
}
