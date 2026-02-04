package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	. "github.com/unixpickle/model3d/shorthand"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	FlowerLayerThickness   = 1.0
	FlowerLayerRadius      = 2.2
	FlowerLayerFlowerDepth = 0.03
)

func FlowerLayer() (Solid3, toolbox3d.CoordColorFunc) {
	numFlowers := math.Floor(2 * math.Pi * FlowerLayerRadius / FlowerLayerThickness)
	flowerTheta := math.Pi * 2 / numFlowers
	flowerMesh := model3d.NewMesh()

	for i := 0.0; i < numFlowers; i++ {
		offset := flowerTheta * i
		for r := 0.05; r < flowerTheta/2-0.03; r += 0.01 {
			delta := 0.05
			for theta := 0.0; theta < 2*math.Pi; theta += delta {
				t1 := theta
				t2 := math.Min(theta+delta, 2*math.Pi)
				r1 := (math.Sin(r*1000+t1*10) + 10) / 10 * r
				r2 := (math.Sin(r*1000+t2*10) + 10) / 10 * r
				x1 := offset + flowerTheta/2 + r1*math.Cos(t1)
				y1 := FlowerLayerRadius * (flowerTheta/2 + r1*math.Sin(t1))
				x2 := offset + flowerTheta/2 + r2*math.Cos(t2)
				y2 := FlowerLayerRadius * (flowerTheta/2 + r2*math.Sin(t2))

				p1 := XYZ(math.Cos(x1)*FlowerLayerRadius, math.Sin(x1)*FlowerLayerRadius, y1)
				p2 := XYZ(math.Cos(x2)*FlowerLayerRadius, math.Sin(x2)*FlowerLayerRadius, y2)
				flowerMesh.Add(&model3d.Triangle{p1, p2, p2})
			}
		}
	}

	solid := Join3(
		Cylinder(
			Origin3,
			Z(FlowerLayerThickness),
			FlowerLayerRadius,
		),
		model3d.NewColliderSolidHollow(model3d.MeshToCollider(flowerMesh), FlowerLayerFlowerDepth),
	)
	return solid, toolbox3d.ConstantCoordColorFunc(Gray(1.0))
}
