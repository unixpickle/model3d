package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	// Some conventions:
	// - Distances in millimeters, angles in radians
	// - Length => x-aligned, width => y-aligned, height => z-aligned

	// Flexural modulus of the material we're printing in. Rough estimates
	// should be sufficient.
	// - Nylon: ~1.8 GPa
	// - ABS: ~2 GPa
	// - Markforged Onyx: ~2.9 GPa
	// - PLA: ~4 GPa
	FlexuralModulusGigaPascals = 2.9

	// Length and height of the clip :-)
	Length = 80.0
	Height = 10.0

	// Parameters for a hinge that can be printed in-place without supports!
	//
	// For the joint angle, the natural response is to set this to something
	// small to decrease the footprint of the printed model, but due to the
	// hinge geometry a higher angle actually makes it easier to print the clip
	// without supports. (the inner solid helps support the overhang on the
	// upper portion of the hinge)
	HingeOuterRadius   = 5.0
	HingeInnerRadius   = 2.5
	HingeKnuckleHeight = Height * 0.5
	HingeJointAngle    = -45.0 / 180.0 * math.Pi

	// The flexy bit is the length of material that needs to bend when the
	// latch closes.  To get the same deflection per unit force, we compute the
	// length of the flexy bit as `length multiplier * cbrt(flexural modulus *
	// cross sectional area)`. A little bit hand-wavey but should get OK
	// results.
	FlexyBitThickness        = 2.0
	FlexyBitToStructureGap   = 1.0
	FlexyBitLengthMultiplier = 5.16679

	// The "tab" is the part of the inner solid that clicks into the "clip" on
	// the outer solid. The Z draft angle is symmetrically applied to two sides
	// of it and: makes the overhang printable without supports, helps with
	// alignment when the clip is closed, reduces some forces that might lead
	// to delamination on the clip.
	TabZDraftAngle = 60.0 / 180.0 * math.Pi
	TabDepth       = 2.5
	TabHeight      = Height * 0.8

	// The "clip" sits at the end of the outer solid, at the end of the flexy
	// bit.
	ClipEndWidth     = HingeOuterRadius*2 + 2.0
	ClipEndThickness = TabDepth
	// ^If your layer adhesion's not great (e.g. for ABS) or you just want a
	// stronger part, I'd recommend increasing ClipEndThickness to something like
	// `TabDepth + CloseFitClearance + 1.0`. My subjective opinion is that this
	// is less pretty — it converts the feature for mating with the tab from a
	// through "hole" to a blind one — but this will reduce the shear stresses
	// between layers.

	// Rib parameters. These connect the main structure of the outer solid to
	// the flexy bit.
	RibSpacing   = 13.0
	RibThickness = 2.5

	// The tooth is a triangular feature aligned along the length of the clip.
	// There's a positive side on the outer solid and a negative side on the
	// inner solid; they mate to align the clip and create a better seal.
	ToothSize = 1.5

	// Clearances!
	CloseFitClearance          = 0.2
	LooseFitClearance          = 0.5
	ExtremelyLooseFitClearance = 0.7

	// Some aliases/driven dimensions.
	OuterXLimit = Length - HingeOuterRadius
	InnerXLimit = OuterXLimit - ClipEndThickness - LooseFitClearance
	TabXLimit   = OuterXLimit - ClipEndThickness + TabDepth
)

func main() {
	// We'll be defining two solids: the "outer" one, which has the flexy bit +
	// clip, and the "inner" one, which has a tab that clicks into the clip.
	outer_solid := MakeOuterSolid()
	inner_solid := MakeInnerSolid()

	solid := &model3d.JoinedSolid{
		outer_solid,
		model3d.RotateSolid(inner_solid, model3d.Z(1), HingeJointAngle),
	}

	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.2, 8)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("chip_clip.stl")

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func MakeOuterSolid() model3d.Solid {
	// Compute lengths for the flexy bit.
	flexy_bit_cross_sectional_area := (Height * FlexyBitThickness)
	flexy_bit_length := FlexyBitLengthMultiplier * math.Cbrt(FlexuralModulusGigaPascals*flexy_bit_cross_sectional_area)
	rib_distance_from_x_limit := flexy_bit_length + RibThickness/2.0 + ClipEndThickness

	// Define the outer component as an extruded 2D profile.
	outer_solid := model3d.ProfileSolid(
		&model2d.SubtractedSolid{
			Positive: model2d.JoinedSolid{
				// Hinge. (hole is subtracted later)
				&model2d.Circle{Center: model2d.Coord{X: 0.0, Y: 0.0}, Radius: HingeOuterRadius},
				// Main body.
				model2d.CheckedFuncSolid(
					model2d.Coord{X: 0.0, Y: HingeOuterRadius - ClipEndWidth -
						ClipEndThickness},
					model2d.Coord{X: OuterXLimit, Y: HingeOuterRadius},
					func(coord model2d.Coord) bool {
						rib_x := OuterXLimit - rib_distance_from_x_limit - coord.X
						closest_rib_x := math.Round(rib_x/RibSpacing) * RibSpacing
						clip_start := model2d.Coord{X: OuterXLimit -
							ClipEndThickness/2.0, Y: HingeOuterRadius -
							ClipEndThickness/2.0}

						return coord.X >= 0 && (
						// Outer flexy bit.
						(coord.Y >= HingeOuterRadius-FlexyBitThickness && coord.Y <=
							HingeOuterRadius && coord.X <=
							OuterXLimit-ClipEndThickness/2.0) ||
							// Clip arm.
							(model2d.Segment{
								clip_start,
								clip_start.Add(model2d.Coord{X: 0.0, Y: -ClipEndWidth}),
							}.Dist(coord) <= ClipEndThickness/2.0) ||
							// Ribs.
							(coord.Y >= 0 && coord.Y <= HingeOuterRadius &&
								coord.X >= HingeOuterRadius &&
								coord.X <= OuterXLimit-rib_distance_from_x_limit+RibThickness/2.0 &&
								math.Abs(rib_x-closest_rib_x) <=
									RibThickness/2.0) ||
							// Rigid structure.
							(coord.Y >= 0 &&
								coord.Y <= HingeOuterRadius-FlexyBitThickness-FlexyBitToStructureGap &&
								coord.X <= InnerXLimit))
					},
				),
			},
			// Hole for hinge.
			Negative: &model2d.Circle{
				Center: model2d.Coord{},
				Radius: HingeInnerRadius + ExtremelyLooseFitClearance,
			},
		},
		-Height/2.0,
		Height/2.0,
	)
	outer_solid = &model3d.SubtractedSolid{
		Positive: outer_solid,
		// Clearance for inner component.
		Negative: model3d.CheckedFuncSolid(
			model3d.Coord3D{X: -HingeOuterRadius, Y: -HingeOuterRadius - LooseFitClearance, Z: -Height / 2.0},
			model3d.Coord3D{X: OuterXLimit, Y: 0.0, Z: Height / 2.0},
			func(coord model3d.Coord3D) bool {
				return coord.Y >= -HingeOuterRadius-LooseFitClearance && coord.Y <= 0.0 &&
					// Hinge area.
					((coord.Z >= -HingeKnuckleHeight/2.0-LooseFitClearance &&
						coord.Z <= HingeKnuckleHeight/2.0+LooseFitClearance &&
						coord.X >= -HingeOuterRadius && coord.X <= InnerXLimit) ||
						// Clip tab.
						TabFunc(coord, LooseFitClearance))
			},
		),
	}
	return &model3d.JoinedSolid{
		outer_solid,
		MakeTooth(0.0, HingeOuterRadius+ExtremelyLooseFitClearance, InnerXLimit),
	}
}

func MakeInnerSolid() model3d.Solid {
	inner_solid := &model3d.JoinedSolid{
		// Start with the hinge pin.
		&model3d.Cylinder{P1: model3d.Coord3D{X: 0.0, Y: 0.0, Z: -Height / 2.0},
			P2: model3d.Coord3D{X: 0.0, Y: 0.0, Z: Height / 2.0}, Radius: HingeInnerRadius},
		// Then do everything else!
		model3d.CheckedFuncSolid(
			model3d.Coord3D{X: -HingeOuterRadius, Y: -HingeOuterRadius, Z: -Height / 2.0},
			model3d.Coord3D{X: OuterXLimit, Y: -CloseFitClearance, Z: Height / 2.0},
			func(coord model3d.Coord3D) bool {
				return coord.Y >= -HingeOuterRadius && coord.Y <= -CloseFitClearance && (
				// Connector beam.
				(coord.Z >= -HingeKnuckleHeight/2.0 &&
					coord.Z <= HingeKnuckleHeight/2.0 &&
					coord.X >= 0 &&
					coord.X <= InnerXLimit) ||
					// Main structure.
					(coord.Z >= -Height/2.0 && coord.Z <= Height/2.0 &&
						coord.XY().Norm() >= HingeOuterRadius+ExtremelyLooseFitClearance &&
						coord.X >= 0 && coord.X <= InnerXLimit) ||
					// Clip tab.
					TabFunc(coord, 0.0))
			},
		),
	}
	return &model3d.SubtractedSolid{
		Positive: inner_solid,
		// Clearance for tooth.
		Negative: MakeTooth(CloseFitClearance, HingeOuterRadius,
			InnerXLimit+ExtremelyLooseFitClearance),
	}
}

// Tab + tooth helpers. These are mating features, with an `expand` term
// for adding clearance to the negative side.
func TabFunc(coord model3d.Coord3D, expand float64) bool {
	return (
	// X bounds.
	coord.X >= InnerXLimit && coord.X <= TabXLimit+expand &&
		// Z draft. Computed via overhang angle.
		math.Abs(coord.Z)-expand <=
			TabHeight/2.0-(coord.X-InnerXLimit)/math.Tan(TabZDraftAngle) &&
		// Y draft. This is the clip/tab engagement angle.
		(-coord.Y+expand)/(HingeOuterRadius-CloseFitClearance) >=
			(coord.X-InnerXLimit)/(TabXLimit-InnerXLimit))
}
func MakeTooth(expand float64, x_min float64, x_max float64) model3d.Solid {
	size := ToothSize + expand
	return model3d.CheckedFuncSolid(
		model3d.Coord3D{X: x_min, Y: -size, Z: -size},
		model3d.Coord3D{X: x_max, Y: 0.0, Z: size},
		func(coord model3d.Coord3D) bool {
			return (coord.X >= x_min && coord.X <= x_max && coord.Y <= 0 &&
				math.Abs(coord.Z) <= size+coord.Y)
		},
	)
}
