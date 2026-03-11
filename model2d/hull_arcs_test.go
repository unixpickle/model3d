package model2d

import (
	"math"
	"math/rand"
	"testing"

	"github.com/unixpickle/splaytree"
)

func TestArcHullArcCircleCollision(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	for trial := 0; trial < 100; trial++ {
		start := rng.Float64()*(math.Pi*2) - math.Pi
		end := rng.Float64()*(math.Pi*2) - math.Pi
		arc := &ArcHullArc{
			Circle: Circle{Center: NewCoordRandNorm(rng), Radius: rng.Float64() + 0.1},
			Start:  start,
			End:    end,
		}

		for i := 0; i < 10; i++ {
			p := NewCoordRandNorm(rng)
			closestDist := math.Inf(1)
			arc.iterPoints(1000)(func(c Coord) bool {
				d := c.Dist(p)
				if d < closestDist {
					closestDist = d
				}
				return true
			})
			if arc.CircleCollision(p, math.Max(closestDist-0.001, 0)) {
				t.Fatal("found incorrect collision")
			}
			if !arc.CircleCollision(p, closestDist+1e-5) {
				t.Fatal("missing correct collision")
			}
		}
	}
}

func TestArcHullTouchAngles(t *testing.T) {
	c1 := &Circle{Radius: 0.5, Center: Y(1)}
	c2 := &Circle{Radius: 1, Center: Y(-1)}
	start, end, ok := convexHullTouchAngles(c1, c2)
	if !ok {
		t.Fatal("missing touch")
	}
	if start < end {
		t.Errorf("unexpected start/end: %f, %f", start, end)
	}
	start, end, ok = convexHullTouchAngles(c2, c1)
	if !ok {
		t.Fatal("missing touch")
	}
	if start > end {
		t.Errorf("unexpected start/end: %f, %f", start, end)
	}

	// For two identical circles,
	c1.Radius = 1
	start, end, ok = convexHullTouchAngles(c1, c2)
	if !ok {
		t.Fatal("missing touch")
	}

	startPointOnC1 := XY(math.Cos(start), math.Sin(start))
	endPointOnC1 := XY(math.Cos(end), math.Sin(end))
	if startPointOnC1.Dist(XY(-1, 0)) > 1e-5 {
		t.Errorf("bad start point: %v", startPointOnC1)
	}
	if endPointOnC1.Dist(XY(1, 0)) > 1e-5 {
		t.Errorf("bad end point: %v", endPointOnC1)
	}
}

func TestArcHullFindAtStartCenterAngle(t *testing.T) {
	gen := rand.New(rand.NewSource(0))
	hull := testingRadialArcHull(t)
	for i := 0; i < 1000; i++ {
		theta := (gen.Float64() - 0.5) * math.Pi * 2
		ray := &Ray{
			Origin:    hull.StartCenter,
			Direction: XY(math.Cos(theta), math.Sin(theta)),
		}
		firstArc, secondArc := hull.findAtStartCenterAngle(theta)
		if firstArc == nil {
			t.Fatal("no collision detected")
		} else if secondArc == nil {
			// We have an arc collision.
			_, ok := firstArc.FirstRayCollision(ray)
			if !ok {
				t.Error("got arc collision from helper but no ray collision detected on arc")
			}
		} else {
			// Collision on a segment.
			segment := Segment{firstArc.EndCoord(), secondArc.StartCoord()}
			if _, ok := segment.FirstRayCollision(ray); !ok {
				t.Error("got segment collision from helper but no collision detected on segment")
			}
		}
	}
}

func TestArcHullContainsMatchesConvexHullMesh(t *testing.T) {
	const (
		numTrials       = 100
		minCircles      = 3
		maxCircles      = 15
		pointsPerCircle = 100
		samplesPerTrial = 10000
		minMatchRate    = 0.99
		minCircleRadius = 0.1
		maxRadiusJitter = 1.5
		maxPoints       = 5
	)

	rng := rand.New(rand.NewSource(1337))
	for trial := 0; trial < numTrials; trial++ {
		numCircles := rng.Intn(maxCircles-minCircles+1) + minCircles
		numPoints := rng.Intn(maxPoints)

		circles := make([]*Circle, numCircles+numPoints)

		for i := 0; i < numCircles; i++ {
			circle := &Circle{
				Center: NewCoordRandNorm(rng),
				Radius: minCircleRadius + rng.Float64()*maxRadiusJitter,
			}
			circles[i] = circle
		}

		for i := numCircles; i < numCircles+numPoints; i++ {
			circle := &Circle{
				Center: NewCoordRandNorm(rng),
				Radius: 0,
			}
			circles[i] = circle
		}

		testArcHullWithMeshComparison(t, circles, rng)
		testArcHullWithPointsOnCircles(t, circles)
	}
}

func TestArcHullWithCirclesOnHull(t *testing.T) {
	const trialsWithMesh = 10
	rng := rand.New(rand.NewSource(1337))
	for trial := 0; trial < 1000; trial++ {
		circles := []*Circle{
			{Radius: 0.1, Center: X(1)},
			{Radius: 0.1, Center: X(-1)},
		}
		for i := 0; i < 10; i++ {
			mesh := approximateArcHull(circles, 100)

			// Sample a random point on the hull.
			totalLength := 0.0
			mesh.Iterate(func(s *Segment) {
				totalLength += s.Length()
			})
			stopAtLength := rng.Float64() * totalLength * 0.99999
			curLen := 0.0
			var p Coord
			mesh.Iterate(func(s *Segment) {
				curLen += s.Length()
				if curLen > stopAtLength {
					alongLength := curLen - stopAtLength
					p = s[0].Add(s[1].Sub(s[0]).Normalize().Scale(alongLength))
				}
			})

			// Add a circle to this point with a random radius.
			circles = append(circles, &Circle{Center: p, Radius: math.Abs(rng.NormFloat64()) + 0.1})

			// Make sure the current hull is valid.
			testArcHullWithPointsOnCircles(t, circles)
			if trial < trialsWithMesh {
				testArcHullWithMeshComparison(t, circles, rng)
			}
		}
	}
}

func TestArcHullEdgeContainment(t *testing.T) {
	t.Run("Case1", func(t *testing.T) {
		circles := []*Circle{
			{Center: X(1), Radius: 0.1},
			{Center: X(-1), Radius: 0.1},
			{Center: XY(-1.3871680344747346, 0.17382331452787636), Radius: 0.164636966334192},
		}
		point := XY(-1.0000387168034475, -0.0999726176685472)
		if !circles[1].Contains(point) {
			t.Fatal("point not contained in circle")
		}
		hull := NewArcHull(circles)
		if !hull.Contains(point) {
			t.Fatal("point not contained in hull")
		}
	})

	t.Run("Case2", func(t *testing.T) {
		circles := []*Circle{
			{Center: X(1), Radius: 0.1},
			{Center: X(-1), Radius: 0.1},
			{Center: XY(4.874573940334379, -2.51231018683631), Radius: 0.1722198627155709},
			{Center: XY(4.096828336914301, 0.5821883465874217), Radius: 0.3760090581875929},
			{Center: XY(-0.13602880511933968, 22.832258374574707), Radius: 0.6101253229599962},
			{Center: XY(4.666116625084469, -4.928822061610319), Radius: 0.6602389499989475},
		}
		point := XY(5.0456492257872805, -2.510836492536864)
		anyContains := false
		for _, c := range circles {
			if c.Contains(point) {
				anyContains = true
			}
		}
		if !anyContains {
			t.Fatal("no circle contains point")
		}
		hull := NewArcHull(circles)
		if !hull.Contains(point) {
			t.Fatal("point not contained in hull")
		}
	})
}

func testArcHullWithMeshComparison(t *testing.T, circles []*Circle, rng *rand.Rand) {
	const (
		pointsPerCircle = 1000
		samplesPerTrial = 10000
		minMatchRate    = 0.99
	)

	meshSolid := approximateArcHull(circles, pointsPerCircle).Solid()
	arcHull := NewArcHull(circles)

	matches := 0
	minPoint, maxPoint := BoundsUnion(circles)
	for i := 0; i < samplesPerTrial; i++ {
		point := NewCoordRandBounds(minPoint, maxPoint, rng)
		if arcHull.Contains(point) == meshSolid.Contains(point) {
			matches++
		}
	}

	matchRate := float64(matches) / float64(samplesPerTrial)
	if matchRate < minMatchRate {
		t.Fatalf("match rate %f < %f", matchRate, minMatchRate)
	}
}

func testArcHullWithPointsOnCircles(t *testing.T, circles []*Circle) {
	const (
		pointsPerCircle = 100
		radiusInset     = 0.9999
	)

	arcHull := NewArcHull(circles)

	for _, c := range circles {
		for j := 0; j < pointsPerCircle; j++ {
			theta := 2 * math.Pi * float64(j) / float64(pointsPerCircle)
			hullPoint := c.Center.Add(NewCoordPolar(theta, c.Radius))
			insetPoint := arcHull.StartCenter.Add(hullPoint.Sub(arcHull.StartCenter).Scale(radiusInset))
			if !arcHull.Contains(insetPoint) {
				t.Fatal("arc point is not inside of hull", insetPoint)
			}
		}
	}
}

func approximateArcHull(circles []*Circle, pointsPerCircle int) *Mesh {
	var hullPoints []Coord

	for _, c := range circles {
		for j := 0; j < pointsPerCircle; j++ {
			theta := 2 * math.Pi * float64(j) / float64(pointsPerCircle)
			hullPoints = append(hullPoints, c.Center.Add(NewCoordPolar(theta, c.Radius)))
		}
	}

	return ConvexHullMesh(hullPoints)
}

func testingRadialArcHull(t *testing.T) *ArcHull {
	circles := make([]*Circle, 10)
	for i := range circles {
		theta := float64(i) / float64(len(circles)) * math.Pi * 2
		circles[i] = &Circle{Radius: 0.1, Center: XY(math.Cos(theta), math.Sin(theta))}
	}
	hull := &ArcHull{
		StartCenter: circles[0].Center,
		Tree:        &splaytree.Tree[*ArcHullArc]{},
	}
	totalAngle := 0.0
	for i, c := range circles {
		prevCircle := circles[(i+len(circles)-1)%len(circles)]
		nextCircle := circles[(i+1)%len(circles)]
		endAngle, _, ok := convexHullTouchAngles(prevCircle, c)
		if !ok {
			t.Fatal("no touch angles")
		}
		startAngle, _, ok := convexHullTouchAngles(c, nextCircle)
		if !ok {
			t.Fatal("no touch angles")
		}
		arc := &ArcHullArc{
			Circle: Circle{
				Center: c.Center,
				Radius: c.Radius,
			},
			Start: startAngle,
			End:   endAngle,
		}
		hull.Tree.Insert(arc)
		totalAngle += clockwiseDelta(arc.Start, arc.End)
	}
	if math.Abs(totalAngle-math.Pi*2) > 1e-8 {
		t.Fatalf("bad total angle: %f", totalAngle)
	}
	return hull
}
