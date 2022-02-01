package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	GrooveSize  = 0.12
	GrooveEdge  = 0.2
	GrooveSlack = 0.02

	WallMountBackThickness = 0.15
	WallMountSideThickness = 0.2
	WallMountWidth         = 4.0
	WallMountHeight        = 4.0

	BinHeight       = 5.0
	BinCornerRadius = 0.5
	BinThickness    = 0.2
)

func main() {
	mount := WallMountSolid()
	ax := &toolbox3d.AxisSqueeze{
		Axis:  toolbox3d.AxisY,
		Min:   GrooveSize + WallMountSideThickness + 0.1,
		Max:   WallMountHeight - 0.1,
		Ratio: 0.1,
	}
	mountMesh := model3d.MarchingCubesConj(mount, 0.02, 8, ax)
	mountMesh.SaveGroupedSTL("mount.stl")

	bin := BinSolid()
	binMesh := model3d.MarchingCubesSearch(bin, 0.02, 8)
	binMesh.SaveGroupedSTL("bin.stl")
}

func WallMountSolid() model3d.Solid {
	return model3d.CheckedFuncSolid(
		model3d.XYZ(0, 0, 0),
		model3d.XYZ(
			WallMountWidth,
			WallMountHeight,
			WallMountBackThickness+GrooveEdge+GrooveSize,
		),
		func(c model3d.Coord3D) bool {
			// Back of the holder.
			if c.Z < WallMountBackThickness {
				return true
			}
			// Sides and bottom (with groove).
			grooveIndent := math.Max(0, c.Z-(WallMountBackThickness+GrooveEdge))
			return c.X-WallMountSideThickness < grooveIndent ||
				WallMountWidth-WallMountSideThickness-c.X < grooveIndent ||
				c.Y-WallMountSideThickness < grooveIndent
		},
	)
}

func WallMountNegative() model3d.Solid {
	inset := WallMountSideThickness + GrooveSlack
	slopeStart := WallMountBackThickness + GrooveEdge - GrooveSlack
	maxZ := slopeStart + GrooveSlack*2 + GrooveSize*2
	return model3d.CheckedFuncSolid(
		model3d.XYZ(inset, inset, WallMountBackThickness+GrooveSlack),
		model3d.XYZ(WallMountWidth-inset, WallMountHeight, maxZ+GrooveEdge),
		func(c model3d.Coord3D) bool {
			edgeInset := 0.0
			if c.Z < slopeStart {
			} else if c.Z < slopeStart+GrooveSize+GrooveSlack {
				edgeInset = c.Z - slopeStart
			} else {
				edgeInset = math.Max(0, maxZ-c.Z)
			}
			return c.X-inset >= edgeInset && WallMountWidth-inset-c.X >= edgeInset &&
				c.Y-inset >= edgeInset
		},
	)
}

func BinSolid() model3d.Solid {
	negative := model3d.RotateSolid(WallMountNegative(), model3d.X(1), math.Pi/2)
	// TODO: create bin itself
	return negative
}
