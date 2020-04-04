package model3d

import (
	"fmt"
	"testing"
)

func TestMarchingCubesDeterminism(t *testing.T) {
	table1 := mcLookupTable()
	fmt.Println(allMcRotations())
	fmt.Println(allMcRotations())
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
	for i := 0; i < 30; i++ {
		mesh := MarchingCubes(randomSolid{}, 0.1)
		if mesh.NeedsRepair() {
			t.Error("mesh needs repair")
		}
		if numSing := len(mesh.SingularVertices()); numSing != 0 {
			t.Error("mesh has", numSing, "singular vertices")
		}
		if numInts := mesh.SelfIntersections(); numInts != 0 {
			t.Error("mesh has", numInts, "self-intersections")
		}
		if _, numNorms := mesh.RepairNormals(1e-8); numNorms != 0 {
			t.Error("mesh has", numNorms, "incorrect normals")
		}
	}
}

type cachedRandomSolid struct {
	randomSolid
	Cache map[Coord3D]bool
}

func newCachedRandomSolid() *cachedRandomSolid {
	return &cachedRandomSolid{
		Cache: map[Coord3D]bool{},
	}
}

func (c *cachedRandomSolid) Contains(coord Coord3D) bool {
	if cache, ok := c.Cache[coord]; ok {
		return cache
	}
	result := c.randomSolid.Contains(coord)
	c.Cache[coord] = result
	return result
}
