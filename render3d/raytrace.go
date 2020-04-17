package render3d

import (
	"math"
	"math/rand"
	"runtime"
	"sync"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
)

const DefaultEpsilon = 1e-8

// A RecursiveRayTracer renders objects using recursive
// tracing with random sampling.
type RecursiveRayTracer struct {
	Camera *Camera
	Lights []*PointLight

	// FocusPoints are functions which cause rays to
	// bounce more in certain directions, with the aim of
	// reducing variance with no bias.
	FocusPoints []FocusPoint

	// FocusPointProbs stores, for each FocusPoint, the
	// probability that this focus point is used to sample
	// a ray (rather than the BRDF).
	FocusPointProbs []float64

	// MaxDepth is the maximum number of recursions.
	// Setting to 0 is almost equivalent to RayCast, but
	// the ray tracer still checks for shadows.
	MaxDepth int

	// NumSamples is the number of rays to sample.
	NumSamples int

	// MinSamples and MaxStddev control early stopping for
	// pixel sampling. If they are both non-zero, then
	// MinSamples rays are sampled, and then more rays are
	// sampled until the pixel standard deviation goes
	// below MaxStddev, or NumSamples samples are taken.
	MinSamples int
	MaxStddev  float64

	// OversaturatedStddevs controls how few samples are
	// taken at bright parts of the scene.
	//
	// If specified, a pixel may stop being sampled after
	// MinSamples samples if the brightness of that pixel
	// is more than OversaturatedStddevs standard
	// deviations above the maximum brightness (1.0).
	//
	// This can override MaxStddev, since bright parts of
	// the image may have high standard deviations despite
	// having uninteresting specific values.
	OversaturatedStddevs float64

	// Cutoff is the maximum brightness for which
	// recursion is performed. If small but non-zero, the
	// number of rays traced can be reduced.
	Cutoff float64

	// Antialias, if non-zero, specifies a fraction of a
	// pixel to perturb every ray's origin.
	// Thus, 1 is maximum, and 0 means no change.
	Antialias float64

	// Epsilon is a small distance used to move away from
	// surfaces before bouncing new rays.
	// If nil, DefaultEpsilon is used.
	Epsilon float64

	// LogFunc, if specified, is called periodically with
	// progress information.
	//
	// The frac argument specifies the fraction of pixels
	// which have been colored.
	//
	// The sampleRate argument specifies the mean number
	// of rays traced per pixel.
	LogFunc func(frac float64, sampleRate float64)
}

// Render renders the object to an image.
func (r *RecursiveRayTracer) Render(img *Image, obj Object) {
	if r.NumSamples == 0 {
		panic("must set NumSamples to non-zero for RecursiveRayTracer")
	}
	maxX := float64(img.Width) - 1
	maxY := float64(img.Height) - 1
	caster := r.Camera.Caster(maxX, maxY)

	coords := make(chan [3]int, img.Width*img.Height)
	var idx int
	for y := 0; y < img.Height; y++ {
		for x := 0; x < img.Width; x++ {
			coords <- [3]int{x, y, idx}
			idx++
		}
	}
	close(coords)

	progressCh := make(chan int, 1)

	var wg sync.WaitGroup
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gen := rand.New(rand.NewSource(rand.Int63()))
			for c := range coords {
				color, numSamples := r.estimateColor(gen, obj, float64(c[0]), float64(c[1]), caster)
				img.Data[c[2]] = color
				progressCh <- numSamples
			}
		}()
	}

	go func() {
		wg.Wait()
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

// RayVariance estimates the variance of the color
// components in the rendered image for a single ray path.
// It is intended to be used to quickly judge how well
// importance sampling is working.
//
// The variance is averaged over every color component in
// the image.
func (r *RecursiveRayTracer) RayVariance(obj Object, width, height, samples int) float64 {
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
				sampleColor := r.recurse(gen, obj, &ray, 0, Color{X: 1, Y: 1, Z: 1})
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

func (r *RecursiveRayTracer) estimateColor(gen *rand.Rand, obj Object, x, y float64,
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
		sampleColor := r.recurse(gen, obj, &ray, 0, Color{X: 1, Y: 1, Z: 1})
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

func (r *RecursiveRayTracer) recurse(gen *rand.Rand, obj Object, ray *model3d.Ray,
	depth int, scale Color) Color {
	if scale.Sum()/3 < r.Cutoff {
		return Color{}
	}
	collision, material, ok := obj.Cast(ray)
	if !ok {
		return Color{}
	}
	point := ray.Origin.Add(ray.Direction.Scale(collision.Scale))

	dest := ray.Direction.Normalize().Scale(-1)
	color := material.Emission()
	if depth == 0 {
		// Only add ambient light directly to object, not to
		// recursive rays.
		color = color.Add(material.Ambient())
	}
	for _, l := range r.Lights {
		lightDirection := l.Origin.Sub(point)

		shadowRay := r.bounceRay(point, lightDirection)
		shadowCollision, _, ok := obj.Cast(shadowRay)
		if ok && shadowCollision.Scale < 1 {
			continue
		}

		brdf := material.BSDF(collision.Normal, point.Sub(l.Origin).Normalize(), dest)
		color = color.Add(l.ShadeCollision(collision.Normal, lightDirection).Mul(brdf))
	}
	if depth >= r.MaxDepth {
		return color
	}
	nextSource := r.sampleNextSource(gen, point, collision.Normal, dest, material)
	weight := 1 / r.sourceDensity(point, collision.Normal, nextSource, dest, material)
	weight *= math.Abs(nextSource.Dot(collision.Normal))
	reflectWeight := material.BSDF(collision.Normal, nextSource, dest)
	nextRay := r.bounceRay(point, nextSource.Scale(-1))
	nextMask := reflectWeight.Scale(weight)
	nextScale := scale.Mul(nextMask)
	nextColor := r.recurse(gen, obj, nextRay, depth+1, nextScale)
	return color.Add(nextColor.Mul(nextMask))
}

func (r *RecursiveRayTracer) sampleNextSource(gen *rand.Rand, point, normal, dest model3d.Coord3D,
	mat Material) model3d.Coord3D {
	if len(r.FocusPoints) == 0 {
		return mat.SampleSource(gen, normal, dest)
	} else if len(r.FocusPoints) != len(r.FocusPointProbs) {
		panic("FocusPoints and FocusPointProbs must match in length")
	}

	p := gen.Float64()
	for i, prob := range r.FocusPointProbs {
		p -= prob
		if p < 0 {
			return r.FocusPoints[i].SampleFocus(gen, mat, point, normal, dest)
		}
	}

	return mat.SampleSource(gen, normal, dest)
}

func (r *RecursiveRayTracer) sourceDensity(point, normal, source, dest model3d.Coord3D,
	mat Material) float64 {
	if len(r.FocusPoints) == 0 {
		return mat.SourceDensity(normal, source, dest)
	}

	matProb := 1.0
	var prob float64
	for i, focusProb := range r.FocusPointProbs {
		prob += focusProb * r.FocusPoints[i].FocusDensity(mat, point, normal, source, dest)
		matProb -= focusProb
	}

	return prob + matProb*mat.SourceDensity(normal, source, dest)
}

func (r *RecursiveRayTracer) bounceRay(point model3d.Coord3D, dir model3d.Coord3D) *model3d.Ray {
	eps := r.Epsilon
	if eps == 0 {
		eps = DefaultEpsilon
	}
	return &model3d.Ray{
		// Prevent a duplicate collision from being
		// detected when bouncing off an existing
		// object.
		Origin:    point.Add(dir.Normalize().Scale(eps)),
		Direction: dir,
	}
}
