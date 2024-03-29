package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
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
	BinDepth        = 4.0
	BinCornerRadius = 0.5
	BinThickness    = 0.2

	Epsilon = 0.01
)

func main() {
	mount := WallMountSolid()
	ax := &toolbox3d.AxisSqueeze{
		Axis:  toolbox3d.AxisY,
		Min:   GrooveSize + WallMountSideThickness + 0.1,
		Max:   WallMountHeight - 0.1,
		Ratio: 0.1,
	}
	log.Println("Creating mount mesh...")
	mountMesh := model3d.MarchingCubesConj(mount, Epsilon, 8, ax)
	log.Println("Simplifying mount mesh...")
	mountMesh = SimplifyMesh(mountMesh)
	log.Println("Saving mount mesh...")
	mountMesh.SaveGroupedSTL("mount.stl")

	bin := BinSolid()
	ss := BinSqueeze()
	log.Println("Creating bin mesh...")
	binMesh := model3d.MarchingCubesConj(bin, Epsilon, 8, ss.Transform(bin))
	log.Println("Simplifying bin mesh...")
	binMesh = SimplifyMesh(binMesh)
	log.Println("Saving bin mesh...")
	binMesh.SaveGroupedSTL("bin.stl")

	log.Println("Rendering mount...")
	render3d.SaveRandomGrid("rendering_mount.png", mountMesh, 3, 3, 300, nil)

	log.Println("Rendering bin...")
	render3d.SaveRandomGrid("rendering_bin.png", binMesh, 3, 3, 300, nil)
}

func SimplifyMesh(m *model3d.Mesh) *model3d.Mesh {
	oldCount := len(m.TriangleSlice())
	m = m.EliminateCoplanar(1e-5)
	log.Printf(" => went from %d to %d triangles", oldCount, len(m.TriangleSlice()))
	return m
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
	negative := model3d.RotateSolid(
		model3d.RotateSolid(WallMountNegative(), model3d.Y(1), math.Pi),
		model3d.X(1),
		math.Pi/2,
	)
	negMin, negMax := negative.Min(), negative.Max()
	minX := negMin.X + BinThickness/2
	maxX := negMax.X - BinThickness/2
	minY := negMax.Y + BinThickness/2 - GrooveEdge
	maxY := minY + BinDepth
	minZ := negMin.Z
	maxZ := minZ + BinHeight - BinThickness/2

	// The 2D shape of the center of the basket surface.
	basketPath := model2d.JoinedCurve{
		model2d.BezierCurve{
			model2d.XY(minX, minY),
			model2d.XY(maxX, minY),
		},
		model2d.BezierCurve{
			model2d.XY(maxX, minY),
			model2d.XY(maxX, maxY-BinCornerRadius),
		},
		model2d.BezierCurve{
			model2d.XY(maxX, maxY-BinCornerRadius),
			model2d.XY(maxX, maxY),
			model2d.XY(maxX-BinCornerRadius, maxY),
		},
		model2d.BezierCurve{
			model2d.XY(maxX-BinCornerRadius, maxY),
			model2d.XY(minX+BinCornerRadius, maxY),
		},
		model2d.BezierCurve{
			model2d.XY(minX+BinCornerRadius, maxY),
			model2d.XY(minX, maxY),
			model2d.XY(minX, maxY-BinCornerRadius),
		},
		model2d.BezierCurve{
			model2d.XY(minX, maxY-BinCornerRadius),
			model2d.XY(minX, minY),
		},
	}
	basketMesh2d := model2d.CurveMesh(basketPath, 1000)
	collider2d := model2d.MeshToCollider(basketMesh2d)
	baseSolid2d := model2d.NewColliderSolid(collider2d)
	basketSolid2d := model2d.NewColliderSolidHollow(
		collider2d,
		BinThickness/2,
	)
	baseSolid3d := model3d.ProfileSolid(baseSolid2d, minZ, minZ+BinThickness)
	basketSolid3d := model3d.ProfileSolid(basketSolid2d, minZ, maxZ)

	var rimSegments []model3d.Segment
	basketMesh2d.Iterate(func(s *model2d.Segment) {
		rimSegments = append(rimSegments, model3d.NewSegment(
			model3d.XYZ(s[0].X, s[0].Y, maxZ),
			model3d.XYZ(s[1].X, s[1].Y, maxZ),
		))
	})
	basketRim := toolbox3d.LineJoin(BinThickness/2, rimSegments...)

	return model3d.JoinedSolid{negative, baseSolid3d, basketSolid3d, basketRim}
}

func BinSqueeze() *toolbox3d.SmartSqueeze {
	bin := BinSolid()

	ss := &toolbox3d.SmartSqueeze{
		Axis:         toolbox3d.AxisZ,
		SqueezeRatio: 0.1,
		PinchRange:   0.03,
		PinchPower:   0.25,
	}

	// Don't lower resolution of bottom of negative where
	// a groove is, and don't mess with the Z of the bottom
	// of the bin.
	ss.AddUnsqueezable(bin.Min().Z, bin.Min().Z+math.Max(BinThickness, GrooveSize+GrooveSlack))

	// Flatten out top of the mount negative.
	neg := WallMountNegative()
	ss.AddPinch(bin.Min().Z + neg.Max().Y - neg.Min().Y)

	// Don't lower resolution of rounded rim.
	ss.AddUnsqueezable(bin.Max().Z-BinThickness, bin.Max().Z+1e-5)

	return ss
}
