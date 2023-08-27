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
)

func CupSolid() (model3d.Solid, toolbox3d.CoordColorFunc) {
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
			} else if c.Z > CupHeight-CupContentsDepth {
				r := c.XY().Norm()
				return r < radius && r > radius-CupRimThickness
			}
			return c.XY().Norm() < radius
		},
	)
	colorFn := func(c model3d.Coord3D) render3d.Color {
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
