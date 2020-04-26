package render3d

import (
	"math"
	"math/rand"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model3d"
)

// A rayRenderer renders objects using any algorithm that
// can render pixels given an outgoing ray.
type rayRenderer struct {
	RayColor func(gen *rand.Rand, obj Object, ray *model3d.Ray) Color

	Camera               *Camera
	NumSamples           int
	MinSamples           int
	MaxStddev            float64
	OversaturatedStddevs float64
	Antialias            float64
	LogFunc              func(frac float64, sampleRate float64)
}

func (r *rayRenderer) Render(img *Image, obj Object) {
	if r.NumSamples == 0 {
		panic("must set NumSamples to non-zero for rayRenderer")
	}
	maxX := float64(img.Width) - 1
	maxY := float64(img.Height) - 1
	caster := r.Camera.Caster(maxX, maxY)

	progressCh := make(chan int, 1)
	go func() {
		mapCoordinates(img.Width, img.Height, func(gen *rand.Rand, x, y, idx int) {
			color, numSamples := r.estimateColor(gen, obj, float64(x), float64(y), caster)
			img.Data[idx] = color
			progressCh <- numSamples
		})
		close(progressCh)
	}()

	updateInterval := essentials.MaxInt(1, img.Width*img.Height/1000)
	var pixelsComplete int
	var samplesTaken int
	for n := range progressCh {
		if r.LogFunc != nil {
			pixelsComplete++
			samplesTaken += n
			if pixelsComplete%updateInterval == 0 {
				r.LogFunc(float64(pixelsComplete)/float64(img.Width*img.Height),
					float64(samplesTaken)/float64(pixelsComplete))
			}
		}
	}
}

func (r *rayRenderer) RayVariance(obj Object, width, height, samples int) float64 {
	if samples < 2 {
		panic("need to take at least two samples")
	}

	maxX := float64(width) - 1
	maxY := float64(height) - 1
	caster := r.Camera.Caster(maxX, maxY)

	var totalVariance float64
	var totalCount float64

	gen := rand.New(rand.NewSource(rand.Int63()))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			ray := model3d.Ray{
				Origin:    r.Camera.Origin,
				Direction: caster(float64(x), float64(y)),
			}
			var colorSum Color
			var colorSqSum Color
			for i := 0; i < samples; i++ {
				sampleColor := r.RayColor(gen, obj, &ray)
				colorSum = colorSum.Add(sampleColor)
				colorSqSum = colorSqSum.Add(sampleColor.Mul(sampleColor))
			}
			mean := colorSum.Scale(1 / float64(samples))
			variance := colorSqSum.Scale(1 / float64(samples)).Sub(mean.Mul(mean))

			// Bessel's correction.
			variance = variance.Scale(float64(samples) / float64(samples-1))

			totalVariance += variance.Sum()
			totalCount += 3
		}
	}
	return totalVariance / totalCount
}

func (r *rayRenderer) estimateColor(gen *rand.Rand, obj Object, x, y float64,
	caster func(x, y float64) model3d.Coord3D) (sampleMean Color, numSamples int) {
	ray := model3d.Ray{Origin: r.Camera.Origin}
	ray.Direction = caster(x, y)
	var colorSum Color
	var colorSqSum Color

SampleLoop:
	for numSamples = 0; numSamples < r.NumSamples; numSamples++ {
		if r.Antialias != 0 {
			dx := r.Antialias * (gen.Float64() - 0.5)
			dy := r.Antialias * (gen.Float64() - 0.5)
			ray.Direction = caster(x+dx, y+dy)
		}
		sampleColor := r.RayColor(gen, obj, &ray)
		colorSum = colorSum.Add(sampleColor)

		if r.MinSamples == 0 || r.MaxStddev == 0 {
			continue
		}

		colorSqSum = colorSqSum.Add(sampleColor.Mul(sampleColor))

		if numSamples < r.MinSamples || numSamples < 2 {
			continue
		}

		mean := colorSum.Scale(1 / float64(numSamples))
		variance := colorSqSum.Scale(1 / float64(numSamples)).Sub(mean.Mul(mean))
		populationRescale := math.Sqrt(float64(numSamples)) / float64(numSamples-1)
		meanArr := mean.Array()
		for i, variance := range variance.Array() {
			if variance < 0 {
				// Variance is so low that our estimate is
				// actually negative due to rounding error.
				continue
			}
			stddev := math.Sqrt(variance) * populationRescale
			switch true {
			case stddev < r.MaxStddev:
			case r.OversaturatedStddevs != 0 &&
				meanArr[i]-r.OversaturatedStddevs*stddev > 1:
			default:
				continue SampleLoop
			}
		}

		// Early stopping due to statistical constraints
		// being satisfied.
		break
	}
	return colorSum.Scale(1 / float64(numSamples)), numSamples
}
