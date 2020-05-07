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

	// MaxLightDepth, if non-zero, limits the number of
	// light path vertices.
	// Set to 1 to simply sample the area lights.
	MaxLightDepth int

	// MinDepth, if non-zero, is the minimum number of
	// edges in a path before roulette sampling.
	//
	// If unspecified, no roulette sampling is performed
	// except for Cutoff.
	MinDepth int

	// These fields control how many samples are taken per
	// pixel of the image.
	// See RecursiveRayTracer for more details.
	NumSamples           int
	MinSamples           int
	MaxStddev            float64
	OversaturatedStddevs float64
	Convergence          func(mean, stddev Color) bool

	// RouletteDelta is the maximum intensity for roulette
	// sampling to be performed.
	// If zero, all deterministic connections are checked.
	RouletteDelta float64

	// PowerHeuristic, if non-zero, is used for multiple
	// importance sampling of paths.
	// A value of 2 is recommended, and a value of 1 is
	// equivalent to the balance heuristic used by
	// default.
	PowerHeuristic float64

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

// RenderVariance computes the variance per pixel using a
// fixed number of rays per pixel, and writes the results
// as pixels in an image.
func (b *BidirPathTracer) RenderVariance(img *Image, obj Object, numSamples int) {
	b.rayRenderer().RenderVariance(img, obj, numSamples)
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
		Convergence:          b.Convergence,
		Antialias:            b.Antialias,
		LogFunc:              b.LogFunc,
	}
}

func (b *BidirPathTracer) rayColor(g *goInfo, obj Object, ray *model3d.Ray) Color {
	if g.Extra == nil {
		g.Extra = newBptPathCache(b.MaxDepth + b.maxLightDepth())
	}
	cache := g.Extra.(*bptPathCache)

	b.sampleEyePath(g.Gen, obj, ray, cache.EyePath)
	b.sampleLightPath(g.Gen, obj, cache.LightPath)

	var totalColor Color
	totalEmission := b.Light.TotalEmission()
	allPathCombinations(cache.EyePath, cache.LightPath, cache, totalEmission,
		func(density float64, intensity Color, p1, p2 model3d.Coord3D) {
			if intensity.Sum() < 1e-8 {
				return
			}

			p := cache.JoinedPath
			var weight float64
			if b.PowerHeuristic == 0 {
				p.Densities(totalEmission, b.MaxDepth, b.MaxLightDepth, func(d float64) {
					weight += d
				})
			} else {
				scale := math.Pow(density, -(b.PowerHeuristic-1)/b.PowerHeuristic)
				p.Densities(totalEmission, b.MaxDepth, b.MaxLightDepth, func(d float64) {
					weight += math.Pow(d*scale, b.PowerHeuristic)
				})
			}
			color := intensity.Scale(1.0 / weight)

			if p1 != p2 {
				// Roulette sampling only when a collision
				// check is needed.
				brightness := math.Max(color.X, math.Max(color.Y, color.Z))
				if b.RouletteDelta > 0 && brightness < b.RouletteDelta {
					keepProb := brightness / b.RouletteDelta
					if g.Gen.Float64() > keepProb {
						return
					}
					color = color.Scale(1 / keepProb)
				}

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

func (b *BidirPathTracer) sampleEyePath(gen *rand.Rand, obj Object, ray *model3d.Ray,
	out *bptEyePath) {
	out.Clear()
	pathEnder := newBptPathEnder(b.MinDepth, b.Cutoff)
	for i := 0; i < b.MaxDepth; i++ {
		coll, mat, ok := obj.Cast(ray)
		if !ok {
			break
		}
		point := ray.Origin.Add(ray.Direction.Scale(coll.Scale))
		dest := ray.Direction.Scale(-1).Normalize()
		nextSource := mat.SampleSource(gen, coll.Normal, dest)
		vertex := out.Extend()
		*vertex = bptPathVertex{
			Point:         point,
			Normal:        coll.Normal,
			Source:        nextSource,
			Dest:          dest,
			Emission:      mat.Emission(),
			Material:      mat,
			RouletteScale: pathEnder.RouletteScale(),
		}
		vertex.EvalMaterial()
		ray = b.bounceRay(point, nextSource.Scale(-1))
		if pathEnder.End(gen, i, vertex.BSDF.Scale(vertex.SourceDot()/vertex.SourceDensity)) {
			break
		}
	}
}

func (b *BidirPathTracer) sampleLightPath(gen *rand.Rand, obj Object, out *bptLightPath) {
	origin, normal, emission := b.Light.SampleLight(gen)

	dest := sampleAngularDest(gen, normal)
	out.Clear()
	*out.Extend() = bptPathVertex{
		Point:         origin,
		Normal:        normal,
		Source:        normal.Scale(-1),
		Dest:          dest,
		BSDF:          Color{},
		Emission:      emission,
		RouletteScale: 1.0,
	}
	out.Last().EvalMaterial()

	ray := b.bounceRay(origin, dest)

	pathEnder := newBptPathEnder(b.MinDepth, b.Cutoff)
	for i := 0; i < b.maxLightDepth()-1; i++ {
		coll, mat, ok := obj.Cast(ray)
		if !ok {
			break
		}
		point := ray.Origin.Add(ray.Direction.Scale(coll.Scale))
		source := ray.Direction
		nextDest := SampleDest(mat, gen, coll.Normal, source)
		vertex := out.Extend()
		*vertex = bptPathVertex{
			Point:         point,
			Normal:        coll.Normal,
			Source:        source,
			Dest:          nextDest,
			Emission:      mat.Emission(),
			Material:      mat,
			RouletteScale: pathEnder.RouletteScale(),
		}
		vertex.EvalMaterial()
		ray = b.bounceRay(point, nextDest)
		if pathEnder.End(gen, i, vertex.BSDF.Scale(vertex.DestDot()/vertex.DestDensity)) {
			break
		}
	}
}

func (b *BidirPathTracer) maxLightDepth() int {
	if b.MaxLightDepth != 0 {
		return b.MaxLightDepth
	}
	return b.MaxDepth
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

func (b *BidirPathTracer) pathEnder() bptPathEnder {
	return bptPathEnder{
		MinLength: b.MinDepth,
		Cutoff:    b.Cutoff,
	}
}

type bptPathEnder struct {
	MinLength int
	Cutoff    float64

	currentRoulette float64
	fullMask        Color
	rouletteMask    Color
}

func newBptPathEnder(minLength int, cutoff float64) bptPathEnder {
	return bptPathEnder{
		MinLength:       minLength,
		Cutoff:          cutoff,
		currentRoulette: 1.0,
		fullMask:        NewColor(1),
		rouletteMask:    NewColor(1),
	}
}

func (b *bptPathEnder) RouletteScale() float64 {
	return b.currentRoulette
}

func (b *bptPathEnder) End(gen *rand.Rand, i int, mask Color) bool {
	b.fullMask = b.fullMask.Mul(mask)
	if mean := b.fullMask.Sum() / 3; mean < b.Cutoff {
		keepProb := mean / b.Cutoff
		if gen.Float64() > keepProb {
			return true
		}
		b.currentRoulette *= 1 / keepProb
	}
	if b.MinLength != 0 && i+1 >= b.MinLength {
		b.rouletteMask = b.rouletteMask.Mul(mask)
		maxVal := math.Max(math.Max(b.rouletteMask.X, b.rouletteMask.Y), b.rouletteMask.Z)
		if maxVal < 1 {
			b.rouletteMask = NewColor(1)
			keepProb := maxVal
			if gen.Float64() > keepProb {
				return true
			}
			b.currentRoulette *= 1 / keepProb
		}
	}
	return false
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

	// RouletteScale is >= 1 and indicates how unlikely
	// this vertex was to be reached due to roulette
	// sampling.
	RouletteScale float64

	// Accumulator is used internally by density
	// calculations but nothing else.
	Accumulator float64
}

func (p *bptPathVertex) EvalMaterial() {
	if p.Material == nil {
		p.DestDensity = 4 * math.Max(0, p.Dest.Dot(p.Normal))
		return
	}
	p.SourceDensity = p.Material.SourceDensity(p.Normal, p.Source, p.Dest)
	p.DestDensity = DestDensity(p.Material, p.Normal, p.Source, p.Dest)
	p.BSDF = p.Material.BSDF(p.Normal, p.Source, p.Dest)
}

func (b *bptPathVertex) SourceDot() float64 {
	return math.Abs(b.Normal.Dot(b.Source))
}

func (b *bptPathVertex) DestDot() float64 {
	return math.Abs(b.Normal.Dot(b.Dest))
}

type bptPathCache struct {
	EyePath    *bptEyePath
	LightPath  *bptLightPath
	JoinedPath *bptLightPath
	Extra      [2]bptPathVertex
}

func newBptPathCache(maxVertices int) *bptPathCache {
	allVerts := make([]bptPathVertex, maxVertices*2)

	var slices [2][]*bptPathVertex
	for i := range slices {
		slice := make([]*bptPathVertex, maxVertices)
		for j := range slice {
			slice[j] = &allVerts[i*maxVertices+j]
		}
		slices[i] = slice
	}
	return &bptPathCache{
		EyePath:    &bptEyePath{bptPath{Points: slices[0]}},
		LightPath:  &bptLightPath{bptPath{Points: slices[1]}},
		JoinedPath: &bptLightPath{bptPath{Points: make([]*bptPathVertex, 0, maxVertices)}},
	}
}

type bptPath struct {
	Points []*bptPathVertex
}

func (b *bptPath) Clear() {
	b.Points = b.Points[:0]
}

func (b *bptPath) Extend() *bptPathVertex {
	idx := len(b.Points)
	b.Points = b.Points[:idx+1]
	return b.Points[idx]
}

func (b *bptPath) Last() *bptPathVertex {
	return b.Points[len(b.Points)-1]
}

func (b *bptPath) Push(v *bptPathVertex) {
	b.Points = append(b.Points, v)
}

type bptEyePath struct {
	// Points go from the eye onward.
	//
	// The eye itself is not included.
	bptPath
}

type bptLightPath struct {
	// Points go from the light onward.
	//
	// The light is the first vertex.
	// If the path was generated from a light source,
	// then the material of this vertex is nil.
	bptPath
}

// Densities computes the sampling density of the path for
// each possible way it could have been sampled.
func (b *bptLightPath) Densities(totalLight float64, maxDepth, maxLightDepth int, f func(float64)) {
	if maxLightDepth == 0 {
		maxLightDepth = maxDepth
	}

	sourceDensityProduct := 1.0
	for i := len(b.Points) - 1; i > 0; i-- {
		p := b.Points[i]
		p.Accumulator = sourceDensityProduct
		sourceDensityProduct *= p.SourceDensity
	}

	// Density of doing a regular backward path trace.
	if len(b.Points) <= maxDepth {
		f(sourceDensityProduct)
	}

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
		lightDensity := b.Points[0].Emission.Sum() / totalLight

		// Density of selecting the point on the light and
		// connecting it to an eye path.
		if len(b.Points)-1 <= maxDepth {
			f(lightDensity * b.Points[1].Accumulator * outArea(0, 1) / cosOut(0))
		}

		// Densities of starting a path on the light and
		// connecting it to the eye path.
		for i, p := range b.Points[2:] {
			if i+1 >= maxLightDepth {
				break
			}
			lightDensity *= b.Points[i].DestDensity
			lightDensity *= cosIn(i+1) / cosOut(i)
			if len(b.Points)-(i+2) <= maxDepth {
				f(p.Accumulator * lightDensity * outArea(i+1, i+2) / cosOut(i+1))
			}
		}
	}
}

// allPathCombinations enumerates all the ways to combine
// the two paths, calling f for each combination along
// with the two points whose connectivity to check.
func allPathCombinations(eye *bptEyePath, light *bptLightPath, c *bptPathCache, totalLight float64,
	f func(density float64, intensity Color, p1, p2 model3d.Coord3D)) {
	out := c.JoinedPath
	eyeDensity := 1.0
	eyeBSDF := NewColor(1.0)
	for i := 1; i <= len(eye.Points); i++ {
		subEye := bptEyePath{bptPath{Points: eye.Points[:i]}}
		if (subEye.Points[i-1].Emission != Color{}) {
			// Full light path has some contribution.
			combinePaths(subEye, bptLightPath{}, c)
			f(eyeDensity, subEye.Points[i-1].Emission.Mul(eyeBSDF),
				model3d.Coord3D{}, model3d.Coord3D{})
		}
		density := eyeDensity * light.Points[0].Emission.Sum() / totalLight
		lightBSDF := light.Points[0].Emission
		for j := 1; j <= len(light.Points); j++ {
			diff := light.Points[j-1].Point.Sub(subEye.Points[i-1].Point)
			outArea := 4 * math.Pi * diff.Dot(diff)

			if j > 1 {
				density *= light.Points[j-2].DestDensity
				density *= light.Points[j-1].SourceDot() / light.Points[j-2].DestDot()
				if j > 2 {
					lightBSDF = lightBSDF.Mul(light.Points[j-2].BSDF)
				}
				lightBSDF = lightBSDF.Scale(light.Points[j-1].SourceDot())
			}

			subLight := bptLightPath{bptPath{Points: light.Points[:j]}}
			combinePaths(subEye, subLight, c)

			// Prevent zero dot products at connection points,
			// since the contribution will always be zero.
			// If we don't do this check, then the power heuristic
			// may compute infinity/infinity and yield NaNs.
			if destDot := out.Points[j-1].DestDot(); destDot > 0 {
				if sourceDot := out.Points[j].SourceDot(); sourceDot > 0 {
					curDensity := density * outArea / destDot
					intensity := eyeBSDF.Mul(lightBSDF).Scale(sourceDot)
					intensity = intensity.Mul(out.Points[j].BSDF)
					intensity = intensity.Scale(light.Points[j-1].RouletteScale *
						eye.Points[i-1].RouletteScale)
					if j > 1 {
						intensity = intensity.Mul(out.Points[j-1].BSDF)
					}
					f(curDensity, intensity, subEye.Points[len(subEye.Points)-1].Point,
						subLight.Points[len(subLight.Points)-1].Point)
				}
			}
		}
		eyeDensity *= eye.Points[i-1].SourceDensity
		eyeBSDF = eyeBSDF.Mul(eye.Points[i-1].BSDF).Scale(eye.Points[i-1].SourceDot())
	}
}

func combinePaths(eye bptEyePath, light bptLightPath, c *bptPathCache) {
	result := c.JoinedPath
	result.Clear()
	if len(light.Points) == 0 {
		result.Push(eye.Points[len(eye.Points)-1])
	} else {
		for _, p := range light.Points[:len(light.Points)-1] {
			result.Push(p)
		}
		p := light.Points[len(light.Points)-1]
		dest := eye.Points[len(eye.Points)-1].Point.Sub(p.Point).Normalize()
		vertex := &c.Extra[0]
		*vertex = bptPathVertex{
			Point:    p.Point,
			Normal:   p.Normal,
			Source:   p.Source,
			Dest:     dest,
			Emission: p.Emission,
			Material: p.Material,
		}
		vertex.EvalMaterial()
		result.Push(vertex)

		p = eye.Points[len(eye.Points)-1]
		vertex1 := &c.Extra[1]
		*vertex1 = bptPathVertex{
			Point:    p.Point,
			Normal:   p.Normal,
			Source:   vertex.Dest,
			Dest:     p.Dest,
			Emission: p.Emission,
			Material: p.Material,
		}
		vertex1.EvalMaterial()
		result.Push(vertex1)
	}

	for i := len(eye.Points) - 2; i >= 0; i-- {
		result.Push(eye.Points[i])
	}
}

func sampleAngularDest(gen *rand.Rand, normal model3d.Coord3D) model3d.Coord3D {
	return (&LambertMaterial{}).SampleSource(gen, normal, normal).Scale(-1)
}
