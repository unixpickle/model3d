package main

import (
	"flag"
	"log"
	"math"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/render3d"

	"github.com/unixpickle/model3d/model3d"
)

func main() {
	var totalLength float64
	var linkLength float64
	var linkThickness float64
	var holderDiameter float64
	var holderBarLength float64
	var holderBarThickness float64
	var attachmentName string

	var spiralRadius float64
	var resolution float64

	flag.Float64Var(&totalLength, "total-length", 18.0, "total length of necklace")
	flag.Float64Var(&linkLength, "link-length", 0.2, "inner diameter of each link")
	flag.Float64Var(&linkThickness, "link-thickness", 0.02, "thickness of links")
	flag.Float64Var(&holderDiameter, "holder-diameter", 0.4, "diameter of the holder link")
	flag.Float64Var(&holderBarLength, "holder-bar-length", 0.9, "length of holder bar")
	flag.Float64Var(&holderBarThickness, "holder-bar-thickness", 0.04, "diameter of holder bar")
	flag.StringVar(&attachmentName, "attachment", "heart", "name of attachment to use")

	flag.Float64Var(&spiralRadius, "spiral-radius", 1.0, "radius of spiral for model layout")
	flag.Float64Var(&resolution, "resolution", 0.005, "resolution for marching cubes")

	flag.Parse()

	log.Println("Calculating spiral...")
	numLinks := int(math.Ceil(totalLength / linkLength))
	spiral := createSpiralCenters(numLinks, linkLength, linkThickness, spiralRadius)

	log.Println("Creating solids...")

	// Create the larger link for the holder.
	holderSpace := holderDiameter/2 - linkThickness/2 + linkLength/3
	holderDirection := spiral[0].Sub(spiral[1]).Normalize()
	spiral = append([]model3d.Coord3D{spiral[0].Add(holderDirection.Scale(holderSpace))}, spiral...)

	centerLink := len(spiral) / 2
	if centerLink%2 == 1 {
		centerLink++
	}

	// A collection of non-intersecting solids.
	// These solids may interlink, but not collide, so
	// that each one can be meshed separately.
	solids := make([]model3d.Solid, len(spiral))
	for i, center := range spiral {
		axis := model3d.Z(1)
		if i%2 == 1 {
			axis = center.Mul(model3d.XY(1, 1)).Normalize()
		}
		radius := linkLength / 2
		if i == 0 {
			radius = holderDiameter / 2
		}
		if i == centerLink {
			centerDir := center.Mul(model3d.XY(1, 1)).Normalize()
			// Extra offset (linkThickness/2 => linkThickness) to prevent
			// collisions between this link and the others.
			centerOffset := (2.0/3.0)*linkLength + linkThickness
			axis1 := model3d.XY(center.Y, -center.X).Normalize()
			attachmentLink := &model3d.Torus{
				Center:      center.Add(centerDir.Scale(centerOffset)),
				Axis:        axis1,
				InnerRadius: linkThickness / 2,
				OuterRadius: radius + linkThickness/2,
			}
			// Fuse attachment to attachment link
			solids = append(solids, model3d.JoinedSolid{
				attachmentLink,
				createAttachment(attachmentLink, attachmentName),
			})
		}
		solids[i] = &model3d.Torus{
			Center:      center,
			Axis:        axis,
			InnerRadius: linkThickness / 2,
			OuterRadius: radius + linkThickness/2,
		}
	}

	// Fuse holder bar to last link.
	holderBar := createHolderBar(spiral, linkLength, holderBarLength, holderBarThickness)
	solids[len(spiral)-1] = append(holderBar, solids[len(spiral)-1])

	log.Println("Creating mesh...")
	mesh := combinedSolidMeshes(solids, resolution)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("necklace.stl")

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func createSpiralCenters(numLinks int, linkLength, linkThickness,
	spiralRadius float64) []model3d.Coord3D {
	// Each cycle around creates 2x the link length of space
	// between links, which should be more than enough.
	heightPerRadian := (linkLength * 2) / (2 * math.Pi)
	distPerLink := (2.0/3.0)*linkLength + linkThickness/2

	var thetaStep float64
	for ts := 0.0; true; ts += 0.00001 {
		p1 := model3d.XY(math.Cos(0), math.Sin(0)).Scale(spiralRadius)
		p2 := model3d.XYZ(math.Cos(ts), math.Sin(ts), ts*heightPerRadian).Scale(spiralRadius)
		if p1.Dist(p2) >= distPerLink {
			thetaStep = ts
			break
		}
	}

	points := make([]model3d.Coord3D, numLinks)
	for i := range points {
		theta := thetaStep * float64(i)
		height := heightPerRadian * theta
		points[i] = model3d.XYZ(math.Cos(theta), math.Sin(theta), height).Scale(spiralRadius)
	}
	return points
}

func createAttachment(link *model3d.Torus, name string) model3d.Solid {
	rawAttachment := Attachment(name)
	xOffset := -rawAttachment.Min().X
	for !rawAttachment.Contains(model3d.X(-xOffset)) {
		xOffset -= 0.001
	}
	tipOffset := link.OuterRadius - link.InnerRadius
	tipDirection := model3d.XY(link.Center.X, link.Center.Y).Normalize()
	linkTip := link.Center.Add(tipDirection.Scale(tipOffset))
	angle := math.Atan2(link.Center.Y, link.Center.X)

	return model3d.TransformSolid(model3d.JoinedTransform{
		&model3d.Translate{Offset: model3d.X(xOffset)},
		model3d.Rotation(model3d.Z(1), angle),
		&model3d.Translate{Offset: linkTip},
	}, rawAttachment)
}

func createHolderBar(spiral []model3d.Coord3D, linkLength, length,
	thickness float64) model3d.JoinedSolid {
	holderBarAxis := spiral[len(spiral)-1].Mul(model3d.XY(1, 1)).Normalize()
	holderBarDir := spiral[len(spiral)-1].Sub(spiral[len(spiral)-2]).Normalize()
	holderBarCenter := spiral[len(spiral)-1].Add(holderBarDir.Scale(linkLength/2 + thickness/2))
	holderBarCylinder := &model3d.Cylinder{
		P1:     holderBarCenter.Sub(holderBarAxis.Scale(length / 2)),
		P2:     holderBarCenter.Add(holderBarAxis.Scale(length / 2)),
		Radius: thickness / 2,
	}
	holderBar := model3d.JoinedSolid{holderBarCylinder}

	// Rounded tips.
	for _, p := range []model3d.Coord3D{holderBarCylinder.P1, holderBarCylinder.P2} {
		holderBar = append(holderBar, &model3d.Sphere{
			Center: p,
			Radius: holderBarCylinder.Radius,
		})
	}

	return holderBar
}

func combinedSolidMeshes(solids []model3d.Solid, resolution float64) *model3d.Mesh {
	mesh := model3d.NewMesh()
	for _, solid := range solids {
		if torus, ok := solid.(*model3d.Torus); ok {
			mesh.AddMesh(torusToMesh(torus, resolution))
		} else {
			mesh.AddMesh(model3d.MarchingCubesSearch(solid, resolution, 8))
		}
	}
	return mesh
}

func torusToMesh(torus *model3d.Torus, resolution float64) *model3d.Mesh {
	outmostRadius := torus.InnerRadius + torus.OuterRadius
	outmostCircum := 2 * math.Pi * outmostRadius
	outerSteps := essentials.MaxInt(5, int(math.Ceil(outmostCircum/resolution)))
	innerCircum := 2 * math.Pi * torus.InnerRadius
	innerSteps := essentials.MaxInt(5, int(math.Ceil(innerCircum/resolution)))

	innerTheta := func(i int) float64 {
		return float64(i%innerSteps) * 2 * math.Pi / float64(innerSteps)
	}
	outerTheta := func(i int) float64 {
		return float64(i%outerSteps) * 2 * math.Pi / float64(outerSteps)
	}

	axis := torus.Axis.Normalize()
	basis1, basis2 := axis.Normalize().OrthoBasis()
	coord := func(outer, inner int) model3d.Coord3D {
		outerAngle := outerTheta(outer)
		outerVec := basis1.Scale(math.Cos(outerAngle)).Add(basis2.Scale(math.Sin(outerAngle)))
		innerAngle := innerTheta(inner)
		innerVec := outerVec.Scale(math.Cos(innerAngle)).Add(axis.Scale(math.Sin(innerAngle)))
		innerCenter := torus.Center.Add(outerVec.Scale(torus.OuterRadius))
		return innerCenter.Add(innerVec.Scale(torus.InnerRadius))
	}

	mesh := model3d.NewMesh()
	for i := 0; i < outerSteps; i++ {
		for j := 0; j < innerSteps; j++ {
			mesh.AddQuad(coord(i, j), coord(i, j+1), coord(i+1, j+1), coord(i+1, j))
		}
	}
	return mesh
}
