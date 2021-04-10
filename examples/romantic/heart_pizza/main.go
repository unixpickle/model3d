package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	CrustRadius        = 0.15
	CheeseHeight       = 0.08
	PepperoniRadius    = 0.25
	PepperoniThickness = 0.03
	PepperoniEpsilon   = 0.01
)

func main() {
	log.Println("Creating solid...")
	solid := model3d.JoinedSolid{
		GetHeartRim(),
		GetPizzaBase(),
		GetPepperonis(0),
	}

	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.02, 8)

	log.Println("Decimating...")
	dec := model3d.Decimator{
		// Only eliminate nearly-planar surfaces.
		FeatureAngle:  0.01,
		PlaneDistance: 1e-3,
		// Never eliminate the top, since it needs lots of
		// triangles for coloration.
		FilterFunc: func(c model3d.Coord3D) bool {
			return c.Z < 0
		},
	}
	mesh = dec.Decimate(mesh)

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, ColorFunc())

	log.Println("Saving mesh...")
	colorFunc := ColorFunc()
	triColor := model3d.VertexColorsToTriangle(func(c model3d.Coord3D) [3]float64 {
		r, g, b := render3d.RGB(colorFunc(c, model3d.RayCollision{}))
		return [3]float64{r, g, b}
	})
	mesh.SaveMaterialOBJ("pizza.zip", triColor)
}

func GetHeartRim() model3d.Solid {
	outline := GetHeartOutline().Decimate(200)
	segments := []model3d.Segment{}
	outline.Iterate(func(s *model2d.Segment) {
		segments = append(segments, model3d.NewSegment(
			model3d.XY(s[0].X, s[0].Y),
			model3d.XY(s[1].X, s[1].Y),
		))
	})
	return toolbox3d.LineJoin(CrustRadius, segments...)
}

func GetPizzaBase() model3d.Solid {
	outline := GetHeartOutline()
	solid2d := model2d.NewColliderSolid(model2d.MeshToCollider(outline))
	solid3d := model3d.ProfileSolid(solid2d, -CrustRadius, CheeseHeight)
	return solid3d
}

func GetHeartOutline() *model2d.Mesh {
	mesh := model2d.MustReadBitmap("heart.png", nil).FlipY().Mesh().SmoothSq(30).Scale(0.0015)
	return mesh.Translate(mesh.Min().Mid(mesh.Max()).Scale(-1))
}

func GetPepperonis(expand float64) model3d.Solid {
	centers := []model2d.Coord{
		model2d.XY(-1, 0.7),
		model2d.XY(0.8, 0.8),
		model2d.XY(1.2, 0.4),
		model2d.XY(0.4, -0.7),
		model2d.XY(-0.6, -0.2),
	}
	res := model3d.JoinedSolid{}
	for _, c := range centers {
		res = append(res, &model3d.Cylinder{
			P1:     model3d.XYZ(c.X, c.Y, CheeseHeight-0.01),
			P2:     model3d.XYZ(c.X, c.Y, CheeseHeight+PepperoniThickness+expand),
			Radius: PepperoniRadius + expand,
		})
	}
	return res.Optimize()
}

func ColorFunc() render3d.ColorFunc {
	rim := GetHeartRim()
	outline := model2d.NewColliderSolid(model2d.MeshToCollider(GetHeartOutline()))

	densities := make([]func(c model2d.Coord) float64, 5)
	for i := range densities {
		densities[i] = RandCheeseDensity()
	}
	pepperonis := GetPepperonis(PepperoniEpsilon)

	return func(c model3d.Coord3D, rc model3d.RayCollision) render3d.Color {
		if pepperonis.Contains(c) {
			return render3d.NewColorRGB(0.75, 0.35, 0.25)
		} else if c.Z < CheeseHeight+0.01 && c.Z > CheeseHeight-0.01 &&
			!rim.Contains(c) && outline.Contains(c.XY()) {
			// Cheesy side of the pizza.
			cheeseDensity := 0.0
			for _, d := range densities {
				cheeseDensity += d(c.XY())
			}
			cheeseFrac := cheeseDensity / float64(len(densities))

			// Limit the number of unique colors.
			cheeseFrac = math.Round(cheeseFrac*15.0) / 15.0

			cheeseColor := render3d.NewColorRGB(1, 1, 0)
			sauceColor := render3d.NewColorRGB(0.83, 0.35, 0.23)
			return cheeseColor.Scale(cheeseFrac).Add(sauceColor.Scale(1 - cheeseFrac))
		} else {
			// Crust of the pizza.
			return render3d.NewColorRGB(0.85*0.9, 0.65*0.9, 0.2*0.9)
		}
	}
}

func RandCheeseDensity() func(c model2d.Coord) float64 {
	cheeseDots := []model2d.Coord{}
	for i := 0; i < 800; i++ {
		cheeseDots = append(cheeseDots, model2d.NewCoordRandBounds(model2d.XY(-4, -4), model2d.XY(4, 4)))
	}
	cheeseTree := model2d.NewCoordTree(cheeseDots)

	return func(c model2d.Coord) float64 {
		cheeseDist := cheeseTree.NearestNeighbor(c).SquaredDist(c)
		return math.Exp(-cheeseDist / 0.02)
	}
}
