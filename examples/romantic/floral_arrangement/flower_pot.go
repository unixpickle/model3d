package main

import (
	"fmt"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	PotRimThickness = 0.2
	PotRimRadius    = 2.5
	PotTextBorder   = 0.2
	PotTextHeight   = 0.5
	PotTopRadius    = 2.4
	PotBottomRadius = 2.2
	PotHeight       = 2.0

	// Used to allow the pot to be hollowed out after printing.
	PotSideThickness    = 0.3
	PotBottomHoleRadius = 0.25
)

func FlowerPot() model3d.Solid {
	letters := WrappedLetters("♡ I LOVE YOU, NUGGIE ♡", PotRimRadius, PotTextHeight, 0.1)
	rim := &model3d.SubtractedSolid{
		Positive: &model3d.Cylinder{
			P1:     model3d.Z(letters.Min().Z - PotTextBorder),
			P2:     model3d.Z(letters.Max().Z + PotTextBorder),
			Radius: PotRimRadius,
		},
		Negative: &model3d.Cylinder{
			P1:     model3d.Z(letters.Min().Z - PotTextBorder - 1e-5),
			P2:     model3d.Z(letters.Max().Z + PotTextBorder + 1e-5),
			Radius: PotRimRadius - PotRimThickness,
		},
	}
	vaseBody := toolbox3d.ClampZMin(
		CurvyCone(
			model3d.Z(rim.Min().Z),
			model3d.Z(rim.Min().Z-PotHeight*PotRimRadius/(PotRimRadius-PotBottomRadius)),
			PotRimRadius,
			PotRimRadius-PotTopRadius,
		),
		rim.Min().Z-PotHeight,
	)
	exterior := model3d.JoinedSolid{
		letters,
		rim,
		vaseBody,
	}
	exteriorMesh := model3d.MarchingCubesSearch(exterior, 0.02, 8)
	interior := model3d.NewColliderSolidInset(
		model3d.MeshToCollider(exteriorMesh),
		PotSideThickness,
	)
	bottomHole := &model3d.Cylinder{
		P1:     model3d.Z(exterior.Min().Z - 1e-5),
		P2:     model3d.Z(exterior.Min().Z + PotSideThickness + 0.03),
		Radius: PotBottomHoleRadius,
	}
	return &model3d.SubtractedSolid{
		Positive: exterior,
		Negative: model3d.JoinedSolid{
			interior,
			bottomHole,
		},
	}
}

func CurvyCone(base, tip model3d.Coord3D, radius, inset float64) model3d.Solid {
	baseCone := &model3d.Cone{
		Base:   base,
		Tip:    tip,
		Radius: radius,
	}
	return model3d.CheckedFuncSolid(
		baseCone.Min(),
		baseCone.Max(),
		func(c model3d.Coord3D) bool {
			thetaOffset := (c.Z - base.Z)
			theta := math.Atan2(c.Y, c.X)
			r := radius - inset*math.Pow(math.Cos(theta*15+5*thetaOffset), 2)
			cone := model3d.Cone{
				Base:   base,
				Tip:    tip,
				Radius: r,
			}
			return cone.Contains(c)
		},
	)
}

func WrappedLetters(text string, radius, textHeight, thickness float64) model3d.Solid {
	wireframe := LetterWireframe(text).Scale(textHeight)
	wireframe = wireframe.MapCoords(func(c model3d.Coord3D) model3d.Coord3D {
		theta := c.X / radius
		return model3d.XYZ(math.Cos(theta)*radius, math.Sin(theta)*radius, c.Z)
	})
	return model3d.NewColliderSolidHollow(model3d.MeshToCollider(wireframe), thickness)
}

func LetterWireframe(text string) *model3d.Mesh {
	letterCurves := map[rune][]model2d.BezierCurve{
		'I': {
			{model2d.XY(0.0, 1.0), model2d.XY(0.4, 1.0)},
			{model2d.XY(0.2, 1.0), model2d.XY(0.2, 0.0)},
			{model2d.XY(0.0, 0.0), model2d.XY(0.4, 0.0)},
		},
		'L': {
			{model2d.XY(0.0, 1.0), model2d.XY(0.0, 0.0)},
			{model2d.XY(0.0, 0.0), model2d.XY(0.6, 0.0)},
		},
		'O': {
			{model2d.XY(0.0, 0.5), model2d.XY(0.0, 0.0), model2d.XY(0.3, 0.0)},
			{model2d.XY(0.3, 0.0), model2d.XY(0.6, 0.0), model2d.XY(0.6, 0.5)},
			{model2d.XY(0.0, 0.5), model2d.XY(0.0, 1.0), model2d.XY(0.3, 1.0)},
			{model2d.XY(0.3, 1.0), model2d.XY(0.6, 1.0), model2d.XY(0.6, 0.5)},
		},
		'V': {
			{model2d.XY(0.0, 1.0), model2d.XY(0.3, 0.0)},
			{model2d.XY(0.6, 1.0), model2d.XY(0.3, 0.0)},
		},
		'E': {
			{model2d.XY(0.0, 0.0), model2d.XY(0.0, 1.0)},
			{model2d.XY(0.0, 0.0), model2d.XY(0.6, 0.0)},
			{model2d.XY(0.0, 0.5), model2d.XY(0.6, 0.5)},
			{model2d.XY(0.0, 1.0), model2d.XY(0.6, 1.0)},
		},
		'Y': {
			{model2d.XY(0.0, 1.0), model2d.XY(0.3, 0.5)},
			{model2d.XY(0.3, 0.5), model2d.XY(0.6, 1.0)},
			{model2d.XY(0.3, 0.5), model2d.XY(0.3, 0.0)},
		},
		'U': {
			{model2d.XY(0.0, 1.0), model2d.XY(0.0, 0.3)},
			{model2d.XY(0.6, 1.0), model2d.XY(0.6, 0.3)},
			{model2d.XY(0.0, 0.3), model2d.XY(0.0, 0.0), model2d.XY(0.3, 0.0)},
			{model2d.XY(0.6, 0.3), model2d.XY(0.6, 0.0), model2d.XY(0.3, 0.0)},
		},
		'N': {
			{model2d.XY(0.0, 0.0), model2d.XY(0.0, 1.0)},
			{model2d.XY(0.6, 0.0), model2d.XY(0.6, 1.0)},
			{model2d.XY(0.0, 1.0), model2d.XY(0.6, 0.0)},
		},
		'G': {
			{model2d.XY(0.6, 0.9), model2d.XY(0.1, 1.0), model2d.XY(0.0, 0.9), model2d.XY(0.0, 0.5)},
			{model2d.XY(0.6, 0.1), model2d.XY(0.1, 0.0), model2d.XY(0.0, 0.1), model2d.XY(0.0, 0.5)},
			{model2d.XY(0.6, 0.1), model2d.XY(0.6, 0.5)},
			{model2d.XY(0.3, 0.5), model2d.XY(0.6, 0.5)},
		},
		',': {
			{model2d.XY(0.1, 0.1), model2d.XY(0.1, 0.0), model2d.XY(0.0, -0.05)},
		},
		'♡': widenCurves(1.25, []model2d.BezierCurve{
			{model2d.XY(0.4, 0.0), model2d.XY(0.1, 0.6)},
			{
				model2d.XY(0.1, 0.6),
				model2d.XY(-0.1, 1.0),
				model2d.XY(0.2, 1.2),
				model2d.XY(0.4, 0.8),
			},
			{model2d.XY(0.8-0.4, 0.0), model2d.XY(0.8-0.1, 0.6)},
			{
				model2d.XY(0.8-0.1, 0.6),
				model2d.XY(0.8+0.1, 1.0),
				model2d.XY(0.8-0.2, 1.2),
				model2d.XY(0.8-0.4, 0.8),
			},
		}),
	}
	mesh := model3d.NewMesh()
	x := 0.0
	for _, ch := range text {
		if ch == ' ' {
			x += 0.8
			continue
		}
		if _, ok := letterCurves[ch]; !ok {
			panic(fmt.Sprintf("unsupported character: '%c'", ch))
		}
		curves := letterCurves[ch]
		maxX := 0.0
		for _, c := range curves {
			model2d.CurveMesh(c, 100).Iterate(func(s *model2d.Segment) {
				p1 := model3d.XZ(s[0].X+x, s[0].Y)
				p2 := model3d.XZ(s[1].X+x, s[1].Y)
				maxX = math.Max(maxX, math.Max(p1.X, p2.X))
				mesh.Add(&model3d.Triangle{p1, p2, p2})
			})
		}
		x = maxX + 0.3
	}
	return mesh
}

func widenCurves(scale float64, curves []model2d.BezierCurve) []model2d.BezierCurve {
	for _, c := range curves {
		for i, p := range c {
			c[i] = p.Mul(model2d.XY(scale, 1.0))
		}
	}
	return curves
}
