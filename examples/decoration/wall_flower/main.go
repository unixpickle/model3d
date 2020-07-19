package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model2d"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	// Radius of the sphere we project everything onto.
	Radius = 1.0

	// Radius spanned by the flower itself.
	PedalLength = 0.8

	// Thickness in every direction of the flower surface.
	Thickness = 0.1

	// Details of the cylindrical base.
	BaseThickness = 0.1
	BaseRadius    = 0.5
)

func main() {
	log.Println("Creating solid...")
	solid := model3d.JoinedSolid{
		NewFlowerShape(),
		BaseSolid{},
	}

	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.01, 8)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("flower.stl")

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

type FlowerShape struct {
	Projection model2d.PointSDF
	MinVal     model3d.Coord3D
	MaxVal     model3d.Coord3D
}

func NewFlowerShape() *FlowerShape {
	mesh := model2d.NewMesh()

	// Polar coordinate flower shape.
	flowerPoint := func(i int) model2d.Coord {
		i = i % 10000
		theta := (float64(i) / 9999.0) * math.Pi * 2
		r := 0.3 + 0.7*math.Pow(math.Abs(math.Cos(3*theta)), 0.5)
		return model2d.XY(math.Cos(theta), math.Sin(theta)).Scale(r)
	}
	for i := 0; i < 10000; i++ {
		mesh.Add(&model2d.Segment{flowerPoint(i), flowerPoint(i + 1)})
	}

	sphere := &model3d.Sphere{Radius: Radius}
	mesh = mesh.MapCoords(func(c model2d.Coord) model2d.Coord {
		c = c.Scale(PedalLength)

		// Project the point onto a sphere to get the geo coords.
		ray := &model3d.Ray{
			Origin:    model3d.XYZ(c.X, c.Y, Radius),
			Direction: model3d.Z(-1),
		}
		rc, _ := sphere.FirstRayCollision(ray)
		rcPoint := ray.Origin.Add(ray.Direction.Scale(rc.Scale))
		geoPoint := rcPoint.Geo()

		return model2d.XY(geoPoint.Lat, geoPoint.Lon)
	})

	return &FlowerShape{
		Projection: model2d.MeshToSDF(mesh),
		// Fairly loose bounds, since the exact bounds are
		// hard to compute.
		MinVal: model3d.XYZ(-Radius, -Radius, 0),
		MaxVal: model3d.XYZ(Radius, Radius, Radius),
	}
}

func (f *FlowerShape) Min() model3d.Coord3D {
	return f.MinVal
}

func (f *FlowerShape) Max() model3d.Coord3D {
	return f.MaxVal
}

func (f *FlowerShape) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(f, c) {
		return false
	}
	c.Z -= Radius
	c.Z *= -1
	proj := f.project(c)
	return c.Dist(proj) < Thickness
}

func (f *FlowerShape) project(c model3d.Coord3D) model3d.Coord3D {
	geo := c.Geo()
	closest, signedDist := f.Projection.PointSDF(model2d.XY(geo.Lat, geo.Lon))
	if signedDist > 0 {
		// Interior projections simply land on the sphere.
		return geo.Coord3D().Scale(Radius)
	}
	// Exterior projections hit the boundary of the shape.
	// We assume that the closest point in geo coordinates
	// is approximately the closest point in space.
	geo = model3d.GeoCoord{Lat: closest.X, Lon: closest.Y}.Normalize()
	return geo.Coord3D().Scale(Radius)
}

type BaseSolid struct{}

func (b BaseSolid) Min() model3d.Coord3D {
	return model3d.XYZ(-BaseRadius, -BaseRadius, -BaseThickness)
}

func (b BaseSolid) Max() model3d.Coord3D {
	return model3d.XYZ(BaseRadius, BaseRadius, Radius)
}

func (b BaseSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(b, c) {
		return false
	}
	cylinderDist := c.XY().Norm()
	sphereDist := c.Dist(model3d.Z(Radius))
	return cylinderDist < BaseRadius && sphereDist >= Radius
}
