package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func PenguinBody() (obj model3d.Solid, colors []interface{}) {
	torso := PenguinTorso()
	arms := PenguinArms()
	eyes := PenguinEyes()
	beak := PenguinBeak()
	feet := PenguinFeet()

	// We color the inside of the penguin body with white.
	// If we use a uniform inset, then there is not enough
	// white in the face, so we reduce the inset as we get
	// closer to the head.
	profile := PenguinProfile()
	profileCollider := model2d.MeshToCollider(profile)
	profileSolid := model2d.NewColliderSolid(profileCollider)
	xf := &model3d.Scale{Scale: 2.0}
	torso = model3d.TransformSolid(xf, torso)
	arms = model3d.TransformSolid(xf, arms)
	eyes = model3d.TransformSolid(xf, eyes)
	beak = model3d.TransformSolid(xf, beak)
	feet = model3d.TransformSolid(xf, feet)
	fullSolid := (model3d.JoinedSolid{torso, arms, eyes, beak, feet}).Optimize()
	return fullSolid, []interface{}{
		torso, toolbox3d.CoordColorFunc(func(c model3d.Coord3D) render3d.Color {
			inset := 0.2 - 0.1*c.Z/2.1
			if c.Y > 0 || !profileSolid.Contains(c.XZ()) || profileCollider.CircleCollision(c.XZ(), inset) {
				return render3d.NewColor(0.1)
			} else {
				return render3d.NewColor(1.0)
			}
		}).Transform(xf),
		eyes, toolbox3d.ConstantCoordColorFunc(render3d.NewColor(0.0)),
		arms, toolbox3d.ConstantCoordColorFunc(render3d.NewColor(0.1)),
		beak, toolbox3d.ConstantCoordColorFunc(render3d.NewColorRGB(1.0, 0.5, 0.0)),
		feet, toolbox3d.ConstantCoordColorFunc(render3d.NewColorRGB(1.0, 0.5, 0.0)),
	}
}

func PenguinTorso() model3d.Solid {
	profile := PenguinProfile()
	shape := model3d.RevolveSolid(model2d.NewColliderSolid(model2d.MeshToCollider(profile)),
		model3d.Z(1))

	// Squeeze to make it less radially symmetric.
	depthMapping := model2d.BezierCurve{
		model2d.XY(0, 0),
		model2d.XY(0.5, 0.5),
		model2d.XY(0.7, 1.0),
	}
	return model3d.CheckedFuncSolid(shape.Min(), shape.Max(), func(c model3d.Coord3D) bool {
		newY := depthMapping.EvalX(math.Abs(c.Y))
		if c.Y < 0 {
			newY *= -1
		}
		return shape.Contains(model3d.XYZ(c.X, newY, c.Z))
	})
}

func PenguinProfile() *model2d.Mesh {
	points := [][2]float64{
		{0, -0.1},
		{0.2, -0.05},
		{0.3, 0.1},
		{0.3, 0.4},
		{0.25, 0.6},
		{0.5, 0.95},
		{0.6, 1.2},
		{0.7, 1.6},
		{0.5, 2.1},
		{0.0, 2.1},
	}
	var segs []*model2d.Segment
	for i := 1; i < len(points); i++ {
		segs = append(segs, &model2d.Segment{
			model2d.NewCoordArray(points[i-1]),
			model2d.NewCoordArray(points[i]),
		})
	}
	roughOutline := model2d.NewMeshSegments(segs)
	roughOutline.AddMesh(roughOutline.MapCoords(model2d.XY(-1, 1).Mul))
	roughOutline, _ = roughOutline.Scale(-1).RepairNormals(1e-5)
	return roughOutline.Subdivide(5).SmoothSq(5).Translate(model2d.Y(2.1))
}

func PenguinArms() model3d.Solid {
	v := model3d.XYZ(0.2, 0, 0.7).Normalize()
	segmentFn := func(t float64) [2]model3d.Coord3D {
		x := -0.58 + 0.3*(1.5*t*t*t-0.6*t)
		y := 0.0 - t*0.5
		z := 0.8 - 0.4*t*t
		radius := 0.25 - 0.25*t
		center := model3d.XYZ(x, y, z)
		return [2]model3d.Coord3D{
			center.Add(v.Scale(radius)),
			center.Sub(v.Scale(radius)),
		}
	}

	baseMesh := model3d.NewMesh()
	delta := 0.01
	for t := 0.0; t+delta <= 1.0; t += delta {
		s1 := segmentFn(t)
		s2 := segmentFn(t + delta)
		baseMesh.AddQuad(s1[0], s1[1], s2[1], s2[0])
	}

	baseSolid := model3d.NewColliderSolidHollow(model3d.MeshToCollider(baseMesh), 0.08)
	flippedMesh := baseMesh.MapCoords(model3d.XYZ(-1, 1, 1).Mul)
	mirrored := model3d.NewColliderSolidHollow(model3d.MeshToCollider(flippedMesh), 0.08)

	return model3d.JoinedSolid{baseSolid, mirrored}
}

func PenguinEyes() model3d.Solid {
	return model3d.JoinedSolid{
		&model3d.Sphere{
			Center: model3d.XYZ(-0.068, -0.22, 1.9),
			Radius: 0.06,
		},
		&model3d.Sphere{
			Center: model3d.XYZ(0.068, -0.22, 1.9),
			Radius: 0.06,
		},
	}
}

func PenguinBeak() model3d.Solid {
	beekProfile := model2d.BezierCurve{
		model2d.XY(-0.15, 0.0),
		model2d.XY(-0.1, 0.1),
		model2d.XY(0.1, 0.1),
		model2d.XY(0.15, 0.0),
	}
	return model3d.TranslateSolid(
		model3d.CheckedFuncSolid(
			model3d.XYZ(-0.15, -0.34, -0.1),
			model3d.XYZ(0.15, -0.24, 0.1),
			func(c model3d.Coord3D) bool {
				radiusScale := 1 / math.Max(1e-5, (0.34+c.Y)/0.1)
				tx := c.X * radiusScale
				if tx < -0.15 || tx > 0.15 {
					return false
				}
				z := beekProfile.EvalX(tx) / radiusScale
				return math.Abs(c.Z) < z
			},
		),
		model3d.Z(1.7),
	)
}

func PenguinFeet() model3d.Solid {
	rad := 0.075
	dir := model3d.XYZ(-0.3, 0, 0.3)
	dir1 := model3d.Rotation(model3d.Y(1), 0.5).Apply(dir)
	dir2 := model3d.Rotation(model3d.Y(1), -0.5).Apply(dir)
	origin := model3d.XYZ(-0.35, -0.5, 0.05)
	segs := []model3d.Segment{
		model3d.NewSegment(origin, origin.Add(dir)),
		model3d.NewSegment(origin, origin.Add(dir1)),
		model3d.NewSegment(origin, origin.Add(dir2)),
		model3d.NewSegment(origin, origin.Add(model3d.Y(0.3))),
	}

	min, max := segs[0][0], segs[0][0]
	for _, s := range segs {
		for _, p := range s {
			min = min.Min(p)
			max = max.Max(p)
		}
	}

	min = min.AddScalar(-rad * 2)
	max = max.AddScalar(rad * 2)
	min.Z = 0.0
	foot1 := model3d.CheckedFuncSolid(min, max, func(c model3d.Coord3D) bool {
		var sum float64
		for _, s := range segs {
			x := s.Dist(c)
			x = 1 / (x * x)
			x = x * x
			sum += x
		}
		return sum > (1 / (rad * rad * rad * rad))
	})
	foot2 := model3d.TransformSolid(&model3d.Matrix3Transform{
		Matrix: &model3d.Matrix3{-1.0, 0.0, 0.0, 0.0, 1.0, 0.0, 0.0, 0.0, 1.0},
	}, foot1)
	return (model3d.JoinedSolid{foot1, foot2}).Optimize()
}
