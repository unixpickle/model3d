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
	potColorMesh := ColorFuncMesh(pot)

	log.Println("Creating arrangement...")
	potColor := toolbox3d.ConstantCoordColorFunc(render3d.NewColorRGB(0.85, 0.46, 0.24))
	flowers, colorFuncs := CompileArrangement(CreateArrangement())
	colorFuncs[potColorMesh] = potColor

	log.Println("Creating color func...")
	colorFunc := toolbox3d.JoinedMeshCoordColorFunc(colorFuncs).Cached()

	log.Println("Creating mesh...")
	solid := (model3d.JoinedSolid{
		pot,
		flowers,
	}).Optimize()
	mesh := model3d.MarchingCubesSearch(solid, 0.02, 8)

	log.Println("Decimating mesh...")
	dec := &model3d.Decimator{
		FeatureAngle:     0.03,
		BoundaryDistance: 1e-5,
		PlaneDistance:    1e-4,
		FilterFunc: func(c model3d.Coord3D) bool {
			cColor := colorFunc(c)
			for _, t := range mesh.Find(c) {
				for _, p := range t {
					if p != c && colorFunc(p) != cColor {
						return false
					}
				}
			}
			return true
		},
	}
	preCount := len(mesh.TriangleSlice())
	mesh = dec.Decimate(mesh)
	postCount := len(mesh.TriangleSlice())
	log.Printf("Went from %d to %d triangles", preCount, postCount)

	log.Println("Rendering...")
	render3d.SaveRendering("rendering.png", mesh, model3d.YZ(8.0, 5.5), 1024, 1024,
		colorFunc.RenderColor)

	log.Println("Saving...")
	mesh.SaveMaterialOBJ("arrangement.zip", colorFunc.TriangleColor)
	log.Println("Done.")
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
			Base:   model2d.XY(math.Cos(theta), math.Sin(theta)).Scale(1.2),
			Height: 1.3,
			Flower: flowerBB,
		})
		results = append(results, &FlowerDesc{
			Base:   model2d.XY(math.Cos(theta+spacing), math.Sin(theta+spacing)),
			Height: 1.5,
			Flower: flowerRose,
		})
		results = append(results, &FlowerDesc{
			Base:   model2d.XY(math.Cos(theta), math.Sin(theta)).Scale(0.7),
			Height: 2.6,
			Flower: flowerPurple,
		})
	}
	return results
}

func CompileArrangement(fs []*FlowerDesc) (model3d.Solid, map[*model3d.Mesh]any) {
	solids := model3d.JoinedSolid{}
	stemColor := render3d.NewColorRGB(55.0/255, 102.0/255, 54.0/255)
	colorFuncs := map[*model3d.Mesh]any{}
	for i, f := range fs {
		log.Printf("Compiling flower %d/%d", i+1, len(fs))
		stem := NewStem(model3d.XY(f.Base.X, f.Base.Y), f.Height, f.Flower.Tilt)
		solids = append(solids, stem.Solid)
		colorFuncs[ColorFuncMesh(stem.Solid)] = stemColor
		flower := f.Flower.Place(stem.Tip)
		solids = append(solids, flower.Solid)
		colorFuncs[flower.ColorMesh] = flower.ColorFunc
	}
	return solids.Optimize(), colorFuncs
}

func ColorFuncMesh(s model3d.Solid) *model3d.Mesh {
	return model3d.MarchingCubesSearch(s, 0.025, 8)
}
