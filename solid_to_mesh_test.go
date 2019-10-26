package model3d

import (
	"math"
	"testing"
)

func TestSolidToMeshValid(t *testing.T) {
	solid := &SubtractedSolid{
		Positive: PumpkinSolid{Scale: 1},
		Negative: PumpkinSolid{Scale: 0.9},
	}
	mesh := SolidToMesh(solid, 0.1, 0, 0, 0)
	if mesh.NeedsRepair() {
		t.Fatal("mesh needs repair")
	}
}

type PumpkinSolid struct {
	Scale float64
}

func (p PumpkinSolid) Min() Coord3D {
	return Coord3D{X: -p.Scale * 1.6, Y: -p.Scale * 1.6, Z: -p.Scale * 1.6}
}

func (p PumpkinSolid) Max() Coord3D {
	return p.Min().Scale(-1)
}

func (p PumpkinSolid) Contains(c Coord3D) bool {
	g := c.Geo()
	r := p.Scale * (1 + 0.1*math.Abs(math.Sin(g.Lon*4)) + 0.5*math.Cos(g.Lat))
	return c.Norm() <= r
}
