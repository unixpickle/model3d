package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d"
)

const NumSides = 12

func main() {
	log.Println("Creating diamond polytope...")
	system := TriangularDiamondPolytope()
	log.Println("Exporting diamond...")
	mesh := system.Mesh()
	mesh.SaveGroupedSTL("diamond.stl")
	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)

	CreateStand(mesh)
}

func CreateStand(diamond *model3d.Mesh) {
	log.Println("Creating stand...")
	diamond = diamond.MapCoords(func(c model3d.Coord3D) model3d.Coord3D {
		c.Z *= -1
		c.X *= -1
		return c
	})
	solid := model3d.NewColliderSolid(model3d.MeshToCollider(diamond))

	standSolid := &model3d.SubtractedSolid{
		Positive: &model3d.CylinderSolid{
			P1:     model3d.Coord3D{Z: solid.Min().Z},
			P2:     model3d.Coord3D{Z: solid.Min().Z + 0.5},
			Radius: 1.0,
		},
		Negative: solid,
	}
	mesh := model3d.SolidToMesh(standSolid, 0.01, 0, 0, 0)
	smoother := &model3d.MeshSmoother{
		StepSize:           0.1,
		Iterations:         200,
		ConstraintDistance: 0.01,
		ConstraintWeight:   0.1,
	}
	mesh = smoother.Smooth(mesh)
	mesh = mesh.FlattenBase(0)

	mesh.SaveGroupedSTL("stand.stl")
	model3d.SaveRandomGrid("rendering_stand.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)
}

// TriangularDiamondPolytope creates a polytope for a
// diamond with a triangular cut base.
func TriangularDiamondPolytope() model3d.ConvexPolytope {
	z1 := 0.5
	z2 := 1.0
	for {
		secZ := (z1 + z2) / 2
		secDelta := OptimalPolytopeDelta(secZ)
		cp := CreatePolytope(secZ, secDelta)
		if math.Abs(z2-z1) < 1e-5 {
			return cp
		}
		if PolytopeTriangleScore(cp) < 1e-5 {
			z2 = secZ
		} else {
			z1 = secZ
		}
	}
}

// OptimalPolytopeDelta computes the secondaryDelta that
// minimizes the area difference between the two equations
// that govern the base of the diamond.
func OptimalPolytopeDelta(secZ float64) float64 {
	areaDelta := func(secDelta float64) float64 {
		cp := CreatePolytope(secZ, secDelta)
		m := cp.Mesh()
		norm1 := cp[1].Normal
		norm2 := cp[2].Normal
		var area1, area2 float64
		m.Iterate(func(t *model3d.Triangle) {
			if t.Area() < 1e-5 {
				return
			}
			if t.Normal().Dot(norm1) > 1-1e-5 {
				area1 += t.Area()
			} else if t.Normal().Dot(norm2) > 1-1e-5 {
				area2 += t.Area()
			}
		})
		return area1 - area2
	}
	d1 := -0.2
	d2 := 0.2
	for i := 0; i < 16; i++ {
		d := (d1 + d2) / 2
		score := areaDelta(d)
		if score < 0 {
			d1 = d
		} else {
			d2 = d
		}
	}
	return (d1 + d2) / 2
}

// PolytopeTriangleScore gives a number measuring how
// close the bottoms of the diamond's two base equations
// are. Minimum value is zero, at which point the two
// equations meet exactly.
func PolytopeTriangleScore(cp model3d.ConvexPolytope) float64 {
	norm1 := cp[1].Normal
	norm2 := cp[2].Normal

	var min1, min2 model3d.Coord3D
	cp.Mesh().Iterate(func(t *model3d.Triangle) {
		if t.Area() < 1e-5 {
			return
		}
		if t.Normal().Dot(norm1) > 1-1e-5 {
			min1 = min1.Min(t.Min())
		} else if t.Normal().Dot(norm2) > 1-1e-5 {
			min2 = min2.Min(t.Min())
		}
	})
	return math.Abs(min1.Z - min2.Z)
}

// CreatePolytope creates a diamond polytope with two
// parameters that govern the shape of the base.
// These parameters are intended to be tuned for a
// perfectly triangular cut on the base.
func CreatePolytope(secondaryZ, secondaryDelta float64) model3d.ConvexPolytope {
	system := model3d.ConvexPolytope{
		&model3d.LinearConstraint{
			Normal: model3d.Coord3D{Z: -1},
			Max:    0.4,
		},
	}
	iAngle := math.Pi * 2 / NumSides
	for i := 0; i < NumSides; i++ {
		theta := float64(i) * iAngle
		p1 := model3d.Coord3D{X: math.Cos(theta), Y: math.Sin(theta)}
		p2 := model3d.Coord3D{X: math.Cos(theta + iAngle/2), Y: math.Sin(theta + iAngle/2)}
		n1 := model3d.Coord3D{X: p1.X, Y: p1.Y, Z: -secondaryZ}.Normalize()
		n2 := model3d.Coord3D{X: p2.X, Y: p2.Y, Z: -1}.Normalize()
		n3 := model3d.Coord3D{X: p1.X, Y: p1.Y, Z: 0.8}.Normalize()
		system = append(system,
			&model3d.LinearConstraint{
				Normal: n1,
				Max:    n1.Dot(p1),
			},
			&model3d.LinearConstraint{
				Normal: n2,
				Max:    n2.Dot(p2) + secondaryDelta,
			},
			&model3d.LinearConstraint{
				Normal: n3,
				Max:    n3.Dot(p1),
			},
		)
	}
	return system
}
