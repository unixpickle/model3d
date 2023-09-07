package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	CupFacets         = 16
	CupFacetTopHeight = 0.2
	CupFacetFlatten   = 0.05
	CupBottomRadius   = 0.6
	CupTopRadius      = 0.9
	CupHeight         = 3
	CupRimHeight      = 0.4
	CupRimThickness   = 0.1
	CupContentsDepth  = 0.05
	CupEngravingDepth = 0.03
)

func CupSolid() (model3d.Solid, toolbox3d.CoordColorFunc) {
	engraving := ReadEngraving()
	solid := model3d.CheckedFuncSolid(
		model3d.XYZ(-CupTopRadius, -CupTopRadius, 0),
		model3d.XYZ(CupTopRadius, CupTopRadius, CupHeight),
		func(c model3d.Coord3D) bool {
			fracTop := c.Z / CupHeight
			radius := fracTop*CupTopRadius + (1-fracTop)*CupBottomRadius
			if c.Z < CupHeight-CupRimHeight {
				theta := math.Atan2(c.X, c.Y)
				if theta < 0 {
					theta += math.Pi * 2
				}
				facetX := math.Abs(math.Mod(CupFacets*theta/(math.Pi*2), 1)-0.5) * 2
				facetZ := 1 - math.Min(CupFacetTopHeight, (CupHeight-CupRimHeight)-c.Z)/
					CupFacetTopHeight
				r := model2d.XY(facetX, facetZ).Norm()
				if r < 1 {
					facetDepth := math.Sqrt(1 - r*r)
					radius -= facetDepth * CupFacetFlatten
				}
			}

			if c.Z > CupHeight-CupRimHeight {
				engravingX := math.Atan2(c.X, c.Y)
				engravingY := (CupHeight-c.Z)/CupRimHeight - 0.5
				if engraving.Contains(model2d.XY(engravingX, engravingY)) {
					radius -= CupEngravingDepth
				}
			}

			r := c.XY().Norm()
			if c.Z > CupHeight-CupContentsDepth {
				if r < radius-CupRimThickness {
					return false
				}
			}

			return r < radius
		},
	)
	colorFn := func(c model3d.Coord3D) render3d.Color {
		if c.Z > CupHeight-CupRimHeight {
			engravingX := math.Atan2(c.X, c.Y)
			engravingY := (CupHeight-c.Z)/CupRimHeight - 0.5
			if engraving.Contains(model2d.XY(engravingX, engravingY)) {
				return render3d.NewColorRGB(0.5*0x65/255.0, 0.5*0xbc/255.0, 0.5*0xd4/255.0)
			}
		}

		fracTop := c.Z / CupHeight
		radius := fracTop*CupTopRadius + (1-fracTop)*CupBottomRadius
		r := c.XY().Norm()
		if r <= radius-CupRimThickness && c.Z > CupHeight-CupRimHeight {
			return render3d.NewColor(1.0)
		} else {
			return render3d.NewColorRGB(0x65/255.0, 0xbc/255.0, 0xd4/255.0)
		}
	}
	return solid, colorFn
}

func ReadEngraving() model2d.Solid {
	mesh := model2d.MustReadBitmap("text.png", nil).FlipX().Mesh().SmoothSq(10)
	mesh = mesh.Translate(mesh.Min().Mid(mesh.Max()).Scale(-1))
	mesh = mesh.Scale(3.0 / (mesh.Max().X - mesh.Min().X))
	mesh = mesh.MapCoords(model2d.XY(0.8, 1.0).Mul) // make it look less stretched
	return model2d.NewColliderSolidInset(model2d.MeshToCollider(mesh), -0.02)
}
