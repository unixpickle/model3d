package main

import (
	"flag"
	"math"

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

	numLinks := int(math.Ceil(totalLength / linkLength))
	spiral := createSpiralCenters(numLinks, linkLength, linkThickness, spiralRadius)

	// Create the larger link for the holder.
	holderSpace := holderDiameter/2 + linkThickness/2 - linkLength/3
	holderDirection := spiral[0].Sub(spiral[1]).Normalize()
	spiral = append([]model3d.Coord3D{spiral[0].Add(holderDirection.Scale(holderSpace))}, spiral...)

	centerLink := len(spiral) / 2
	if centerLink%2 == 1 {
		centerLink++
	}

	links := make(model3d.JoinedSolid, len(spiral))
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
			links = append(links, attachmentLink, createAttachment(attachmentLink, attachmentName))
		}
		links[i] = &model3d.Torus{
			Center:      center,
			Axis:        axis,
			InnerRadius: linkThickness / 2,
			OuterRadius: radius + linkThickness/2,
		}
	}

	// Create the holder bar.
	holderBarAxis := spiral[len(spiral)-1].Mul(model3d.XY(1, 1)).Normalize()
	holderBarDir := spiral[len(spiral)-1].Sub(spiral[len(spiral)-2]).Normalize()
	holderBarCenter := spiral[len(spiral)-1].Add(holderBarDir.Scale(linkLength/2 + holderBarThickness/2))
	holderBar := &model3d.Cylinder{
		P1:     holderBarCenter.Sub(holderBarAxis.Scale(holderBarLength / 2)),
		P2:     holderBarCenter.Add(holderBarAxis.Scale(holderBarLength / 2)),
		Radius: holderBarThickness / 2,
	}
	links = append(links, holderBar)
	// Rounded tips for holder bar.
	for _, p := range []model3d.Coord3D{holderBar.P1, holderBar.P2} {
		links = append(links, &model3d.Sphere{
			Center: p,
			Radius: holderBar.Radius,
		})
	}

	fastLinks := links.Optimize()

	mesh := model3d.MarchingCubesSearch(fastLinks, resolution, 8)
	mesh.SaveGroupedSTL("necklace.stl")
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
		&model3d.Matrix3Transform{Matrix: model3d.NewMatrix3Rotation(model3d.Z(1), angle)},
		&model3d.Translate{Offset: linkTip},
	}, rawAttachment)
}
