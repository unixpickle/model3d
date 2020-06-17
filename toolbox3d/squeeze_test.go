package toolbox3d

import (
	"fmt"
	"math"
	"testing"

	"github.com/unixpickle/model3d/model3d"
)

func TestAxisSqueeze(t *testing.T) {
	for _, axis := range []Axis{AxisX, AxisY, AxisZ} {
		t.Run(fmt.Sprintf("Axis%d", axis), func(t *testing.T) {
			t.Run("Smaller", func(t *testing.T) {
				testTransform(t, &AxisSqueeze{
					Min:   -0.1,
					Max:   0.9,
					Axis:  axis,
					Ratio: 0.8,
				})
			})
			t.Run("Larger", func(t *testing.T) {
				testTransform(t, &AxisSqueeze{
					Min:   -0.1,
					Max:   0.9,
					Axis:  axis,
					Ratio: 1.2,
				})
			})
		})
	}
}

func testTransform(t *testing.T, transform model3d.Transform) {
	solid := testingSolid()
	mesh := model3d.MarchingCubesSearch(solid, 0.02, 8)
	sdf := model3d.MeshToSDF(mesh)

	transformed := model3d.TransformSolid(transform, solid)

	solid1 := model3d.TransformSolid(transform.Inverse(), transformed)
	mesh1 := model3d.MarchingCubesSearch(solid1, 0.02, 8)
	sdf1 := model3d.MeshToSDF(mesh1)

	mesh2 := model3d.MarchingCubesSearch(transformed, 0.02, 8)
	mesh2 = mesh2.MapCoords(transform.Inverse().Apply)
	sdf2 := model3d.MeshToSDF(mesh2)
	solid2 := model3d.NewColliderSolid(model3d.MeshToCollider(mesh2))

	min, max := solid.Min(), solid.Max()
	min = min.Sub(model3d.Coord3D{X: 1, Y: 1, Z: 1}.Scale(0.1))
	max = max.Add(model3d.Coord3D{X: 1, Y: 1, Z: 1}.Scale(0.1))
	for i := 0; i < 1000; i++ {
		c := model3d.NewCoord3DRandUniform().Mul(max.Sub(min)).Add(min)
		dist := sdf.SDF(c)
		dist1 := sdf1.SDF(c)
		dist2 := sdf2.SDF(c)
		if math.Abs(dist-dist1) > 0.04 {
			t.Errorf("bad SDF at %v", c)
		}
		if math.Abs(dist-dist2) > 0.04 {
			t.Errorf("bad SDF at %v", c)
		}
		if math.Abs(dist) > 0.04 {
			contained := solid.Contains(c)
			contained1 := solid1.Contains(c)
			contained2 := solid2.Contains(c)
			if contained != contained1 {
				t.Errorf("disagreement on solid transform at %v", c)
			}
			if contained != contained2 {
				t.Errorf("disagreement on mesh transform at %v", c)
			}
		}
	}
}

func testingSolid() model3d.Solid {
	return model3d.JoinedSolid{
		&model3d.Cylinder{
			P1:     model3d.XY(0.2, 0.3),
			P2:     model3d.XZ(0.3, 0.5),
			Radius: 0.1,
		},
		&model3d.Cylinder{
			P1:     model3d.X(0.2),
			P2:     model3d.XZ(0.3, 0.5),
			Radius: 0.1,
		},
		&model3d.Sphere{Center: model3d.XZ(0.25, 0.25), Radius: 0.2},
	}
}
