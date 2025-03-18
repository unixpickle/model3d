package toolbox3d

import (
	"math"
	"math/rand"
	"testing"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

func TestHeigthMapInterp(t *testing.T) {
	rand.Seed(0)
	hm := createRandomizedHeightMap()
	for i := 0; i < 1000; i++ {
		c1 := model2d.NewCoordRandBounds(
			hm.Min.Sub(model2d.XY(0.05, 0.05)),
			hm.Max.Add(model2d.XY(0.05, 0.05)),
		)
		c2 := c1.Add(model2d.NewCoordRandUniform().Scale(1e-6))
		h1 := math.Sqrt(hm.HeightSquaredAt(c1))
		h2 := math.Sqrt(hm.HeightSquaredAt(c2))
		if math.Abs(h1-h2) > 1e-4 {
			t.Errorf("going from %v to %v resulted in heights %f, %f", c1, c2, h1, h2)
		}
	}
}

func TestHeightMapAdd(t *testing.T) {
	rand.Seed(0)
	h1 := createRandomizedHeightMap()
	h2 := h1.Copy()
	hAdd := createRandomizedHeightMap()

	h1.AddHeightMap(hAdd)
	hAdd.Min = hAdd.Min.Add(model2d.XY(1e-8, -1e-8))
	hAdd.Max = hAdd.Max.Add(model2d.XY(1e-8, -1e-8))
	h2.AddHeightMap(hAdd)

	for i, x := range h1.Data {
		a := h2.Data[i]
		if math.Abs(x-a) > 1e-4 {
			t.Fatalf("unexpected interpolation: got %f but expected %f", a, x)
		}
	}
}

func TestHeightMapAddSphere(t *testing.T) {
	rand.Seed(0)
	h := NewHeightMap(model2d.XY(-1, -1), model2d.XY(1, 1), 1000)
	h.AddSphere(model2d.XY(0.1, 0.1), 0.3)

	expectedSDF := &model3d.Sphere{Center: model3d.XY(0.1, 0.1), Radius: 0.3}
	actualMesh := model3d.MarchingCubesSearch(HeightMapToSolidBidir(h), 0.01, 8)
	actualSDF := model3d.MeshToSDF(actualMesh)

	for i := 0; i < 1000; i++ {
		coord := model3d.NewCoord3DRandNorm()
		actual := actualSDF.SDF(coord)
		expected := expectedSDF.SDF(coord)
		if math.Abs(actual-expected) > 0.01 {
			t.Errorf("unexpected SDF at %v (expected %f but got %f)",
				coord, expected, actual)
		}
	}
}

func TestHeightMapMesh(t *testing.T) {
	rand.Seed(0)
	h := NewHeightMap(model2d.XY(-1, -1), model2d.XY(1, 1), 100)
	for i := 0; i < h.Rows; i++ {
		for j := 0; j < h.Cols; j++ {
			if rand.Intn(2) == 0 {
				h.Data[j+i*h.Cols] = rand.Float64() * 2
			}
		}
	}

	testMesh := func(t *testing.T, mesh *model3d.Mesh) {
		if mesh.NeedsRepair() {
			t.Error("mesh has bad edges")
		}
		if n := len(mesh.SingularVertices()); n != 0 {
			t.Errorf("mesh has %d singular vertices", n)
		}
		if _, n := mesh.RepairNormals(1e-5); n != 0 {
			t.Errorf("mesh contains %d bad normals", n)
		}
	}

	t.Run("Flat", func(t *testing.T) {
		testMesh(t, h.Mesh())
	})
	t.Run("Bidir", func(t *testing.T) {
		testMesh(t, h.MeshBidir())
	})
}

func createRandomizedHeightMap() *HeightMap {
	result := NewHeightMap(model2d.XY(0.1, 0.2), model2d.XY(0.3, 0.7), 1000)
	for i := 0; i < rand.Intn(100)+10; i++ {
		center := model2d.NewCoordRandBounds(result.Min, result.Max)
		result.AddSphere(center, rand.Float64()*0.05)
	}
	return result
}

func BenchmarkHeightMapMesh(b *testing.B) {
	h := NewHeightMap(model2d.XY(-1, -1), model2d.XY(1, 1), 100)
	h.AddSphere(model2d.XY(0.2, 0.2), 0.3)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h.Mesh()
	}
}
