package model2d

const (
	DefaultMedialAxisIters = 32
	DefaultMedialAxisEps   = 1e-8
)

// ProjectMedialAxis projects the point c onto the medial
// axis of the shape defined by SDF p.
//
// The iters argument specifies the number of search steps
// to perform to narrow down the medial axis.
// If 0, DefaultMedialAxisIters is used.
//
// The eps argument specifies how close two points on the
// surface of p must be to be considered the same point.
// If 0, DefaultMedialAxisEps is used.
//
// The bounds of p are used to aid computation. Thus, it
// is important to get tight bounds on the SDF for
// maximally accurate results.
func ProjectMedialAxis(p PointSDF, c Coord, iters int, eps float64) Coord {
	if iters == 0 {
		iters = DefaultMedialAxisIters
	}
	if eps == 0 {
		eps = DefaultMedialAxisEps
	}

	initPoint, initSDF := p.PointSDF(c)
	if initSDF < eps {
		// Randomly perturb c to avoid the boundary and
		// therefore find the normal of p.
		initPoint, initSDF = p.PointSDF(c.Add(NewCoordRandUnit().Scale(eps)))
	}

	maxScale := p.Max().Dist(p.Min())
	if initSDF < 0 {
		// If we start outside the boundary, we may have
		// to search further to find the medial axis.
		maxScale += -initSDF
	}

	minPoint := c
	maxPoint := c.Add(c.Sub(initPoint).Normalize().Scale(maxScale))

	for i := 0; i < iters; i++ {
		c1 := minPoint.Mid(maxPoint)
		curPoint, _ := p.PointSDF(c1)
		if curPoint.Dist(initPoint) < eps {
			minPoint = c1
		} else {
			maxPoint = c1
		}
	}
	// Always be on the conservative side to avoid
	// crossing the medial axis.
	return minPoint
}
