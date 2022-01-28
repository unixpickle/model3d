package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func main() {
	log.Println("Creating flower pot...")
	pot := FlowerPot()
	potSDF := ColorFuncSDF(pot)

	log.Println("Creating arrangement...")
	potColor := toolbox3d.ConstantCoordColorFunc(render3d.NewColorRGB(0.85, 0.46, 0.24))
	flowers, colorFunc := CompileArrangement(CreateArrangement(), potSDF, potColor)

	log.Println("Creating mesh...")
	solid := (model3d.JoinedSolid{
		pot,
		flowers,
	}).Optimize()
	mesh := model3d.MarchingCubesSearch(solid, 0.02, 8)

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 512, colorFunc.RenderColor)

	log.Println("Saving...")
	mesh.SaveGroupedSTL("arrangement.stl")
}

type FlowerDesc struct {
	Base   model2d.Coord
	Height float64
	Flower *Flower
}

func CreateArrangement() []*FlowerDesc {
	log.Println("Creating bermuda buttercup...")
	flowerBB := NewBermudaButtercup()
	log.Println("Creating rose...")
	flowerRose := NewRose()
	log.Println("Creating purple...")
	flowerPurple := NewPurpleRowFlower()
	log.Println("Done creating flowers.")
	var results []*FlowerDesc
	for i := 0; i < 3; i++ {
		theta := float64(i) * 2 * math.Pi / 3
		spacing := 2 * math.Pi / 6
		results = append(results, &FlowerDesc{
			Base:   model2d.XY(math.Cos(theta), math.Sin(theta)),
			Height: 1.3,
			Flower: flowerBB,
		})
		results = append(results, &FlowerDesc{
			Base:   model2d.XY(math.Cos(theta+spacing), math.Sin(theta+spacing)),
			Height: 1.7,
			Flower: flowerRose,
		})
		results = append(results, &FlowerDesc{
			Base:   model2d.XY(math.Cos(theta), math.Sin(theta)),
			Height: 2.2,
			Flower: flowerPurple,
		})
	}
	return results
}

func CompileArrangement(fs []*FlowerDesc, colorFuncArgs ...interface{}) (model3d.Solid,
	toolbox3d.CoordColorFunc) {
	solids := model3d.JoinedSolid{}
	stemColor := toolbox3d.ConstantCoordColorFunc(render3d.NewColorRGB(0, 0.8, 0))
	for i, f := range fs {
		log.Printf("Compiling flower %d/%d", i+1, len(fs))
		stem := NewStem(model3d.XY(f.Base.X, f.Base.Y), f.Height, f.Flower.Tilt)
		solids = append(solids, stem.Solid)
		colorFuncArgs = append(colorFuncArgs, ColorFuncSDF(stem.Solid), stemColor)
		flower := f.Flower.Place(stem.Tip)
		solids = append(solids, flower.Solid)
		colorFuncArgs = append(colorFuncArgs, flower.ColorSDF, flower.ColorFunc)
	}
	return solids.Optimize(), toolbox3d.JoinedCoordColorFunc(colorFuncArgs...).Cached()
}

func ColorFuncSDF(s model3d.Solid) model3d.SDF {
	return model3d.MeshToSDF(model3d.MarchingCubesSearch(s, 0.025, 8))
}
