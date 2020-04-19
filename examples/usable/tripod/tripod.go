package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func CreateTripod() model3d.Solid {
	return model3d.IntersectedSolid{
		model3d.JoinedSolid{
			createLeg(0),
			createLeg(2 * math.Pi / 3.0),
			createLeg(4 * math.Pi / 3.0),

			model3d.StackedSolid{
				&model3d.Cylinder{
					P1:     model3d.Coord3D{Z: TripodHeadZ},
					P2:     model3d.Coord3D{Z: TripodHeadZ + TripodHeadHeight},
					Radius: TripodHeadRadius,
				},
				&toolbox3d.ScrewSolid{
					P2:         model3d.Coord3D{Z: CradleBottomThickness - ScrewSlack},
					Radius:     ScrewRadius,
					GrooveSize: ScrewGroove,
				},
			},
		},

		// Cut off bottom of leg cylinders.
		&model3d.Rect{
			MinVal: model3d.Coord3D{X: math.Inf(-1), Y: math.Inf(-1)},
			MaxVal: model3d.Coord3D{X: math.Inf(1), Y: math.Inf(1), Z: TripodHeight * 2},
		},
	}
}

func createLeg(theta float64) model3d.Solid {
	legEnd := model3d.Coord3D{X: TripodLegSpanRadius * math.Cos(theta), Y: TripodLegSpanRadius * math.Sin(theta)}
	footEnd := legEnd.Scale((TripodFootOutset + TripodLegSpanRadius) / TripodLegSpanRadius)
	return &model3d.SubtractedSolid{
		Positive: model3d.JoinedSolid{
			&model3d.Cylinder{
				P1:     model3d.Coord3D{Z: TripodHeight},
				P2:     legEnd,
				Radius: TripodLegRadius,
			},
			&model3d.Cylinder{
				P1:     footEnd,
				P2:     footEnd.Add(model3d.Coord3D{Z: TripodFootHeight}),
				Radius: TripodFootRadius,
			},
		},
		Negative: &toolbox3d.ScrewSolid{
			P1:         footEnd.Add(model3d.Coord3D{Z: -1e-5}),
			P2:         footEnd.Add(model3d.Coord3D{Z: 1000}),
			Radius:     ScrewRadius + ScrewSlack,
			GrooveSize: ScrewGroove,
		},
	}
}
