package main

import (
	"math"
	"strings"

	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

const (
	ImageAttachmentCornerRounding = 0.02
	ImageAttachmentThickness      = 0.08
	ImageAttachmentEngravingDepth = 0.01
)

func Attachment(name string) model3d.Solid {
	if name == "sphere" {
		return sphereAttachment()
	} else if name == "heart" {
		return heartAttachment()
	} else {
		mesh, engraving := readAttachmentMeshes(name)
		collider2d := model2d.MeshToCollider(mesh)
		solid2d := model2d.NewColliderSolid(collider2d)
		th := ImageAttachmentThickness/2 - ImageAttachmentCornerRounding
		solid3d := model3d.ProfileSolid(solid2d, -th, th)
		collider3d := model3d.ProfileCollider(collider2d, -th, th)
		baseSolid := model3d.JoinedSolid{
			solid3d,
			model3d.NewColliderSolidHollow(collider3d, ImageAttachmentCornerRounding),
		}
		if engraving != nil {
			engraving2d := model2d.NewColliderSolid(model2d.MeshToCollider(engraving))
			engraving3d := model3d.ProfileSolid(
				engraving2d,
				ImageAttachmentThickness/2-ImageAttachmentEngravingDepth,
				ImageAttachmentThickness/2+1e-5,
			)
			return &model3d.SubtractedSolid{Positive: baseSolid, Negative: engraving3d}
		}
		return baseSolid
	}
}

func readAttachmentMeshes(files string) (mesh *model2d.Mesh, engraving *model2d.Mesh) {
	baseImage, engravingPath := splitBaseAndEngraving(files)
	mesh = model2d.MustReadBitmap(baseImage, nil).Mesh().SmoothSq(50)
	if engravingPath != "" {
		engraving = model2d.MustReadBitmap(engravingPath, nil).Mesh().SmoothSq(50)
	}
	size := mesh.Max().Sub(mesh.Min())
	scale := 0.5 / size.MaxCoord()
	mesh = mesh.Scale(scale)
	if engraving != nil {
		engraving = engraving.Scale(scale)
	}
	mesh = mesh.MapCoords(flipXY)
	if engraving != nil {
		engraving = engraving.MapCoords(flipXY)
	}
	center := mesh.Min().Mid(mesh.Max()).Scale(-1)
	mesh = mesh.Translate(center)
	if engraving != nil {
		engraving = engraving.Translate(center)
	}
	return
}

func flipXY(c model2d.Coord) model2d.Coord {
	return model2d.XY(c.Y, c.X)
}

func splitBaseAndEngraving(files string) (string, string) {
	if strings.Contains(files, ":") {
		parts := strings.Split(files, ":")
		if len(parts) != 2 {
			panic("expected exactly 0 or 1 ':' character")
		}
		return parts[0], parts[1]
	}
	return files, ""
}

func sphereAttachment() model3d.Solid {
	return &model3d.Sphere{Radius: 0.15}
}

func heartAttachment() model3d.Solid {
	shape := model2d.BezierCurve{
		model2d.XY(0, 0),
		model2d.XY(-3, 0),
		model2d.XY(-1, 7),
		model2d.XY(4, 2),
		model2d.XY(4, 0),
	}
	mesh := model2d.NewMesh()
	for t := 0.0; t < 1.0; t += 0.01 {
		nextT := math.Min(1.0, t+0.01)
		mesh.Add(&model2d.Segment{
			shape.Eval(t),
			shape.Eval(nextT),
		})
		mesh.Add(&model2d.Segment{
			shape.Eval(t).Mul(model2d.XY(1, -1)),
			shape.Eval(nextT).Mul(model2d.XY(1, -1)),
		})
	}
	mesh = mesh.Scale(0.4 / 7.0)
	solid2d := model2d.NewColliderSolid(model2d.MeshToCollider(mesh))
	sdf2d := model2d.MeshToSDF(mesh)
	spheres := toolbox3d.NewHeightMap(solid2d.Min(), solid2d.Max(), 1024)
	spheres.AddSpheresSDF(sdf2d, 2000, 1e-5, 0.08)
	return toolbox3d.HeightMapToSolidBidir(spheres)
}
