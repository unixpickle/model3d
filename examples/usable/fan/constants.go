package main

import "math"

const (
	ScrewRadius     = 0.3
	ScrewGrooveSize = 0.06
	ScrewSlack      = 0.05

	PropellerHubRadius = 0.7

	BladeRadius    = 3.5
	BladeThickness = 0.15
	BladeDepth     = 1.5
	BladeCount     = 5

	SpineThickness    = 0.4
	SpineWasherSize   = 0.2
	SpineWasherRadius = HoleRadius + 0.1
	SpineWidth        = 1.2
	SpineLength       = 8.0

	HoleRadius      = 0.36
	PoleRadius      = 0.33
	PoleExtraLength = 0.04

	GearThickness     = 0.4
	GearModule        = 0.1
	GearAddendum      = 0.08
	GearDedendum      = 0.08
	GearPressureAngle = 25 * math.Pi / 180
	GearHelicalAngle  = 20 * math.Pi / 180
	SmallGearTeeth    = 20
	LargeGearTeeth    = 40
	GearAirGap        = 0.04
	LargeGearRadius   = GearModule * LargeGearTeeth / 2
	GearDistance      = GearAirGap + GearModule*(SmallGearTeeth+LargeGearTeeth)/2

	CrankGearSections     = 8
	CrankGearRimSize      = 0.4
	CrankGearCenterRadius = 0.5
	CrankGearPoleSize     = 0.3
	CrankHandleRadius     = 0.2
	CrankHandleLength     = 2.0

	CrankBoltRadius    = 0.5
	CrankBoltThickness = 0.09
)
