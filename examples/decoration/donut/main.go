package main

import (
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	NumSprinkles    = 50
	SprinkleLength  = 0.06
	SprinkleRadius  = 0.02
	SprinkleEpsilon = 0.0025
	SprinkleSpacing = 0.15
	ColorEpsilon    = 0.01
)

func main() {
	log.Println("Creating solid...")
	sprinkles := NewSprinkles()
	donut := DonutBody()
	solid := model3d.JoinedSolid{
		donut,
		sprinkles.Solid,
	}

	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.005, 8)

	log.Println("Decimating mesh...")
	d := &model3d.Decimator{
		FeatureAngle:     0.03,
		BoundaryDistance: 1e-5,
		PlaneDistance:    1e-4,
		FilterFunc:       ColorFilter(mesh, ColorFunc(sprinkles)),
	}
	prev := len(mesh.TriangleSlice())
	mesh = d.Decimate(mesh)
	post := len(mesh.TriangleSlice())
	log.Printf("Went from %d -> %d triangles", prev, post)

	log.Println("Rendering...")
	render3d.SaveRendering("rendering.png", mesh, model3d.YZ(4, 4), 600, 600, ColorFunc(sprinkles))

	log.Println("Exporting...")
	f, err := os.Create("donut.zip")
	essentials.Must(err)
	defer f.Close()

	colorFunc := ColorFunc(sprinkles)
	triColor := model3d.VertexColorsToTriangle(func(c model3d.Coord3D) [3]float64 {
		r, g, b := render3d.RGB(colorFunc(c, model3d.RayCollision{}))
		return [3]float64{r, g, b}
	})
	model3d.WriteMaterialOBJ(f, mesh.TriangleSlice(), triColor)
}

func DonutBody() *model3d.Torus {
	return &model3d.Torus{
		Axis:        model3d.Z(1),
		InnerRadius: 0.5,
		OuterRadius: 1.0,
	}
}

func ColorFunc(s *Sprinkles) render3d.ColorFunc {
	icing := InsideIcingFunc()
	return func(c model3d.Coord3D, rc model3d.RayCollision) render3d.Color {
		if s.EpsSolid.Contains(c) {
			return s.Color(c)
		} else if icing(c) {
			// return render3d.NewColorRGB(0.501, 0.85, 0.82)
			return render3d.NewColorRGB(0.97, 0.4, 0.83)
		} else {
			return render3d.NewColorRGB(0.92, 0.76, 0.2)
		}
	}
}

func ColorFilter(m *model3d.Mesh, cf render3d.ColorFunc) func(c model3d.Coord3D) bool {
	changed := map[model3d.Coord3D]bool{}
	m.Iterate(func(t *model3d.Triangle) {
		for _, seg := range t.Segments() {
			rc := model3d.RayCollision{}
			if cf(seg[0], rc) != cf(seg[1], rc) {
				changed[seg[0]] = true
				changed[seg[1]] = true
			}
		}
	})
	points := make([]model3d.Coord3D, 0, len(changed))
	for p := range changed {
		points = append(points, p)
	}
	tree := model3d.NewCoordTree(points)
	return func(c model3d.Coord3D) bool {
		return !tree.SphereCollision(c, ColorEpsilon)
	}
}

func InsideIcingFunc() func(c model3d.Coord3D) bool {
	thetaFunc := NoisyMonotonicFunc()
	return func(c model3d.Coord3D) bool {
		if c.Z < 0 {
			return false
		}
		c2 := c.XY()
		r := c2.Norm()
		theta := math.Atan2(c2.Y, c2.X)
		// outerRadius := 1.3 + 0.15*math.Cos(math.Pow(math.Sin(theta), 2)*5)
		outerRadius := 1.35 + 0.1*math.Cos(thetaFunc(theta+math.Pi)*5)
		innerRadius := 0.6 + 0.05*math.Sin(theta*3)
		return r > innerRadius && r < outerRadius
	}
}

func NoisyMonotonicFunc() func(x float64) float64 {
	res := model2d.BezierCurve{}
	for i := 0; i < 5; i++ {
		x := float64(i) / 4
		y := x
		if i > 0 && i < 4 {
			y += math.Sin(x*1232424.0) * 0.2
		}
		res = append(res, model2d.XY(x*math.Pi*2, y*math.Pi*2))
	}
	return res.EvalX
}

type Sprinkles struct {
	Solid        model3d.Solid
	EpsSolid     model3d.Solid
	EpsSprinkles []model3d.Solid
	Colors       []render3d.Color
}

func NewSprinkles() *Sprinkles {
	donutCollider := &model3d.SolidCollider{
		Solid:               DonutBody(),
		Epsilon:             0.01,
		NormalBisectEpsilon: 1e-5,
	}
	res := &Sprinkles{}
	centers := []model3d.Coord3D{}
	solids := model3d.JoinedSolid{}
	icing := InsideIcingFunc()
	for i := 0; i < NumSprinkles; i++ {
		c := model2d.NewCoordRandNorm().Scale(3)
		if !icing(model3d.XYZ(c.X, c.Y, 1)) {
			i--
			continue
		}
		rc, _ := donutCollider.FirstRayCollision(&model3d.Ray{
			Origin:    model3d.XY(c.X, c.Y),
			Direction: model3d.Z(1),
		})
		sprinkleCenter := model3d.XYZ(c.X, c.Y, rc.Scale)

		tooClose := false
		for _, c := range centers {
			if c.Dist(sprinkleCenter) < SprinkleSpacing {
				tooClose = true
			}
		}
		if tooClose {
			i--
			continue
		}
		centers = append(centers, sprinkleCenter)

		basis1, basis2 := rc.Normal.OrthoBasis()
		theta := rand.Float64() * math.Pi * 2
		axis := basis1.Scale(math.Cos(theta)).Add(basis2.Scale(math.Sin(theta)))
		sprinkle := SprinkleSolid(sprinkleCenter, axis, 0)
		epsSprinkle := SprinkleSolid(sprinkleCenter, axis, SprinkleEpsilon)

		solids = append(solids, sprinkle)
		res.EpsSprinkles = append(res.EpsSprinkles, epsSprinkle)
		res.Colors = append(res.Colors, render3d.NewColorRGB(
			rand.Float64(),
			rand.Float64(),
			rand.Float64(),
		))
	}
	res.Solid = solids.Optimize()
	res.EpsSolid = model3d.JoinedSolid(res.EpsSprinkles).Optimize()
	return res
}

func (s *Sprinkles) Color(c model3d.Coord3D) render3d.Color {
	for i, solid := range s.EpsSprinkles {
		if solid.Contains(c) {
			return s.Colors[i]
		}
	}
	return render3d.NewColor(0)
}

func SprinkleSolid(center model3d.Coord3D, axis model3d.Coord3D, eps float64) model3d.Solid {
	p1 := center.Sub(axis.Scale(SprinkleLength/2 + eps))
	p2 := center.Add(axis.Scale(SprinkleLength/2 + eps))
	return model3d.JoinedSolid{
		&model3d.Cylinder{
			P1:     p1,
			P2:     p2,
			Radius: SprinkleRadius + eps,
		},
		&model3d.Sphere{Center: p1, Radius: SprinkleRadius + eps},
		&model3d.Sphere{Center: p2, Radius: SprinkleRadius + eps},
	}
}
