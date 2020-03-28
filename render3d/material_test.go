package render3d

import (
	"fmt"
	"math"
	"testing"

	"github.com/unixpickle/model3d"
)

func TestLambertMaterialSampling(t *testing.T) {
	testMaterialSampling(t, &LambertMaterial{
		DiffuseColor: Color{X: 1, Y: 0.9, Z: 0.5},
	})
}

func TestPhongMaterialSampling(t *testing.T) {
	for _, alpha := range []float64{0, 0.5, 2} {
		t.Run(fmt.Sprintf("Alpha%.1f", alpha), func(t *testing.T) {
			testMaterialSampling(t, &PhongMaterial{
				Alpha:         alpha,
				SpecularColor: Color{X: 1, Y: 0.9, Z: 0.5},
			})
		})
	}
	t.Run("Diffuse", func(t *testing.T) {
		testMaterialSampling(t, &PhongMaterial{
			Alpha:         2,
			SpecularColor: Color{X: 1, Y: 0.9, Z: 0.5},
			DiffuseColor:  Color{X: 0.3, Y: 0.2, Z: 0.5},
		})
	})
}

func TestRefractPhongMaterialSampling(t *testing.T) {
	for _, alpha := range []float64{0, 0.5, 2} {
		for _, index := range []float64{1.3, 0.7} {
			t.Run(fmt.Sprintf("Alpha%.1fIndex%.1f", alpha, index), func(t *testing.T) {
				testMaterialSampling(t, &RefractPhongMaterial{
					Alpha:             alpha,
					IndexOfRefraction: index,
					RefractColor:      Color{X: 1, Y: 0.9, Z: 0.5},
				})
			})
		}
	}
}

func TestRefractPhongMaterialFlux(t *testing.T) {
	mat := &RefractPhongMaterial{
		Alpha: 1000.0,
		// This flux test stops being passed for higher
		// indices of refraction, where a lot of flux is
		// lost due to approximation errors when a ray
		// enters the material (but not exits).
		IndexOfRefraction: 1.3,
		RefractColor:      Color{X: 1, Y: 1, Z: 1},
	}

	measureFlux := func(normal, source model3d.Coord3D) float64 {
		var totalFlux1 float64
		var count float64
		for i := 0; i < 10000000; i++ {
			dest := mat.SampleSource(normal, source)
			weight := 1 / mat.SourceDensity(normal, dest, source)
			if math.Abs(source.Dot(normal)) < 1e-3 || math.Abs(dest.Dot(normal)) < 1e-3 {
				i--
				continue
			}
			totalFlux1 += weight * mat.BSDF(normal, source, dest).X * math.Abs(dest.Dot(normal))
			count++
		}
		return totalFlux1 / count
	}

	normal := model3d.NewCoord3DRandUnit()
	source := normal.Add(model3d.NewCoord3DRandUnit().Scale(0.4)).Normalize()
	flux1 := measureFlux(normal, source)
	source = normal.Add(model3d.NewCoord3DRandUnit().Scale(0.4)).Normalize().Scale(-1)
	flux2 := measureFlux(normal, source)

	if math.Abs(flux1-flux2) > 1e-2 {
		t.Error("flux on both sides of normal should match:", flux1, flux2)
	}
	if math.Abs(1-(flux1+flux2)/2) > 0.02 {
		t.Error("flux should be nearly 1, but got:", (flux1+flux2)/2)
	}
}

func testMaterialSampling(t *testing.T, m Material) {
	sourceColorFunc := func(source model3d.Coord3D) Color {
		return Color{
			X: source.X + 2*source.Y*source.Y + 3*source.Z*source.Z*source.Z,
			Y: source.Z - source.X + source.Y,
			Z: 1,
		}
	}

	normal := model3d.NewCoord3DRandUnit()
	dest := model3d.NewCoord3DRandUnit()
	for dest.Dot(normal) < 0.1 {
		dest = model3d.NewCoord3DRandUnit()
	}

	var actual Color
	for i := 0; i < 4000000; i++ {
		source := model3d.NewCoord3DRandUnit()
		reflection := m.BSDF(normal, source, dest)
		color := reflection.Mul(sourceColorFunc(source))
		actual = actual.Add(color)
	}

	var expected Color
	for i := 0; i < 4000000; i++ {
		source := m.SampleSource(normal, dest)
		weight := 1 / m.SourceDensity(normal, source, dest)
		reflection := m.BSDF(normal, source, dest)
		color := reflection.Mul(sourceColorFunc(source)).Scale(weight)
		expected = expected.Add(color)
	}

	if actual.Sub(expected).Norm() > actual.Norm()*0.01 {
		t.Errorf("expected %f but got %f", expected, actual)
	}
}
