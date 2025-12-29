package main

import (
	"flag"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

type Args struct {
	TopBottomSpace     float64 `default:"1" help:"Space at top and bottom of rollers."`
	TopBottomThickness float64 `default:"4"`
	RollerHeight       float64 `default:"20" help:"Height of the bearing cylinders."`
	RollerRadius       float64 `default:"7"`
	RollerInset        float64 `default:"0.5"`
	RollerCutoutInset  float64 `default:"0.5"`
	RollerPinRadius    float64 `default:"4"`
	RollerPinSpace     float64 `default:"0.4"`
	OuterRadius        float64 `default:"30"`
	PinFactor          float64 `default:"0.8" help:"Ratio of pins to maximum allowable"`
	PinPegSlack        float64 `default:"0.2" help:"Extra space around pin negatives"`
	Delta              float64 `default:"0.5" help:"Meshification delta"`
}

func main() {
	var args Args
	toolbox3d.AddFlags(&args, nil)
	flag.Parse()

	pinCount := int(args.PinFactor * 2 * math.Pi * args.OuterRadius / (2 * args.RollerRadius))
	pinCenter := func(i int) model2d.Coord {
		spacing := 2 * math.Pi / float64(pinCount)
		x := math.Cos(spacing * float64(i))
		y := math.Sin(spacing * float64(i))
		return model2d.XY(x, y).Scale(args.OuterRadius)
	}

	totalHeight := args.TopBottomThickness*2 + args.TopBottomSpace*2 + args.RollerHeight

	var rollers model3d.JoinedSolid
	var pins model3d.JoinedSolid
	var pinsNegative model3d.JoinedSolid
	for i := 0; i < pinCount; i++ {
		center := pinCenter(i)
		pins = append(pins, model3d.StackSolids(
			&model3d.Cylinder{
				P1:     model3d.XY(center.X, center.Y),
				P2:     model3d.XYZ(center.X, center.Y, totalHeight-args.TopBottomThickness),
				Radius: args.RollerPinRadius,
			},
			&model3d.ConeSlice{
				P1: model3d.XYZ(center.X, center.Y, totalHeight-args.TopBottomThickness),
				P2: model3d.XYZ(center.X, center.Y, totalHeight),
				R1: args.RollerPinRadius,
				R2: args.RollerPinRadius - args.PinPegSlack,
			},
		))
		pinsNegative = append(pinsNegative, &model3d.Cylinder{
			P1:     model3d.XY(center.X, center.Y),
			P2:     model3d.XYZ(center.X, center.Y, totalHeight),
			Radius: args.RollerPinRadius,
		})

		rollerOuterCyl := &model3d.Cylinder{
			P1:     model3d.XYZ(center.X, center.Y, args.TopBottomThickness+args.TopBottomSpace),
			P2:     model3d.XYZ(center.X, center.Y, args.TopBottomThickness+args.TopBottomSpace+args.RollerHeight),
			Radius: args.RollerRadius,
		}
		rollerOuter := model3d.CheckedFuncSolid(
			rollerOuterCyl.Min(),
			rollerOuterCyl.Max(),
			func(c model3d.Coord3D) bool {
				midZ := (rollerOuterCyl.P1.Z + rollerOuterCyl.P2.Z) / 2
				halfHeight := math.Abs(rollerOuterCyl.P1.Z-rollerOuterCyl.P2.Z) / 2
				inset := args.RollerInset * (1 - math.Abs(c.Z-midZ)/halfHeight)
				return c.XY().Dist(rollerOuterCyl.P1.XY()) < rollerOuterCyl.Radius-inset
			},
		)

		rollerCutoutCyl := &model3d.Cylinder{
			P1:     model3d.XYZ(center.X, center.Y, args.TopBottomThickness+args.TopBottomSpace),
			P2:     model3d.XYZ(center.X, center.Y, args.TopBottomThickness+args.TopBottomSpace+args.RollerHeight),
			Radius: args.RollerPinRadius + args.RollerPinSpace,
		}
		rollerCutout := model3d.CheckedFuncSolid(
			rollerCutoutCyl.Min().AddScalar(-args.RollerCutoutInset),
			rollerCutoutCyl.Max().AddScalar(args.RollerCutoutInset),
			func(c model3d.Coord3D) bool {
				midPoint := (rollerCutoutCyl.P2.Z + rollerCutoutCyl.P1.Z) / 2
				fracFromMiddle := math.Abs(c.Z-midPoint) / (midPoint - rollerCutoutCyl.P1.Z)
				r := rollerCutoutCyl.Radius + args.RollerCutoutInset*(1-fracFromMiddle)
				return c.XY().Dist(rollerCutoutCyl.P1.XY()) <= r
			},
		)
		rollers = append(rollers, model3d.Subtract(
			rollerOuter,
			rollerCutout,
		))
	}

	topOrBottom := model3d.Subtract(
		&model3d.Cylinder{
			P1:     model3d.Z(0),
			P2:     model3d.Z(args.TopBottomThickness),
			Radius: args.OuterRadius + args.RollerRadius,
		},
		&model3d.Cylinder{
			P1:     model3d.Z(0),
			P2:     model3d.Z(args.TopBottomThickness),
			Radius: args.OuterRadius - args.RollerRadius,
		},
	)

	mesh := model3d.DualContour(rollers, args.Delta, true, false)
	mesh = mesh.EliminateCoplanar(1e-5)
	mesh.SaveGroupedSTL("rollers.stl")

	mesh = model3d.DualContour(model3d.JoinedSolid{topOrBottom, pins}, args.Delta, true, false)
	mesh = mesh.EliminateCoplanar(1e-5)
	mesh.SaveGroupedSTL("bottom.stl")

	mesh = model3d.DualContour(model3d.Subtract(topOrBottom, pins), args.Delta, true, false)
	mesh = mesh.EliminateCoplanar(1e-5)
	mesh.SaveGroupedSTL("top.stl")
}
