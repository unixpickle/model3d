package render3d

import (
	"math"
	"math/rand"

	"github.com/unixpickle/model3d/model3d"
)

// BidirPathTracer is a bidirectional path traceb.
//
// Lights in the path tracer should be part of the scene,
// but should also be provided as an AreaLight.
type BidirPathTracer struct {
	Camera *Camera

	// Light is the (possibly joined) area light that is
	// sampled for light-to-eye paths.
	Light AreaLight

	// MaxDepth is the maximum number of edges in a in
	// either direction.
	MaxDepth int

	// These fields control how many samples are taken per
	// pixel of the image.
	// See RecursiveRayTracer for more details.
	NumSamples           int
	MinSamples           int
	MaxStddev            float64
	OversaturatedStddevs float64

	// RouletteDelta is the maximum intensity for roulette
	// sampling to be performed.
	// If zero, all deterministic connections are checked.
	RouletteDelta float64

	// See RecursiveRayTracer for more details.
	Cutoff    float64
	Antialias float64
	Epsilon   float64
	LogFunc   func(frac float64, sampleRate float64)
}

// Render renders the object to an image.
func (b *BidirPathTracer) Render(img *Image, obj Object) {
	b.rayRenderer().Render(img, obj)
}

// RayVariance estimates the variance of the color
// components in the rendered image for a single sample.
//
// The variance is averaged over every color component in
// the image.
func (b *BidirPathTracer) RayVariance(obj Object, width, height, samples int) float64 {
	return b.rayRenderer().RayVariance(obj, width, height, samples)
}

func (b *BidirPathTracer) rayRenderer() *rayRenderer {
	return &rayRenderer{
		RayColor: b.rayColor,

		Camera:               b.Camera,
		NumSamples:           b.NumSamples,
		MinSamples:           b.MinSamples,
		MaxStddev:            b.MaxStddev,
		OversaturatedStddevs: b.OversaturatedStddevs,
		Antialias:            b.Antialias,
		LogFunc:              b.LogFunc,
	}
}

func (b *BidirPathTracer) rayColor(gen *rand.Rand, obj Object, ray *model3d.Ray) Color {
	eye := b.sampleEyePath(gen, obj, ray)
	light := b.sampleLightPath(gen, obj)
	var totalColor Color
	allPathCombinations(eye, light, func(p *bptLightPath, combine bool, p1, p2 model3d.Coord3D) {
		intensity := p.Intensity()
		if intensity.Sum() < 1e-8 {
			return
		}
		color := intensity.Scale(1.0 / p.Density())
		brightness := math.Max(color.X, math.Max(color.Y, color.Z))
		if b.RouletteDelta > 0 && brightness < b.RouletteDelta {
			keepProb := brightness / b.RouletteDelta
			if gen.Float64() > keepProb {
				return
			}
			color = color.Scale(1 / keepProb)
		}

		// Check connectivity.
		if combine {
			ray := b.bounceRay(p1, p2.Sub(p1).Normalize())
			eps := b.Epsilon
			if eps == 0 {
				eps = DefaultEpsilon
			}
			maxDist := p2.Dist(p1) - 2*eps
			if coll, _, ok := obj.Cast(ray); ok && coll.Scale < maxDist {
				return
			}
		}

		totalColor = totalColor.Add(color)
	})
	return totalColor
}

func (b *BidirPathTracer) sampleEyePath(gen *rand.Rand, obj Object, ray *model3d.Ray) *bptEyePath {
	res := &bptEyePath{Origin: ray.Origin}
	mask := NewColor(1.0)
	for i := 0; i < b.MaxDepth && mask.Sum()/3 > b.Cutoff; i++ {
		coll, mat, ok := obj.Cast(ray)
		if !ok {
			break
		}
		point := ray.Origin.Add(ray.Direction.Scale(coll.Scale))
		dest := ray.Direction.Scale(-1).Normalize()
		nextSource := mat.SampleSource(gen, coll.Normal, dest)
		vertex := &bptPathVertex{
			Point:    point,
			Normal:   coll.Normal,
			Source:   nextSource,
			Dest:     dest,
			Emission: mat.Emission(),
			Material: mat,
		}
		vertex.EvalMaterial()
		mask = mask.Mul(vertex.BSDF).Scale(1 / vertex.SourceDensity)
		res.Points = append(res.Points, vertex)
		ray = b.bounceRay(point, nextSource.Scale(-1))
	}
	return res
}

func (b *BidirPathTracer) sampleLightPath(gen *rand.Rand, obj Object) *bptLightPath {
	origin, normal, emission := b.Light.SampleLight(gen)

	dest := sampleAngularDest(gen, normal)
	res := &bptLightPath{
		StartDensity: 1.0 / b.Light.Area(),
		Points: []*bptPathVertex{
			{
				Point:    origin,
				Normal:   normal,
				Source:   normal.Scale(-1),
				Dest:     dest,
				BSDF:     Color{},
				Emission: emission,
			},
		},
	}
	res.Points[0].EvalMaterial()

	ray := b.bounceRay(origin, dest)

	mask := NewColor(1.0)
	for i := 0; i < b.MaxDepth && mask.Sum()/3 > b.Cutoff; i++ {
		coll, mat, ok := obj.Cast(ray)
		if !ok {
			break
		}
		point := ray.Origin.Add(ray.Direction.Scale(coll.Scale))
		source := ray.Direction
		var nextDest model3d.Coord3D
		if a, ok := mat.(AsymMaterial); ok {
			nextDest = a.SampleDest(gen, normal, source)
		} else {
			nextDest = mat.SampleSource(gen, coll.Normal, source.Scale(-1)).Scale(-1)
		}
		vertex := &bptPathVertex{
			Point:    point,
			Normal:   coll.Normal,
			Source:   source,
			Dest:     nextDest,
			Emission: mat.Emission(),
			Material: mat,
		}
		vertex.EvalMaterial()
		mask = mask.Mul(vertex.BSDF).Scale(1 / vertex.DestDensity)
		res.Points = append(res.Points, vertex)
		ray = b.bounceRay(point, nextDest)
	}
	return res
}

func (b *BidirPathTracer) bounceRay(point model3d.Coord3D, dir model3d.Coord3D) *model3d.Ray {
	eps := b.Epsilon
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

type bptPathVertex struct {
	Point  model3d.Coord3D
	Normal model3d.Coord3D

	// Source goes from the light to the eye, even in
	// eye paths.
	Source model3d.Coord3D
	Dest   model3d.Coord3D

	BSDF     Color
	Emission Color

	// Material is nil for vertices generated on a light.
	Material      Material
	SourceDensity float64
	DestDensity   float64
}

func (p *bptPathVertex) EvalMaterial() {
	if p.Material == nil {
		p.DestDensity = 4 * math.Max(0, p.Dest.Dot(p.Normal))
		return
	}
	p.SourceDensity = p.Material.SourceDensity(p.Normal, p.Source, p.Dest)
	if a, ok := p.Material.(AsymMaterial); ok {
		p.DestDensity = a.DestDensity(p.Normal, p.Source, p.Dest)
	} else {
		p.DestDensity = p.Material.SourceDensity(p.Normal, p.Dest.Scale(-1), p.Source.Scale(-1))
	}
	p.BSDF = p.Material.BSDF(p.Normal, p.Source, p.Dest)
}

func (b *bptPathVertex) SourceDot() float64 {
	return math.Abs(b.Normal.Dot(b.Source))
}

func (b *bptPathVertex) DestDot() float64 {
	return math.Abs(b.Normal.Dot(b.Dest))
}

type bptEyePath struct {
	Origin model3d.Coord3D

	// Points goes from the eye onward.
	Points []*bptPathVertex
}

type bptLightPath struct {
	// StartDensity is the inverse of the area.
	StartDensity float64

	// Points goes from the light onward.
	Points []*bptPathVertex
}

// Intensity measures the observed light, assuming the
// path is actually connected.
func (b *bptLightPath) Intensity() Color {
	result := b.Points[0].Emission
	for _, p := range b.Points[1:] {
		result = result.Mul(p.BSDF)
		result = result.Scale(math.Abs(p.Normal.Dot(p.Source)))
	}
	return result
}

// Density computes the total sampling density of the path
// using multiple importance sampling.
func (b *bptLightPath) Density() float64 {
	density := newRunningProduct()
	for _, p := range b.Points[1:] {
		density = density.Mul(p.SourceDensity)
	}

	// Density of doing a regular backward path trace.
	totalDensity := density.Value()

	outArea := func(i1, i2 int) float64 {
		diff := b.Points[i1].Point.Sub(b.Points[i2].Point)
		return 4 * math.Pi * diff.Dot(diff)
	}
	cosIn := func(i int) float64 {
		return b.Points[i].SourceDot()
	}
	cosOut := func(i int) float64 {
		return b.Points[i].DestDot()
	}

	if len(b.Points) > 1 {
		density = density.Mul(b.StartDensity)

		// Density of selecting the point on the light.
		density = density.Div(b.Points[1].SourceDensity)
		totalDensity += density.Mul(outArea(0, 1)).Div(cosOut(0)).Value()

		// Densities of starting a path on the light and
		// connecting it to the eye path.
		for i, p := range b.Points[2:] {
			density = density.Div(p.SourceDensity)
			density = density.Mul(b.Points[i].DestDensity)
			density = density.Div(cosOut(i))
			density = density.Mul(cosIn(i + 1))
			totalDensity += density.Mul(outArea(i+1, i+2)).Div(cosOut(i + 1)).Value()
		}
	}
	return totalDensity
}

// allPathCombinations enumerates all the ways to combine
// the two paths, calling f for each combination along
// with the two points whose connectivity to check.
func allPathCombinations(eye *bptEyePath, light *bptLightPath,
	f func(path *bptLightPath, needsCombine bool, p1, p2 model3d.Coord3D)) {
	outPath := &bptLightPath{
		StartDensity: light.StartDensity,
		Points:       make([]*bptPathVertex, len(eye.Points)+len(light.Points)),
	}
	for i := 1; i <= len(eye.Points); i++ {
		subEye := bptEyePath{
			Origin: eye.Origin,
			Points: eye.Points[:i],
		}
		for j := 0; j <= len(light.Points); j++ {
			subLight := bptLightPath{
				StartDensity: light.StartDensity,
				Points:       light.Points[:j],
			}
			if j == 0 {
				combinePaths(subEye, subLight, outPath)
				f(outPath, false, model3d.Coord3D{}, model3d.Coord3D{})
			} else {
				combinePaths(subEye, subLight, outPath)
				f(outPath, true, subEye.Points[len(subEye.Points)-1].Point,
					subLight.Points[len(subLight.Points)-1].Point)
			}
		}
	}
}

func combinePaths(eye bptEyePath, light bptLightPath, result *bptLightPath) {
	result.Points = result.Points[:0]
	if len(light.Points) == 0 {
		result.Points = append(result.Points, eye.Points[len(eye.Points)-1])
	} else {
		for _, p := range light.Points[:len(light.Points)-1] {
			result.Points = append(result.Points, p)
		}
		p := light.Points[len(light.Points)-1]
		dest := eye.Points[len(eye.Points)-1].Point.Sub(p.Point).Normalize()
		vertex := &bptPathVertex{
			Point:    p.Point,
			Normal:   p.Normal,
			Source:   p.Source,
			Dest:     dest,
			Emission: p.Emission,
			Material: p.Material,
		}
		vertex.EvalMaterial()
		result.Points = append(result.Points, vertex)

		p = eye.Points[len(eye.Points)-1]
		vertex1 := &bptPathVertex{
			Point:    p.Point,
			Normal:   p.Normal,
			Source:   vertex.Dest,
			Dest:     p.Dest,
			Emission: p.Emission,
			Material: p.Material,
		}
		vertex1.EvalMaterial()
		result.Points = append(result.Points, vertex1)
	}

	for i := len(eye.Points) - 2; i >= 0; i-- {
		result.Points = append(result.Points, eye.Points[i])
	}
}

func sampleAngularDest(gen *rand.Rand, normal model3d.Coord3D) model3d.Coord3D {
	return (&LambertMaterial{}).SampleSource(gen, normal, normal).Scale(-1)
}

type runningProduct struct {
	numZeros int
	exponent int
	product  float64
}

func newRunningProduct() runningProduct {
	return runningProduct{product: 0.5, exponent: 1}
}

func (r runningProduct) Mul(x float64) runningProduct {
	if x == 0 {
		r.numZeros++
	} else {
		frac, exp := math.Frexp(x)
		r.exponent += exp
		r.product *= frac
	}
	return r
}

func (r runningProduct) Div(x float64) runningProduct {
	if x == 0 {
		r.numZeros--
	} else {
		frac, exp := math.Frexp(x)
		r.exponent -= exp
		r.product /= frac
	}
	return r
}

func (r runningProduct) Value() float64 {
	if r.numZeros > 0 {
		return 0
	} else if r.numZeros < 0 {
		if r.product < 0 {
			return math.Inf(-1)
		} else {
			return math.Inf(1)
		}
	}
	return math.Ldexp(r.product, r.exponent)
}
