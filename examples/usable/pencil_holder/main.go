package main

import (
	"image"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	HolderRadius    = 0.3
	HolderThickness = 0.1
	HolderHeight    = 2.0

	MinTotalSize     = 3.0
	MaxTotalSize     = 4.0
	BottomThickness  = 0.25
	InscriptionDepth = 0.1
)

func main() {
	log.Println("Creating solid...")
	solid := NewHeartSolid()

	log.Println("Creating mesh...")
	ax := &toolbox3d.AxisSqueeze{
		Axis:  toolbox3d.AxisZ,
		Min:   BottomThickness,
		Max:   HolderHeight - 0.1,
		Ratio: 0.1,
	}
	mesh := model3d.MarchingCubesSearch(model3d.TransformSolid(ax, solid), 0.01, 8)
	mesh = mesh.MapCoords(ax.Inverse().Apply)

	log.Println("Simplifying mesh...")
	mesh = mesh.EliminateCoplanar(1e-8)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("pencil_holder.stl")

	log.Println("Saving rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 4, 4, 300, nil)
}

type HeartSolid struct {
	Image image.Image

	pixelsToInches float64
	totalSize      float64
	holderCenters  []model3d.Coord3D
}

func NewHeartSolid() *HeartSolid {
	f, err := os.Open("schematic.png")
	essentials.Must(err)
	defer f.Close()
	img, err := png.Decode(f)
	essentials.Must(err)

	res := &HeartSolid{
		Image: img,
	}
	// Find the size that causes the smallest gap between
	// the first and last holder.
	var bestDist float64
	for totalSize := MinTotalSize; totalSize < MaxTotalSize; totalSize += 0.01 {
		res.pixelsToInches = totalSize / float64(img.Bounds().Dx())
		centers := res.computeHolderCenters()
		dist := centers[0].Dist(centers[len(centers)-1])
		if dist < bestDist || res.holderCenters == nil {
			bestDist = dist
			res.holderCenters = centers
			res.totalSize = totalSize
		}
	}
	res.pixelsToInches = res.totalSize / float64(img.Bounds().Dx())
	return res
}

func (h *HeartSolid) Min() model3d.Coord3D {
	xy := -(HolderRadius + HolderThickness)
	return model3d.XYZ(xy, xy, -BottomThickness)
}

func (h *HeartSolid) Max() model3d.Coord3D {
	xy := h.totalSize + HolderRadius + HolderThickness
	return model3d.XYZ(xy, xy, HolderHeight)
}

func (h *HeartSolid) Contains(c model3d.Coord3D) bool {
	if c.Min(h.Min()) != h.Min() || c.Max(h.Max()) != h.Max() {
		return false
	}
	if h.isInBase(c) && !h.isInInscription(c) {
		return true
	}
	return h.isInHolderWall(c)
}

func (h *HeartSolid) isInHolderWall(c model3d.Coord3D) bool {
	for _, c1 := range h.holderCenters {
		d := c1.Dist(model3d.Coord3D{X: c.X, Y: c.Y})
		if (d >= HolderRadius || c.Z <= 0) && d <= HolderRadius+HolderThickness {
			return true
		}
	}
	return false
}

func (h *HeartSolid) isInBase(c model3d.Coord3D) bool {
	if c.Z > 0 {
		return false
	}
	x := int(c.X / h.pixelsToInches)
	y := h.Image.Bounds().Dy() - (int(c.Y/h.pixelsToInches) + 1)
	_, _, _, a := h.Image.At(x, y).RGBA()
	return a > 0xffff/2
}

func (h *HeartSolid) isInInscription(c model3d.Coord3D) bool {
	x := int(c.X / h.pixelsToInches)
	y := h.Image.Bounds().Dy() - (int(c.Y/h.pixelsToInches) + 1)
	r, _, _, a := h.Image.At(x, y).RGBA()
	return a > 0xffff/2 && r < 0xffff/2 && c.Z > -InscriptionDepth && c.Z <= 0
}

func (h *HeartSolid) computeHolderCenters() []model3d.Coord3D {
	threshold := HolderRadius*2 + HolderThickness
	result := []model3d.Coord3D{h.polarBorderPoint(math.Pi / 2)}
	for theta := math.Pi / 2; theta < math.Pi*2.5; theta += 0.01 {
		p := h.polarBorderPoint(theta)
		if p.Dist(result[len(result)-1]) >= threshold {
			if p.Dist(result[0]) < threshold {
				// We've looped around the shape.
				break
			}
			result = append(result, p)
		}
	}
	return result
}

func (h *HeartSolid) polarBorderPoint(theta float64) model3d.Coord3D {
	maxY := h.Image.Bounds().Dy()
	x := float64(h.Image.Bounds().Dx()) / 2
	y := float64(h.Image.Bounds().Dy()) / 2
	deltaX := math.Cos(theta)
	deltaY := math.Sin(theta)
	for x > 0 && x < float64(h.Image.Bounds().Dx()) && y > 0 && y < float64(h.Image.Bounds().Dy()) {
		_, _, _, a := h.Image.At(int(x), maxY-(int(y)+1)).RGBA()
		if a < 0xffff/2 {
			break
		}
		y += deltaY
		x += deltaX
	}
	return model3d.Coord3D{X: h.pixelsToInches * x, Y: h.pixelsToInches * y}
}
