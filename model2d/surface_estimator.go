// Generated from templates/surface_estimator.template

package model2d

const (
	DefaultSurfaceEstimatorBisectCount         = 32
	DefaultSurfaceEstimatorNormalSamples       = 40
	DefaultSurfaceEstimatorNormalBisectEpsilon = 1e-4
	DefaultSurfaceEstimatorNormalNoiseEpsilon  = 1e-4
)

const solidSurfaceEstimatorMaxNormalRetries = 10

// SolidSurfaceEstimator estimates collision points and
// normals on the surface of a solid using search.
type SolidSurfaceEstimator struct {
	// The Solid to estimate the surface of.
	Solid Solid

	// BisectCount, if non-zero, specifies the number of
	// bisections to use in Bisect().
	// Default is DefaultSurfaceEstimatorBisectCount.
	BisectCount int

	// NormalSamples, if non-zero, specifies how many
	// samples to use to approximate normals.
	// Default is DefaultSurfaceEstimatorNormalSamples.
	NormalSamples int

	// RandomSearchNormals can be set to true to disable
	// the binary search to compute normals. Instead, an
	// evolution strategy is performed to estimate the
	// gradient by sampling random points at a distance
	// of NormalNoiseEpsilon.
	RandomSearchNormals bool

	// NormalBisectEpsilon, if non-zero, specifies a small
	// distance to use in a bisection-based method to
	// compute approximate normals.
	//
	// The value must be larger than the distance between
	// the surface and points passed to Normal().
	//
	// Default is DefaultSurfaceEstimatorNormalBisectionEpsilon.
	NormalBisectEpsilon float64

	// NormalNoiseEpsilon, if non-zero, specifies a small
	// distance to use in an evolution strategy when
	// RandomSearchNormals is true.
	//
	// The value must be larger than the distance between
	// the surface and points passed to Normal().
	//
	// Default is DefaultSurfaceEstimatorNormalNoiseEpsilon.
	NormalNoiseEpsilon float64

	// AllowNormalBisectFailure, if true, prevents a panic()
	// if normal bisection cannot find directions that pass
	// through the surface.
	//
	// If this is true and normal computation fails, an
	// arbitrary normal direction is returned.
	AllowNormalBisectFailure bool
}

// BisectInterp returns alpha in [min, max] to minimize the
// surface's distance to p1 + alpha * (p2 - p1).
//
// It is assumed that p1 is outside the surface and p2 is
// inside, and that min < max.
func (s *SolidSurfaceEstimator) BisectInterp(p1, p2 Coord, min, max float64) float64 {
	d := p2.Sub(p1)
	count := s.bisectCount()
	for i := 0; i < count; i++ {
		f := (min + max) / 2
		if s.Solid.Contains(p1.Add(d.Scale(f))) {
			max = f
		} else {
			min = f
		}
	}
	return (min + max) / 2
}

// Bisect finds the point between p1 and p2 closest to the
// surface, provided that p1 and p2 are on different sides.
func (s *SolidSurfaceEstimator) Bisect(p1, p2 Coord) Coord {
	var alpha float64
	if s.Solid.Contains(p1) {
		alpha = 1 - s.BisectInterp(p2, p1, 0, 1)
	} else {
		alpha = s.BisectInterp(p1, p2, 0, 1)
	}
	return p1.Add(p2.Sub(p1).Scale(alpha))
}

// Normal computes the normal at a point on the surface.
// The point must be guaranteed to be on the boundary of
// the surface, e.g. from Bisect().
func (s *SolidSurfaceEstimator) Normal(c Coord) Coord {
	if s.RandomSearchNormals {
		return s.esNormal(c)
	} else {
		return s.bisectNormal(c, 0)
	}
}

func (s *SolidSurfaceEstimator) esNormal(c Coord) Coord {
	eps := s.normalNoiseEpsilon()
	count := s.normalSamples()
	if count < 1 {
		panic("need at least one sample to estimate normal with random search")
	}

	var normalSum Coord
	for i := 0; i < count; i++ {
		delta := NewCoordRandUnit()
		c1 := c.Add(delta.Scale(eps))
		if s.Solid.Contains(c1) {
			normalSum = normalSum.Sub(delta)
		} else {
			normalSum = normalSum.Add(delta)
		}
	}
	return normalSum.Normalize()
}

func (s *SolidSurfaceEstimator) bisectNormal(c Coord, retries int) Coord {
	if retries > solidSurfaceEstimatorMaxNormalRetries {
		if s.AllowNormalBisectFailure {
			return NewCoordRandUnit()
		} else {
			panic("could not estimate normal")
		}
	}
	count := s.normalSamples()
	eps := s.normalBisectEpsilon()

	if count < 4 {
		panic("require at least 4 samples to estimate normals with bisection")
	}
	v1 := NewCoordRandUnit().Scale(eps)
	v2 := XY(v1.Y, -v1.X)
	if !s.Solid.Contains(c.Add(v1)) {
		v1 = v1.Scale(-1)
	}
	if s.Solid.Contains(c.Add(v2)) {
		v2 = v2.Scale(-1)
	}
	for j := 2; j < count; j++ {
		mp := v1.Add(v2).Normalize().Scale(eps)
		if s.Solid.Contains(c.Add(mp)) {
			v1 = mp
		} else {
			v2 = mp
		}
	}
	tangent := v1.Add(v2).Normalize()
	res := XY(tangent.Y, -tangent.X)

	if s.Solid.Contains(c.Add(res.Scale(eps))) {
		return res.Scale(-1)
	} else {
		return res
	}
}

func (s *SolidSurfaceEstimator) bisectCount() int {
	if s.BisectCount == 0 {
		return DefaultSurfaceEstimatorBisectCount
	}
	return s.BisectCount
}

func (s *SolidSurfaceEstimator) normalSamples() int {
	if s.NormalSamples == 0 {
		return DefaultSurfaceEstimatorNormalSamples
	}
	return s.NormalSamples
}

func (s *SolidSurfaceEstimator) normalBisectEpsilon() float64 {
	if s.NormalBisectEpsilon == 0 {
		return DefaultSurfaceEstimatorNormalBisectEpsilon
	}
	return s.NormalBisectEpsilon
}

func (s *SolidSurfaceEstimator) normalNoiseEpsilon() float64 {
	if s.NormalNoiseEpsilon == 0 {
		return DefaultSurfaceEstimatorNormalNoiseEpsilon
	}
	return s.NormalNoiseEpsilon
}
