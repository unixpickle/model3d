package model3d

import (
	"math/rand"
	"testing"
)

func TestMarchingCubesDeterminism(t *testing.T) {
	table1 := mcLookupTable()
	for i := 0; i < 10; i++ {
		table2 := mcLookupTable()
		for key, value1 := range table1 {
			value2 := table2[key]
			if len(value1) != len(value2) {
				t.Fatal("mismatched triangle count")
			}
			for j, tri := range value1 {
				if value2[j] != tri {
					t.Fatal("mismatched triangle:", value1, value2)
				}
			}
		}
	}
}

func TestMarchingCubesRandom(t *testing.T) {
	for _, name := range []string{"Normal", "Search"} {
		t.Run(name, func(t *testing.T) {
			for i := 0; i < 30; i++ {
				var mesh *Mesh
				if name == "Normal" {
					mesh = MarchingCubes(randomSolid{}, 0.1)
				} else {
					mesh = MarchingCubesSearch(randomSolid{}, 0.1, 2)
				}
				MustValidateMesh(t, mesh, true)
			}
		})
	}
}

func BenchmarkMarchingCubes(b *testing.B) {
	solid := &CylinderSolid{
		P1:     XYZ(1, 2, 3),
		P2:     XYZ(3, 1, 4),
		Radius: 0.5,
	}
	for i := 0; i < b.N; i++ {
		MarchingCubes(solid, 0.025)
	}
}

type randomSolid struct{}

func (r randomSolid) Min() Coord3D {
	return Coord3D{}
}

func (r randomSolid) Max() Coord3D {
	return XYZ(1, 1, 1)
}

func (r randomSolid) Contains(c Coord3D) bool {
	return InBounds(r, c) && rand.Intn(4) == 0
}
