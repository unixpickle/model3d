package shorthand

import (
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

type Mesh2 = model2d.Mesh
type Mesh3 = model3d.Mesh

func MeshSolid2(mesh *Mesh2, inset float64) Solid2 {
	return model2d.NewColliderSolidInset(model2d.MeshToCollider(mesh), inset)
}

func MeshSolid3(mesh *Mesh3, inset float64) Solid3 {
	return model3d.NewColliderSolidInset(model3d.MeshToCollider(mesh), inset)
}
