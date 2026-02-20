// Generated from templates/metaball.template

package model2d

import "math"

// MetaballFalloffFunc is a function that determines how
// the influence of metaballs falls off outside their
// surface.
type MetaballFalloffFunc func(r float64) float64

// InversePowerMetaballFalloffFunc creates a falloff
// function of the form 1/r^power for r>0.
//
// For r<=0, the returned function always produces +Inf.
//
// power must be strictly positive.
func InversePowerMetaballFalloffFunc(power float64) MetaballFalloffFunc {
	if power <= 0 {
		panic("power must be positive")
	}
	return func(r float64) float64 {
		if r <= 0 {
			return math.Inf(1)
		}
		return math.Pow(r, -power)
	}
}

var (
	linearMetaballFalloffFunc    = InversePowerMetaballFalloffFunc(1)
	quadraticMetaballFalloffFunc = InversePowerMetaballFalloffFunc(2)
	cubicMetaballFalloffFunc     = InversePowerMetaballFalloffFunc(3)
	quarticMetaballFalloffFunc   = InversePowerMetaballFalloffFunc(4)
	quinticMetaballFalloffFunc   = InversePowerMetaballFalloffFunc(5)
)

// LinearMetaballFalloffFunc implements 1/r falloff.
func LinearMetaballFalloffFunc(r float64) float64 {
	return linearMetaballFalloffFunc(r)
}

// QuadraticMetaballFalloffFunc implements 1/r^2 falloff.
func QuadraticMetaballFalloffFunc(r float64) float64 {
	return quadraticMetaballFalloffFunc(r)
}

// CubicMetaballFalloffFunc implements 1/r^3 falloff.
func CubicMetaballFalloffFunc(r float64) float64 {
	return cubicMetaballFalloffFunc(r)
}

// QuarticMetaballFalloffFunc implements 1/r^4 falloff.
func QuarticMetaballFalloffFunc(r float64) float64 {
	return quarticMetaballFalloffFunc(r)
}

// QuinticMetaballFalloffFunc implements 1/r^5 falloff.
func QuinticMetaballFalloffFunc(r float64) float64 {
	return quinticMetaballFalloffFunc(r)
}

// ExponentialMetaballFalloffFunc implements exp(-r)
// falloff for r>0 and clamps to 1 for r<=0.
func ExponentialMetaballFalloffFunc(r float64) float64 {
	if r <= 0 {
		return 1
	}
	return math.Exp(-r)
}

// GaussianMetaballFalloffFunc implements exp(-r^2)
// falloff for r>0 and clamps to 1 for r<=0.
func GaussianMetaballFalloffFunc(r float64) float64 {
	if r <= 0 {
		return 1
	}
	return math.Exp(-r * r)
}

// WyvillMetaballFalloffFunc creates a compact-support
// falloff based on (1-(r/d)^2)^2 for 0<=r<d.
//
// The returned falloff is 0 for r>=d.
// d must be strictly positive.
func WyvillMetaballFalloffFunc(d float64) MetaballFalloffFunc {
	if d <= 0 {
		panic("d must be positive")
	}
	d2 := d * d
	return func(r float64) float64 {
		if r <= 0 {
			return math.Inf(1)
		}
		if r >= d {
			return 0
		}
		ratio2 := (r * r) / d2
		value := 1 - ratio2
		return value * value
	}
}

// A Metaball implements a field f(c) where values greater
// than zero indicate points "outside" of some shape, and
// larger values indicate points "further" away.
//
// The values of the field are related to distances from
// the ground truth shape in Euclidean space. This
// relationship is implemented by MetaballDistBound(),
// which provides an upper-bound on Euclidean distance
// given a field value. This makes it possible to bound a
// level set of the field in Euclidean space, provided the
// bounds of the coordinates where f(c) <= 0.
type Metaball interface {
	// Bounder returns the bounds for the volume where
	// MetaballField() may return values <= 0.
	Bounder

	// MetaballField returns the distance, in some possibly
	// transformed space, of c to the metaball surface.
	//
	// Note that this is not an actual distance in
	// Euclidean coordinates, so for example one could have
	// ||f(c) - f(c1)|| > ||c - c1||.
	//
	// This can happen when scaling a metaball, which
	// effectively changes how fast the field increases as
	// points move away from the surface.
	MetaballField(c Coord) float64

	// MetaballDistBound gives, for a distance to the
	// underlying metaball surface, the minimum value that
	// may be returned by MetaballField.
	//
	// This function must be non-decreasing.
	// For any d and t such that MetaballDistBound(d) >= t,
	// it must be the case that MetaballDistBound(d1) >= t
	// for all d1 >= d.
	MetaballDistBound(d float64) float64
}

// MetaballSolid creates a Solid by smoothly combining
// multiple metaballs.
//
// The f argument determines how MetaballField() values are
// converted to values to be summed across multiple
// metaballs. If nil, QuarticMetaballFalloffFunc is used.
//
// The radiusThreshold is passed through f to determine the
// field threshold. When converting a single metaball to
// a solid, radiusThreshold can be thought of as the max
// value of the metaball's field that is contained within
// the solid.
func MetaballSolid(f MetaballFalloffFunc, radiusThreshold float64, m ...Metaball) Solid {
	return SignedMetaballSolid(f, radiusThreshold, m, nil)
}

// SignedMetaballSolid is like MetaballSolid but with
// support for negative metaballs.
func SignedMetaballSolid(f MetaballFalloffFunc, radiusThreshold float64, pos, neg []Metaball) Solid {
	if len(pos) == 0 {
		return NewRect(Origin, Origin)
	}
	if f == nil {
		f = QuarticMetaballFalloffFunc
	}

	threshold := f(radiusThreshold)

	min, max := BoundsUnion(pos)

	// We need to figure out how much to expand the
	// bounding box to ensure that all points will have
	// total field values less than threshold.
	//
	// We assume that all metaballs take up the whole
	// bounding box, and add their lower bound field
	// values (which correspond to upper-bound falloff
	// values).
	valueForOutset := func(x float64) float64 {
		var sum float64
		for _, mb := range pos {
			sum += f(mb.MetaballDistBound(x))
		}
		return sum
	}

	// Exponential growth to find upper bound on bbox expansion.
	maxOutset := max.Dist(min)
	minOutset := maxOutset * 1e-8
	for i := 0; i < 32; i++ {
		if valueForOutset(maxOutset) > threshold {
			maxOutset *= 2.0
		} else {
			break
		}
	}
	if valueForOutset(maxOutset) > threshold {
		panic("could not find maximum outset")
	}

	// Binary search to narrow in on bbox expansion.
	for i := 0; i < 32; i++ {
		midOutset := (minOutset + maxOutset) / 2
		if valueForOutset(midOutset) > threshold {
			minOutset = midOutset
		} else {
			maxOutset = midOutset
		}
	}

	return CheckedFuncSolid(
		min.AddScalar(-maxOutset),
		max.AddScalar(maxOutset),
		func(c Coord) bool {
			var sum float64
			for _, mb := range pos {
				sum += f(mb.MetaballField(c))
			}
			for _, mb := range neg {
				sum -= f(mb.MetaballField(c))
			}
			return sum > threshold
		},
	)
}

type sdfMetaball struct {
	SDF
}

func (s sdfMetaball) MetaballField(c Coord) float64 {
	return -s.SDF.SDF(c)
}

func (s sdfMetaball) MetaballDistBound(d float64) float64 {
	return d
}

// SDFToMetaball creates a Metaball from the SDF.
// The resulting field is equal to the negative SDF.
func SDFToMetaball(s SDF) Metaball {
	return &sdfMetaball{SDF: s}
}
