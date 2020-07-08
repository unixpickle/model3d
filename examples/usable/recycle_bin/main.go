package main

import (
	"flag"
	"image/color"
	"log"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/model3d/model3d"
)

// InsetSlopeFactor controls how sloped the inscription
// is. Lower values mean a sharper slope.
const InsetSlopeFactor = 0.6

func main() {
	var bin BinSolid

	flag.Float64Var(&bin.BaseWidth, "base-width", 6, "width at the bottom")
	flag.Float64Var(&bin.BaseDepth, "base-depth", 5, "depth at the bottom")
	flag.Float64Var(&bin.TopScale, "top-scale", 7.0/6.0, "scale factor from bottom to top")
	flag.Float64Var(&bin.Height, "height", 6.0, "height of the bin")
	flag.Float64Var(&bin.CornerRadius, "corner-radius", 1.0, "corner rounding size")
	flag.Float64Var(&bin.Thickness, "thickness", 0.2, "side thickness")
	flag.Float64Var(&bin.InscriptionSize, "inscription-size", 2.0, "side of the inscription")
	flag.Parse()

	bin.Inscription = LoadInscription()

	// Get a higher-resolution rim around the top.
	squeeze := &toolbox3d.AxisSqueeze{
		Axis:  toolbox3d.AxisZ,
		Ratio: 3.0,
		Min:   bin.Height - bin.Thickness,
		Max:   bin.Height,
	}

	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(model3d.TransformSolid(squeeze, &bin), 0.02, 16)
	mesh = mesh.MapCoords(squeeze.Inverse().Apply)
	log.Printf("Simplifying mesh (%d triangles)...", len(mesh.TriangleSlice()))
	mesh = mesh.EliminateCoplanar(1e-5)
	log.Printf("Exporting mesh (%d triangles)...", len(mesh.TriangleSlice()))
	mesh.SaveGroupedSTL("recycle_bin.stl")
	log.Println("Creating rendering...")
	render3d.SaveRendering("rendering.png", mesh,
		model3d.Coord3D{Y: -bin.BaseDepth * 2.5, Z: 1.5 * bin.Height},
		500, 500, nil)
}

func LoadInscription() model2d.Collider {
	bmp := model2d.MustReadBitmap("wikipedia_image.png", func(c color.Color) bool {
		_, _, _, a := c.RGBA()
		return a > 0xffff/2
	}).FlipY()
	mesh := bmp.Mesh().SmoothSq(100)
	mesh = mesh.Scale(1.0 / math.Max(float64(bmp.Width), float64(bmp.Height)))
	return model2d.MeshToCollider(mesh)
}

type BinSolid struct {
	BaseWidth float64
	BaseDepth float64
	TopScale  float64
	Height    float64

	CornerRadius float64
	Thickness    float64

	InscriptionSize float64
	Inscription     model2d.Collider
}

func (b *BinSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{
		X: -b.TopScale*b.BaseWidth/2 - b.Thickness/2,
		Y: -b.TopScale*b.BaseDepth/2 - 3*b.Thickness/2,
	}
}

func (b *BinSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{
		X: b.TopScale*b.BaseWidth/2 + b.Thickness/2,
		Y: b.TopScale*b.BaseDepth/2 + 3*b.Thickness/2,
		Z: b.Height,
	}
}

func (b *BinSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(b, c) {
		return false
	}
	frac := c.Z / b.Height
	scale := 1 + (b.TopScale-1)*frac

	width := b.BaseWidth * scale
	depth := b.BaseDepth * scale

	if b.insideInscription(c, depth) {
		return true
	}

	c2 := c.XY()

	// Bottom is fully filled-in.
	if c.Z < b.Thickness {
		return b.insideRoundedRect(c2, width+b.Thickness, depth+b.Thickness,
			b.CornerRadius+b.Thickness/2)
	}

	// Rounded edges on the top.
	thickness := b.Thickness / 2
	if c.Z > b.Height-b.Thickness/2 {
		r := b.Thickness / 2
		thickness = math.Sqrt(math.Pow(r, 2) - math.Pow(r-(b.Height-c.Z), 2))
	}

	return b.insideRoundedRect(c2, width+thickness*2, depth+thickness*2, b.CornerRadius+thickness) &&
		!b.insideRoundedRect(c2, width-thickness*2, depth-thickness*2, b.CornerRadius-thickness)
}

func (b *BinSolid) insideRoundedRect(c2 model3d.Coord2D, width, depth, radius float64) bool {
	if c2.X < -width/2 || c2.X > width/2 || c2.Y < -depth/2 || c2.Y > depth/2 {
		return false
	}

	xDist := math.Min(c2.X+width/2, width/2-c2.X)
	yDist := math.Min(c2.Y+depth/2, depth/2-c2.Y)
	if xDist > radius || yDist > radius {
		return true
	}
	xDist = radius - xDist
	yDist = radius - yDist
	return xDist*xDist+yDist*yDist < radius*radius
}

func (b *BinSolid) insideInscription(c model3d.Coord3D, depth float64) bool {
	if c.Y < 0 {
		// Put symbol on both sides.
		c.Y *= -1
	} else {
		// Fix the direction of the logo.
		c.X *= -1
	}
	baseY := depth/2 + b.Thickness/2
	if c.Y < baseY || c.Y > baseY+b.Thickness {
		return false
	}

	r := b.InscriptionSize / 2
	midY := b.Height / 2
	if c.X < -r || c.X > r || c.Z < midY-r || c.Z > midY+r {
		return false
	}

	inset := InsetSlopeFactor * (c.Y - baseY)
	c2 := model2d.Coord{X: c.X + r, Y: c.Z + r - midY}.Scale(1 / b.InscriptionSize)
	return model2d.ColliderContains(b.Inscription, c2, inset)
}
