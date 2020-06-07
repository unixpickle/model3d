package main

import (
	"io/ioutil"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const AppleHeight = 1.0

func main() {
	bite := &model3d.Torus{
		Center:      model3d.Coord3D{X: -1.7, Y: 0, Z: AppleHeight / 2},
		Axis:        model3d.Z(1),
		OuterRadius: 1.0,
		InnerRadius: 0.3,
	}
	biggerBite := *bite
	biggerBite.InnerRadius += 0.005

	stem := &model3d.Cylinder{
		P1:     model3d.Z(AppleHeight / 2),
		P2:     model3d.Z(AppleHeight * 1.1),
		Radius: AppleHeight / 30,
	}
	biggerStem := *stem
	biggerStem.Radius += 0.005
	biggerStem.P2.Z += 0.005

	solid := &model3d.SubtractedSolid{
		Positive: model3d.JoinedSolid{
			NewAppleSolid(),
			stem,
		},
		Negative: bite,
	}

	mesh := model3d.MarchingCubesSearch(solid, 0.005, 8)

	sig := NewSignature()
	vertexColor := func(c model3d.Coord3D) [3]float64 {
		if biggerBite.Contains(c) {
			if sig.Contains(c) {
				return [3]float64{0, 0, 0}
			}
			return [3]float64{1, 1, 0.5}
		} else if biggerStem.Contains(c) {
			return [3]float64{0.27, 0.21, 0}
		} else {
			return [3]float64{1, 0, 0}
		}
	}
	colorFunc := model3d.VertexColorsToTriangle(vertexColor)

	ioutil.WriteFile("apple.zip", mesh.EncodeMaterialOBJ(colorFunc), 0755)
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 200,
		render3d.TriangleColorFunc(colorFunc))
}

type AppleSolid struct {
	Image model2d.Solid
}

func NewAppleSolid() *AppleSolid {
	// Read and smooth the apple image to reduce
	// artifacts from using pixels.
	mesh := model2d.MustReadBitmap("half_apple.png", nil).FlipY().Mesh().SmoothSq(200)
	collider := model2d.MeshToCollider(mesh)
	solid := model2d.NewColliderSolid(collider)
	return &AppleSolid{Image: model2d.ScaleSolid(solid, AppleHeight/solid.Max().Y)}
}

func (a *AppleSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -a.width(), Y: -a.width(), Z: 0}
}

func (a *AppleSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: a.width(), Y: a.width(), Z: AppleHeight}
}

func (a *AppleSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(a, c) {
		return false
	}
	radius := c.XY().Norm()
	if radius < 1e-5 {
		// Prevent queries directly on the boundary.
		radius = 1e-5
	}
	imageX := a.width() - radius
	imageY := c.Z

	return a.Image.Contains(model2d.Coord{X: imageX, Y: imageY})
}

func (a *AppleSolid) width() float64 {
	return a.Image.Max().X
}

type Signature struct {
	Image *model2d.Bitmap
}

func NewSignature() *Signature {
	return &Signature{Image: model2d.MustReadBitmap("turing_signature.png", nil)}
}

func (s *Signature) Contains(c model3d.Coord3D) bool {
	if c.X > 0 {
		return false
	}
	scale := float64(s.Image.Width) / 0.5
	imageX := s.Image.Width/2 - int(math.Round(c.Y*scale))
	imageY := s.Image.Height/2 - int(math.Round((c.Z-AppleHeight/2-0.05)*scale))
	return s.Image.Get(imageX, imageY)
}
