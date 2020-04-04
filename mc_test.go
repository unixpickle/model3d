package model3d

import "testing"

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
