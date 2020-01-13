package main

const (
	ScrewRadius     = 0.3
	ScrewGrooveSize = 0.06
	ScrewSlack      = 0.04

	PropellerHubRadius = 0.7

	BladeRadius    = 3.5
	BladeThickness = 0.1
	BladeDepth     = 1.0
	BladeCount     = 8

	SpineThickness = 0.5
	SpineWidth     = 1.2
	SpineLength    = 8.0

	HoleRadius = 0.36
	PoleRadius = 0.33

	GearModule     = 0.1
	GearAddendum   = 0.08
	GearDedendum   = 0.08
	SmallGearTeeth = 20
	LargeGearTeeth = 40
	GearAirGap     = 0.02
	GearDistance   = GearAirGap + GearModule*(SmallGearTeeth+LargeGearTeeth)/2
)
