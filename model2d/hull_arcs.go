package model2d

import (
	"math"
	"sort"

	"github.com/unixpickle/splaytree"
)

type ArcHullArc struct {
	// Radius may be zero when this is a point.
	Circle

	// Theta values both between [-pi, pi], going clockwise,
	// where end is exclusive and start is inclusive.
	Start, End float64
}

func (a *ArcHullArc) StartCoord() Coord {
	return a.Center.Add(XY(math.Cos(a.Start), math.Sin(a.Start)).Scale(a.Radius))
}

func (a *ArcHullArc) EndCoord() Coord {
	return a.Center.Add(XY(math.Cos(a.End), math.Sin(a.End)).Scale(a.Radius))
}

func (a *ArcHullArc) ContainsTheta(theta float64) bool {
	return arcContains(a.Start, a.End, theta)
}

func (a *ArcHullArc) WithinCircle(c *Circle) (Coord, bool) {
	const eps = 1e-8

	containsPoint := func(p Coord) bool {
		return p.Dist(c.Center) <= c.Radius+eps
	}

	pointTheta := func(p Coord) float64 {
		diff := p.Sub(a.Center)
		return diff.Atan2()
	}

	bestPoint := Origin
	bestDist := math.Inf(1)
	found := false

	tryPoint := func(p Coord, requireOnArc bool) {
		if requireOnArc {
			theta := pointTheta(p)
			if !a.ContainsTheta(theta) {
				return
			}
		}
		if !containsPoint(p) {
			return
		}
		dist := p.Dist(c.Center)
		if !found || dist < bestDist {
			bestPoint = p
			bestDist = dist
			found = true
		}
	}

	// Always test endpoints.
	tryPoint(a.StartCoord(), false)
	tryPoint(a.EndCoord(), false)

	// Test the arc point closest to c.Center, if that angle lies on the arc.
	if a.Radius == 0 {
		if containsPoint(a.Center) {
			return a.Center, true
		}
		return Origin, false
	}
	closestTheta := c.Center.Sub(a.Center).Atan2()
	if a.ContainsTheta(closestTheta) {
		closestPoint := a.Center.Add(XY(math.Cos(closestTheta), math.Sin(closestTheta)).Scale(a.Radius))
		tryPoint(closestPoint, false)
	}

	// If the two circles intersect, test the intersection points too.
	d := c.Center.Dist(a.Center)
	if d != 0 && d <= c.Radius+a.Radius+eps && d+math.Min(c.Radius, a.Radius)+eps >= math.Max(c.Radius, a.Radius) {
		x := (d*d + a.Radius*a.Radius - c.Radius*c.Radius) / (2 * d)

		h2 := a.Radius*a.Radius - x*x
		if h2 < 0 && h2 > -eps {
			h2 = 0
		}
		if h2 >= 0 {
			h := math.Sqrt(h2)
			dir := c.Center.Sub(a.Center).Scale(1 / d)
			perp := XY(-dir.Y, dir.X)
			base := a.Center.Add(dir.Scale(x))

			p1 := base.Add(perp.Scale(h))
			p2 := base.Sub(perp.Scale(h))
			tryPoint(p1, true)
			tryPoint(p2, true)
		}
	}

	if !found {
		return Origin, false
	}
	return bestPoint, true
}

func (a *ArcHullArc) FirstRayCollision(r *Ray) (collision RayCollision, collides bool) {
	return basicRayColliderFirstRayCollision(a, r)
}

func (a *ArcHullArc) RayCollisions(r *Ray, f func(rc RayCollision)) (count int) {
	if a.Radius == 0 {
		return 0
	}
	(&Circle{Center: a.Center, Radius: a.Radius}).RayCollisions(r, func(rc RayCollision) {
		offset := r.Origin.Add(r.Direction.Scale(rc.Scale)).Sub(a.Center)
		theta := offset.Atan2()
		if a.ContainsTheta(theta) {
			count += 1
			if f != nil {
				f(rc)
			}
		}
	})
	return
}

func (a *ArcHullArc) Compare(a1 *ArcHullArc) int {
	if a.End < a1.End {
		return 1
	} else if a.End > a1.End {
		return -1
	} else if a.Start < a1.Start {
		return 1
	} else if a.Start > a1.Start {
		return -1
	} else if a.Center.X < a1.Center.X {
		return 1
	} else if a.Center.X > a1.Center.X {
		return -1
	} else if a.Center.Y < a1.Center.Y {
		return 1
	} else if a.Center.Y > a1.Center.Y {
		return -1
	} else if a.Radius < a1.Radius {
		return 1
	} else if a.Radius > a1.Radius {
		return -1
	} else {
		return 0
	}
}

func (a *ArcHullArc) pointAbove(p Coord) bool {
	if a.Start == a.End {
		return false
	}
	if p == a.Center {
		return false
	}
	theta := p.Sub(a.Center).Atan2()
	return a.ContainsTheta(theta)
}

func (a *ArcHullArc) collapseIfTinyAngle() {
	if anglesClose(a.Start, a.End) {
		a.End = a.Start
	}
}

type ArcHull struct {
	// Some point within the hull.
	StartCenter Coord

	Tree *splaytree.Tree[*ArcHullArc]
}

func NewArcHull(circles []*Circle) *ArcHull {
	hull := &ArcHull{Tree: &splaytree.Tree[*ArcHullArc]{}}
	if len(circles) == 0 {
		return hull
	}

	sorted := append([]*Circle{}, circles...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Radius > sorted[j].Radius
	})

	first := sorted[0]
	hull.StartCenter = first.Center

	hull.Tree.Insert(&ArcHullArc{Circle: *first, Start: 0, End: -math.Pi})
	hull.Tree.Insert(&ArcHullArc{Circle: *first, Start: -math.Pi, End: 0})
	for i := 1; i < len(sorted); i++ {
		hull.addNextOrderedCircle(sorted[i])
	}

	return hull
}

func (a *ArcHull) addNextOrderedCircle(c *Circle) {
	p, ok := a.findContained(c)
	if !ok {
		// Circle is within the hull already.
		return
	}
	firstArc, secondArc := a.findAtStartCenterAngle(a.startCenterAngle(p))
	if secondArc == nil {
		secondArc = firstArc
	}

	shouldStopClockwise := func(arc *ArcHullArc) (float64, bool) {
		_, theta2, ok := convexHullTouchAngles(c, &arc.Circle)
		if !ok {
			panic("circle is contained in the hull")
		}
		if arc.ContainsTheta(theta2) {
			return theta2, true
		}
		return 0, false
	}
	shouldStopCounterClockwise := func(arc *ArcHullArc) (float64, bool) {
		theta1, _, ok := convexHullTouchAngles(c, &arc.Circle)
		if !ok {
			panic("circle is contained in the hull")
		}
		if arc.ContainsTheta(theta1) {
			return theta1, true
		}
		return 0, false
	}

	deleteArcs := []*ArcHullArc{firstArc}
	if firstArc != secondArc {
		deleteArcs = append(deleteArcs, secondArc)
	}

	a.Tree.Iterate(func(h *ArcHullArc) bool {
		return true
	})

	clockwiseArc := secondArc
	finalClockwiseAngle, stop := shouldStopClockwise(clockwiseArc)
	foundStop := stop
	if !stop {
		splayTreeIterateForwards(a.Tree, clockwiseArc, func(arc *ArcHullArc) bool {
			deleteArcs = append(deleteArcs, arc)
			if angle, stop := shouldStopClockwise(arc); stop {
				finalClockwiseAngle = angle
				clockwiseArc = arc
				foundStop = true
				return false
			}
			return true
		})
	}
	if !foundStop {
		panic("never found end of loop")
	}

	counterClockwiseArc := firstArc
	finalCounterClockwiseAngle, stop := shouldStopCounterClockwise(counterClockwiseArc)
	foundStop = stop
	if !stop {
		splayTreeIterateBackwards(a.Tree, counterClockwiseArc, func(arc *ArcHullArc) bool {
			deleteArcs = append(deleteArcs, arc)
			if angle, stop := shouldStopCounterClockwise(arc); stop {
				finalCounterClockwiseAngle = angle
				counterClockwiseArc = arc
				foundStop = true
				return false
			}
			return true
		})
	}
	if !foundStop {
		panic("never found end of loop")
	}

	for _, arc := range deleteArcs {
		a.Tree.Delete(arc)
	}

	newClockwiseArc := &ArcHullArc{
		Circle: clockwiseArc.Circle,
		Start:  finalClockwiseAngle,
		End:    clockwiseArc.End,
	}
	newCounterClockwiseArc := &ArcHullArc{
		Circle: counterClockwiseArc.Circle,
		Start:  counterClockwiseArc.Start,
		End:    finalCounterClockwiseAngle,
	}
	a.Tree.Insert(newCounterClockwiseArc)
	a.Tree.Insert(newClockwiseArc)
	a.Tree.Insert(&ArcHullArc{
		Circle: *c,
		Start:  finalCounterClockwiseAngle,
		End:    finalClockwiseAngle,
	})
}

// Contains checks if a point is contained inside or exactly on the edge
// of the hull.
func (a *ArcHull) Contains(coord Coord) bool {
	if a.Tree == nil || a.Tree.Root == nil {
		return false
	}
	if coord == a.StartCenter {
		return true
	}
	ray := coord.Sub(a.StartCenter)
	if ray.X == 0 && ray.Y == 0 {
		return true
	}
	rc, ok := a.FirstRayCollision(&Ray{Origin: a.StartCenter, Direction: ray})
	if !ok {
		return false
	}
	return rc.Scale >= 1
}

// findContined finds some point on the current hull that is within c.
func (a *ArcHull) findContained(c *Circle) (Coord, bool) {
	// Simple case: guess a point on c furthest from StartCenter and see if it's
	// outside of the hull.
	if coord, ok := a.findContainedWithGuess(c); ok {
		return coord, true
	}

	// The circle may intersect some pair, and we use "aboveness" to check.
	first, last := a.Tree.Min(), a.Tree.Max()
	if coord, ok := a.findCollisionOnTopOfPair(c, last, first); ok {
		return coord, true
	}

	// Narrow down the chain where the circle may collide with the hull.
	lowerBound := first
	upperBound := last
	node := a.Tree.Root
	var everFound bool
	for node != nil && lowerBound != upperBound {
		if node.Value == lowerBound {
			node = node.Right
			continue
		} else if node.Value == upperBound {
			node = node.Left
			continue
		}

		if coord, ok := node.Value.WithinCircle(c); ok {
			return coord, true
		}

		if a.isCircleOnAboveTop(c, lowerBound, node.Value) {
			upperBound = node.Value
			node = node.Left
		} else {
			lowerBound = node.Value
			node = node.Right
		}
	}
	// Now the chain should be two consecutive arcs.
	if coord, ok := a.findCollisionOnTopOfPair(c, lowerBound, upperBound); ok {
		return coord, true
	}
	if everFound {
		panic("never found!")
	}
	return Origin, false
}

func (a *ArcHull) findContainedWithGuess(c *Circle) (Coord, bool) {
	vector := c.Center.Sub(a.StartCenter)
	n := vector.Norm()
	if n == 0 {
		vector = X(1)
	} else {
		vector = vector.Scale(1 / n)
	}
	guessPoint := c.Center.Add(vector.Scale(c.Radius))
	ray := guessPoint.Sub(a.StartCenter)
	if ray.X == 0 && ray.Y == 0 {
		// The center has zero radius and is touching the start center.
		return Origin, false
	}

	rc, collides := a.FirstRayCollision(&Ray{Origin: a.StartCenter, Direction: ray})
	if !collides {
		// This shouldn't be possible unless the first center had zero radius.
		panic("no ray collision, did first circle have zero radius?")
	}
	if rc.Scale < 1 {
		// The guess point is outside the current convex hull.
		return a.StartCenter.Add(ray.Scale(rc.Scale)), true
	}
	// The guess point is within or exactly on the hull.
	return Origin, false
}

// findCollisionOnTopOfPair is only valid if the two arcs
// are consecutive; not to be used during a narrowing process.
func (a *ArcHull) findCollisionOnTopOfPair(c *Circle, start, end *ArcHullArc) (Coord, bool) {
	if start.Circle == end.Circle {
		arc := &ArcHullArc{Circle: start.Circle, Start: start.Start, End: end.End}
		return arc.WithinCircle(c)
	}

	// If the new circle is already exactly on the hull, then we ignore it.
	if start.Circle == *c || end.Circle == *c {
		return Origin, false
	}

	// If the new circle is completely inside another circle, then we ignore it.
	if c.Center.Dist(start.Center)+c.Radius <= start.Radius {
		return Origin, false
	}
	if c.Center.Dist(end.Center)+c.Radius <= end.Radius {
		return Origin, false
	}

	for _, a := range []*ArcHullArc{start, end} {
		if coord, ok := a.WithinCircle(c); ok {
			return coord, ok
		}
	}

	// Check the connector as well.
	connector := Segment{start.EndCoord(), end.StartCoord()}
	ray := &Ray{Origin: connector[0], Direction: connector[1].Sub(connector[0])}
	if rc, ok := c.FirstRayCollision(ray); ok && rc.Scale < 1 {
		x := ray.Origin.Add(ray.Direction.Scale(rc.Scale))
		return x, true
	}
	return Origin, false
}

func (a *ArcHull) isCircleOnAboveTop(c *Circle, start, end *ArcHullArc) bool {
	_, startToEndTheta, ok := convexHullTouchAngles(&start.Circle, &end.Circle)
	if start.Circle == end.Circle {
		startToEndTheta = start.End
	} else if !ok {
		panic("unexpected circle containment in existing hull")
	}

	topStart := &ArcHullArc{
		Circle: start.Circle,
		Start:  start.End,
		End:    startToEndTheta,
	}
	topEnd := &ArcHullArc{
		Circle: end.Circle,
		Start:  startToEndTheta,
		End:    end.Start,
	}

	// Rounding errors can make tiny arcs look like huge arcs
	// if the start/end angle are out of order.
	topStart.collapseIfTinyAngle()
	topEnd.collapseIfTinyAngle()

	if _, ok := a.findCollisionOnTopOfPair(c, topStart, topEnd); ok {
		return true
	}

	if topStart.pointAbove(c.Center) {
		return true
	}
	if topEnd.pointAbove(c.Center) {
		return true
	}

	p1, p2 := topStart.EndCoord(), topEnd.StartCoord()
	if p1 == p2 {
		return false
	}
	if pointOnOrAboveClockwiseLine(Segment{p1, p2}, c.Center) {
		return true
	}
	return false
}

func (a *ArcHull) FirstRayCollision(r *Ray) (collision RayCollision, collides bool) {
	return basicRayColliderFirstRayCollision(a, r)
}

func (a *ArcHull) RayCollisions(r *Ray, f func(RayCollision)) (count int) {
	if r.Origin == a.StartCenter {
		// Optimization to do collisions in logarithmic time
		angle := r.Direction.Atan2()
		startArc, endArc := a.findAtStartCenterAngle(angle)
		if startArc == nil {
			panic("no collision found from StartCenter")
		}

		tryEndpoints := func(ps [2]Coord) {
			normDir := r.Direction.Normalize()
			var bestDot, bestDist float64
			for _, p := range ps {
				epDir := p.Sub(a.StartCenter)
				epNorm := epDir.Norm()
				dot := epDir.Dot(normDir) / epNorm
				if dot > bestDot {
					bestDot = dot
					bestDist = epNorm
				}
			}
			if bestDot < 0.9999 {
				panic("no ray collision with arc")
			}
			f(RayCollision{
				Scale:  bestDist / r.Direction.Norm(),
				Normal: r.Direction.Normalize().Scale(-1),
			})
		}

		if endArc == nil {
			n := startArc.RayCollisions(r, f)
			if n == 0 {
				tryEndpoints([2]Coord{startArc.StartCoord(), startArc.EndCoord()})
			} else if n == 2 {
				panic("unexpected collision count from inside hull")
			}
			return 1
		}
		segment := Segment{startArc.EndCoord(), endArc.StartCoord()}
		if segment.RayCollisions(r, f) == 0 {
			tryEndpoints(segment)
		}
		return 1
	}

	runSeg := func(prev, cur *ArcHullArc) {
		seg := Segment{prev.EndCoord(), cur.StartCoord()}
		if rc, ok := seg.FirstRayCollision(r); ok {
			count += 1
			if f != nil {
				f(rc)
			}
		}
	}
	var first *ArcHullArc
	var prev *ArcHullArc
	a.Tree.Iterate(func(a *ArcHullArc) bool {
		if first == nil {
			first = a
		} else {
			// Draw a segment from the previous arc to this one.
			runSeg(prev, a)
		}
		count += a.RayCollisions(r, f)
		prev = a
		return true
	})
	if first != prev {
		// Draw a segment from the last to the first arc
		runSeg(prev, first)
	}
	return
}

// Mesh creates a mesh by walking the hull, creating segPerArc
// points for each arc.
func (a *ArcHull) Mesh(segPerArc int) *Mesh {
	var prevPoint Coord
	var firstPoint Coord
	isFirst := true
	hullMesh := NewMesh()
	addPoint := func(c Coord) {
		if isFirst {
			isFirst = false
			firstPoint = c
		} else {
			hullMesh.Add(&Segment{prevPoint, c})
		}
		prevPoint = c
	}
	a.Tree.Iterate(func(h *ArcHullArc) bool {
		// Avoid creating duplicate points.
		if h.Radius == 0 || h.Start == h.End {
			// Avoid creating duplicate coords from walks on zero-radius arcs.
			h.Center.Add(NewCoordPolar(h.Start, h.Radius))
			return true
		}

		totalSize := h.Start - h.End
		if totalSize < 0 {
			totalSize = (h.Start + math.Pi*2 - h.End)
		}
		for t := 0; t < segPerArc; t++ {
			frac := float64(t) / float64(segPerArc-1)
			theta := h.Start - frac*totalSize
			p := h.Center.Add(NewCoordPolar(theta, h.Radius))
			addPoint(p)
		}
		return true
	})
	addPoint(firstPoint)
	return hullMesh
}

// findAtStartCenterAngle finds the ray collision with a ray shooting from the
// StartCenter at a given angle, returning either one arc if the collision hits
// an arc, or a pair of arcs if the collision hits the line between two arcs.
func (a *ArcHull) findAtStartCenterAngle(angle float64) (beforeOrResult, after *ArcHullArc) {
	tryIntersection := func(start, end *ArcHullArc) bool {
		startStart, startEnd := a.startCenterAngles(start)
		endStart, endEnd := a.startCenterAngles(end)
		if arcContains(startStart, startEnd, angle) {
			beforeOrResult, after = start, nil
		} else if arcContains(endStart, endEnd, angle) {
			beforeOrResult, after = end, nil
		} else if start != end && arcContains(startEnd, endStart, angle) {
			beforeOrResult, after = start, end
		} else {
			return false
		}
		return true
	}

	// Check between the last to first segment, otherwise we know
	// they can be used as bounds.
	first, last := a.Tree.Min(), a.Tree.Max()

	if tryIntersection(last, first) {
		return
	}

	lowerBound := first
	upperBound := last
	node := a.Tree.Root
	for node != nil && lowerBound != upperBound {
		if node.Value == lowerBound {
			node = node.Right
			continue
		} else if node.Value == upperBound {
			node = node.Left
			continue
		}

		startAngle, endAngle := a.startCenterAngles(node.Value)
		if arcContains(startAngle, endAngle, angle) {
			return node.Value, nil
		}

		_, lowerEnd := a.startCenterAngles(lowerBound)
		if arcContains(lowerEnd, startAngle, angle) {
			upperBound = node.Value
			node = node.Left
		} else {
			lowerBound = node.Value
			node = node.Right
		}
	}
	tryIntersection(lowerBound, upperBound)
	return
}

func (a *ArcHull) startCenterAngles(arc *ArcHullArc) (start, end float64) {
	return a.startCenterAngle(arc.StartCoord()), a.startCenterAngle(arc.EndCoord())
}

func (a *ArcHull) startCenterAngle(p Coord) float64 {
	v := p.Sub(a.StartCenter)
	return v.Atan2()
}

type basicRayCollider interface {
	RayCollisions(_ *Ray, _ func(RayCollision)) int
}

func basicRayColliderFirstRayCollision(b basicRayCollider, r *Ray) (collision RayCollision, collides bool) {
	b.RayCollisions(r, func(rc RayCollision) {
		if !collides || rc.Scale < collision.Scale {
			collision = rc
			collides = true
		}
	})
	return
}

func normalizeAngle(theta float64) float64 {
	for theta < 0 {
		theta += 2 * math.Pi
	}
	for theta >= 2*math.Pi {
		theta -= 2 * math.Pi
	}
	return theta
}

func clockwiseAngleDelta(theta1, theta2 float64) float64 {
	// Amount you rotate clockwise to go from theta1 to theta2.
	d := normalizeAngle(theta1 - theta2)
	return d
}

func clockwiseMidAngle(theta1, theta2 float64) float64 {
	d := clockwiseAngleDelta(theta1, theta2)
	return normalizeAngle(theta1 - d/2)
}

func clockwiseArcIsOuter(theta1, theta2, toOtherX, toOtherY float64) bool {
	mid := clockwiseMidAngle(theta1, theta2)
	mx := math.Cos(mid)
	my := math.Sin(mid)

	// Outer arc points away from the other center.
	// So midpoint direction should have negative dot product.
	return mx*toOtherX+my*toOtherY < 0
}

// convexHullTouchAngles figures out the two angles that are
// where tangent lines can be drawn between the circles.
//
// They are ordered to be followed theta1, theta2 on c1 and then
// the reverse order on c2 to be clockwise if y-axis points up.
func convexHullTouchAngles(c1, c2 *Circle) (theta1, theta2 float64, ok bool) {
	dx := c2.Center.X - c1.Center.X
	dy := c2.Center.Y - c1.Center.Y
	dr := c1.Radius - c2.Radius

	dist2 := dx*dx + dy*dy
	if dist2 == 0 {
		return 0, 0, false
	}
	if dr*dr > dist2 {
		return 0, 0, false
	}

	dist := math.Sqrt(dist2)
	ex := dx / dist
	ey := dy / dist

	a := dr / dist
	b2 := 1 - a*a
	if b2 < 0 {
		if b2 > -1e-12 {
			b2 = 0
		} else {
			return 0, 0, false
		}
	}
	b := math.Sqrt(b2)

	// Two unit normals for common supporting lines.
	// perp(e) = (-ey, ex)
	n1x := a*ex - b*ey
	n1y := a*ey + b*ex

	n2x := a*ex + b*ey
	n2y := a*ey - b*ex

	theta1 = math.Atan2(n1y, n1x)
	theta2 = math.Atan2(n2y, n2x)

	// Order so that traversing clockwise from theta1 to theta2
	// follows the OUTER hull arc on c1.
	if !clockwiseArcIsOuter(theta1, theta2, dx, dy) {
		theta1, theta2 = theta2, theta1
	}

	return theta1, theta2, true
}

func normalizeTheta(theta float64) float64 {
	for theta <= -math.Pi {
		theta += 2 * math.Pi
	}
	for theta > math.Pi {
		theta -= 2 * math.Pi
	}
	return theta
}

func clockwiseDelta(start, end float64) float64 {
	// Amount you travel going clockwise from start to end.
	return math.Mod(start-end+2*math.Pi, 2*math.Pi)
}

func anglesClose(theta1, theta2 float64) bool {
	minDist := math.Min(clockwiseDelta(theta1, theta2), clockwiseDelta(theta2, theta1))
	return minDist < 1e-8
}

func arcContains(start, end, theta float64) bool {
	if start == end {
		return false
	}
	if start > end {
		// Clockwise without wrapping
		return theta <= start && theta > end
	} else {
		// Clockwise with wrapping, i.e. end > start
		return theta <= start || theta > end
	}
}

func pointOnOrAboveClockwiseLine(seg Segment, p Coord) bool {
	a := seg[0]
	b := seg[1]

	ab := b.Sub(a)
	ap := p.Sub(a)

	// Cross product determines which side of the line p lies on.
	return ab.X*ap.Y-ab.Y*ap.X >= 0
}

func pointAboveArc(center Coord, radius, start, end float64, p Coord) bool {
	startDir := Coord{X: math.Cos(start), Y: math.Sin(start)}
	endDir := Coord{X: math.Cos(end), Y: math.Sin(end)}

	startPt := center.Add(startDir.Scale(radius))
	endPt := center.Add(endDir.Scale(radius))

	// For the clockwise arc from start to end, the outward wedge is:
	//   - on the left side of the start ray
	//   - on the right side of the end ray
	return pointOnOrLeftOfRay(startPt, startDir, p) &&
		pointOnOrRightOfRay(endPt, endDir, p)
}

func pointOnOrLeftOfRay(origin, dir, p Coord) bool {
	rel := p.Sub(origin)
	return dir.X*rel.Y-dir.Y*rel.X >= 0
}

func pointOnOrRightOfRay(origin, dir, p Coord) bool {
	rel := p.Sub(origin)
	return dir.X*rel.Y-dir.Y*rel.X <= 0
}
