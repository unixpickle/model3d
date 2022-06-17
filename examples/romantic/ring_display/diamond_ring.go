package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func DiamondRing() (model3d.Solid, toolbox3d.CoordColorFunc) {
	diamondMesh := CreateDiamondPolytope().Mesh()
	diamondMesh = diamondMesh.Scale(0.08 / (diamondMesh.Max().X - diamondMesh.Min().X))
	diamondMesh = diamondMesh.Rotate(model3d.Y(1), -math.Pi/2)
	diamondMesh = diamondMesh.Translate(diamondMesh.Min().Mid(diamondMesh.Max()).Scale(-1))
	diamond := model3d.NewColliderSolid(model3d.MeshToCollider(diamondMesh))

	rounding := 0.01
	thickness := 0.08 - rounding
	radius := 0.4
	maxDepth := 0.2 - rounding
	minDepth := 0.1 - rounding
	depthFn := model2d.BezierCurve{
		model2d.XY(0.0, maxDepth),
		model2d.XY(0.23*math.Pi, maxDepth),
		model2d.XY(0.33*math.Pi, minDepth),
		model2d.XY(math.Pi, minDepth),
	}
	roughBody := model3d.CheckedFuncSolid(
		model3d.XYZ(-(radius+thickness), -(radius+thickness), 0),
		model3d.XYZ(radius+thickness, radius+thickness, maxDepth*2),
		func(c model3d.Coord3D) bool {
			if math.Abs(c.XY().Norm()-(radius+thickness/2)) > thickness/2 {
				return false
			}
			theta := math.Atan2(c.Y, c.X)
			if theta < 0 {
				theta = -theta
			}
			depth := depthFn.EvalX(theta)
			return math.Abs(c.Z-maxDepth/2) < depth/2
		},
	)
	roughSDF := model3d.DualContourSDF(roughBody, 0.01)
	smoothBody := model3d.SDFToSolid(roughSDF, rounding)

	var diamonds model3d.JoinedSolid
	addDiamond := func(theta, z float64) {
		dz := model3d.TranslateSolid(diamond, model3d.XZ(radius+thickness+rounding, z))
		diamonds = append(diamonds, model3d.RotateSolid(dz, model3d.Z(1), theta))
	}
	for _, z := range []float64{0.0375, maxDepth / 2, maxDepth - 0.0375} {
		for i := -3; i <= 3; i++ {
			theta := float64(i) * 0.45 / 3.0
			addDiamond(theta, z)
		}
	}
	addDiamond(-4.0*0.45/3.0, maxDepth/2-0.035)
	addDiamond(-4.0*0.45/3.0, maxDepth/2+0.035)
	// addDiamond(-5.0*0.45/3.0, maxDepth/2)
	addDiamond(4.0*0.45/3.0, maxDepth/2-0.035)
	addDiamond(4.0*0.45/3.0, maxDepth/2+0.035)
	// addDiamond(5.0*0.45/3.0, maxDepth/2)

	colorFn := func(c model3d.Coord3D) render3d.Color {
		if roughBody.Contains(c) || roughSDF.SDF(c) > -(rounding+0.01) {
			return render3d.NewColor(0.8)
		} else {
			return render3d.NewColorRGB(0.6, 0.8, 1.0)
		}
	}

	return model3d.JoinedSolid{diamonds.Optimize(), smoothBody}, colorFn
}

// CreateDiamondPolytope copied from examples/decoration/diamond.go.
func CreateDiamondPolytope() model3d.ConvexPolytope {
	const (
		NumSides   = 12
		BaseHeight = 0.4
		TipHeight  = 1.2
	)

	system := model3d.ConvexPolytope{
		&model3d.LinearConstraint{
			Normal: model3d.Coord3D{Z: -1},
			Max:    BaseHeight,
		},
	}

	addTriangle := func(t *model3d.Triangle) {
		n := t.Normal()

		// Make sure the normal points outward.
		if n.Dot(t[0]) < 0 {
			t[0], t[1] = t[1], t[0]
		}

		system = append(system, &model3d.LinearConstraint{
			Normal: t.Normal(),
			Max:    t[0].Dot(t.Normal()),
		})
	}

	iAngle := math.Pi * 2 / NumSides
	rimPoint := func(i int) model3d.Coord3D {
		return model3d.Coord3D{
			X: math.Cos(float64(i) * iAngle),
			Y: math.Sin(float64(i) * iAngle),
		}
	}
	basePoint := func(i int) model3d.Coord3D {
		return model3d.Coord3D{
			X: math.Cos((float64(i) + 0.5) * iAngle),
			Y: math.Sin((float64(i) + 0.5) * iAngle),
		}.Scale(1 - BaseHeight).Sub(model3d.Z(BaseHeight))
	}
	tipPoint := model3d.Z(TipHeight)

	for i := 0; i < NumSides; i++ {
		addTriangle(&model3d.Triangle{
			rimPoint(i),
			rimPoint(i + 1),
			tipPoint,
		})
		addTriangle(&model3d.Triangle{
			rimPoint(i),
			rimPoint(i + 1),
			basePoint(i),
		})
		addTriangle(&model3d.Triangle{
			basePoint(i),
			basePoint(i + 1),
			rimPoint(i + 1),
		})
	}
	return system
}
