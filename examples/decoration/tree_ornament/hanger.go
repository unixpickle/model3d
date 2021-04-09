package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

const (
	HangerThickness   = 0.2
	HangerMountWidth  = 0.8
	HangerMountHeight = 3.0
	HangerArmLength   = 5.0
	HangerArmWidth    = 0.3
	HangerHookDrop    = 1.0
	HangerHookRadius  = 0.5
	HangerHookWidth   = 0.15

	HangerSpacerLength = 3.0
)

func CreateHangerSolid() model3d.Solid {
	hookProfile := CreateHangerHookProfile()
	return model3d.JoinedSolid{
		// Part that mounts on wall.
		&model3d.Rect{
			MinVal: model3d.XYZ(0, 0, 0),
			MaxVal: model3d.XYZ(HangerMountHeight, HangerThickness, HangerMountWidth),
		},
		// Arm that extends over to tree.
		&model3d.Rect{
			MinVal: model3d.XYZ(HangerMountHeight-HangerThickness, 0, 0),
			MaxVal: model3d.XYZ(HangerMountHeight, HangerArmLength, HangerArmWidth),
		},
		model3d.ProfileSolid(hookProfile, 0, HangerHookWidth),
	}
}

func CreateHangerSpacerSolid() model3d.Solid {
	return model3d.JoinedSolid{
		&model3d.Rect{
			MinVal: model3d.XYZ(0, 0, 0),
			MaxVal: model3d.XYZ(HangerMountWidth, HangerSpacerLength, HangerThickness),
		},
		&model3d.Rect{
			MinVal: model3d.XYZ(0, 0, 0),
			MaxVal: model3d.XYZ(HangerMountWidth, HangerThickness, HangerMountHeight),
		},
		&model3d.Rect{
			MinVal: model3d.XYZ(0, HangerSpacerLength-HangerThickness, 0),
			MaxVal: model3d.XYZ(HangerMountWidth, HangerSpacerLength, HangerMountHeight),
		},
	}
}

func CreateHangerHookProfile() model2d.Solid {
	trace := model2d.NewMesh()
	trace.Add(&model2d.Segment{
		model2d.XY(0, 0),
		model2d.XY(-HangerHookDrop, 0),
	})
	prev := model2d.XY(-HangerHookDrop, 0)
	center := model2d.XY(-HangerHookDrop, -HangerHookRadius)
	for t := math.Pi / 2; t < math.Pi*1.5; t += 0.01 {
		next := center.Add(model2d.XY(math.Cos(t), math.Sin(t)).Scale(HangerHookRadius))
		trace.Add(&model2d.Segment{prev, next})
		prev = next
	}

	offset := model2d.XY(HangerMountHeight-HangerThickness, HangerArmLength-HangerThickness/2)
	trace = trace.Translate(offset)
	return model2d.NewColliderSolidHollow(model2d.MeshToCollider(trace), HangerThickness/2)
}
