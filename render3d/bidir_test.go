package render3d

import (
	"math"
	"math/rand"
	"runtime"
	"testing"

	"github.com/unixpickle/model3d/model3d"
)

func TestBidirPathTracer(t *testing.T) {
	scene, light := testingScene()
	camera := NewCameraAt(model3d.Coord3D{Y: -17, Z: 2}, model3d.Z(2), math.Pi/3.6)

	pt := &RecursiveRayTracer{
		Camera: camera,
		FocusPoints: []FocusPoint{
			&SphereFocusPoint{
				Center: model3d.Coord3D{Z: 5, Y: -19},
				Radius: 1,
			},
			&SphereFocusPoint{
				Center: model3d.Coord3D{X: 3, Z: 5, Y: -19},
				Radius: 0.5,
			},
		},
		FocusPointProbs: []float64{0.2, 0.1},
		MaxDepth:        10,
		NumSamples:      100000,
		MinSamples:      1000,
		MaxStddev:       0.002,
	}

	groundTruth := NewImage(4, 4)
	pt.Render(groundTruth, scene)

	bpt := &BidirPathTracer{
		Camera:     camera,
		Light:      light,
		MaxDepth:   10,
		NumSamples: 100000,
		MinSamples: 1000,
		MaxStddev:  0.002,
	}

	runTest := func(t *testing.T) {
		rand.Seed(0)
		actual := NewImage(4, 4)
		bpt.Render(actual, scene)

		for i, a := range actual.Data {
			x := groundTruth.Data[i]
			if a.Dist(x) > 0.02 || math.IsNaN(a.Sum()) || math.IsInf(a.Sum(), 0) {
				t.Errorf("expected %v but got %v", x, a)
			}
		}
	}

	t.Run("Normal", runTest)

	bpt.MinDepth = 1
	bpt.NumSamples *= 5
	t.Run("RoulettePath", runTest)
}

func BenchmarkBidirPathTracer(b *testing.B) {
	oldProcs := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(oldProcs)

	scene, light := testingScene()
	camera := NewCameraAt(model3d.Coord3D{Y: -17, Z: 2}, model3d.Z(2), math.Pi/3.6)
	bpt := &BidirPathTracer{
		Camera:     camera,
		Light:      light,
		MaxDepth:   10,
		NumSamples: b.N,
	}
	actual := NewImage(5, 5)

	b.ResetTimer()
	bpt.Render(actual, scene)
}

func testingScene() (Object, AreaLight) {
	light := JoinAreaLights(
		NewSphereAreaLight(&model3d.Sphere{
			Center: model3d.Coord3D{Z: 5, Y: -19},
			Radius: 1,
		}, NewColor(100.0)),
		NewSphereAreaLight(&model3d.Sphere{
			Center: model3d.Coord3D{X: 3, Z: 5, Y: -19},
			Radius: 0.5,
		}, NewColor(130.0)),
	)
	scene := JoinedObject{
		&ColliderObject{
			Collider: model3d.MeshToCollider(
				model3d.NewMeshRect(
					model3d.XYZ(-10, -10, -10),
					model3d.XYZ(10, 20, 0),
				).Scale(-1),
			),
			Material: &LambertMaterial{
				DiffuseColor: NewColor(0.3),
			},
		},
		light,
	}
	return scene, light
}
