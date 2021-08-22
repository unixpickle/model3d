// Command fix_mask improves the WASP_MYFACE_MASK to be
// printed with many fewer support structures.
// In particular, it rotates the mask to be flat against
// the build plate, and then modifies the holes to be
// printable without supports.
//
// Download the file WASP_MYFACE_MASK.stl from either:
// https://www.3dwasp.com/en/3d-printed-mask-from-3d-scanning/
// https://www.3dwasp.com/en/downloads/my-mask-3d-printed-mask-from-3d-scanning/
package main

import (
	"fmt"
	"log"
	"math"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "Usage: fix_mask <mask.stl> <output.stl> <ring.stl>")
		os.Exit(1)
	}
	inputPath := os.Args[1]
	outputPath := os.Args[2]
	ringPath := os.Args[3]

	log.Println("Loading input mesh...")
	r, err := os.Open(inputPath)
	essentials.Must(err)
	defer r.Close()
	triangles, err := model3d.ReadSTL(r)
	essentials.Must(err)

	mesh := model3d.NewMeshTriangles(triangles)
	mesh = AutoRotate(mesh)
	mesh = FixSupport(mesh)
	mesh.SaveGroupedSTL(outputPath)

	ringMesh := CreateFilterRing(mesh)
	ringMesh.SaveGroupedSTL(ringPath)
}

// FixSupport slightly moves up triangles which require
// support on an FDM printer. In particular, it fixes the
// holes in the mask.
func FixSupport(input *model3d.Mesh) *model3d.Mesh {
	// Don't fix the bottom of the mask, just the holes.
	minZ := input.Min().Z + 30

	for i := 0; i < 100; i++ {
		var numChanged int
		input = input.MapCoords(func(c model3d.Coord3D) model3d.Coord3D {
			// Don't try to remove supports from bottom.
			if c.Z < minZ {
				return c
			}
			var normalGradient model3d.Coord3D
			for _, t := range input.Find(c) {
				if t.Normal().Z > -math.Sin(0.95*math.Pi/4) {
					continue
				}
				normalGradient = normalGradient.Add(NormalZGradient(*t, c))
			}
			if (normalGradient == model3d.Coord3D{}) {
				return c
			}
			numChanged++
			return c.Add(normalGradient.Normalize().Scale(0.01))
		})
		if numChanged == 0 {
			log.Println("Fixed support in", i, "iterations.")
			break
		}
	}
	return input
}

func NormalZGradient(t model3d.Triangle, c model3d.Coord3D) model3d.Coord3D {
	var idx int
	for i, c1 := range t {
		if c1 == c {
			idx = i
			break
		}
	}

	normal := t.Normal()
	var result [3]float64
	for i := 0; i < 3; i++ {
		arr := c.Array()
		arr[i] += 1e-5
		c1 := model3d.NewCoord3DArray(arr)
		t[idx] = c1
		normal1 := t.Normal()
		t[idx] = c
		result[i] = (normal1.Z - normal.Z) / 1e-5
	}

	return model3d.NewCoord3DArray(result)
}

// AutoRotate rotates the mask to be flat on the filter
// side, preventing the need for as much support.
func AutoRotate(inputMesh *model3d.Mesh) *model3d.Mesh {
	log.Println("Finding initial starting angle...")
	// These angles both _almost_ flatten out the base.
	minAngle := 45.0 * (math.Pi / 180)
	maxAngle := 60.0 * (math.Pi / 180)
	var bestArea float64
	var bestAngle float64

	log.Println("Roughly predicting angle...")
	for angle := minAngle; angle < maxAngle; angle += 0.01 {
		area := EvaluateAngle(inputMesh, angle, 1)
		if area > bestArea {
			bestArea = area
			bestAngle = angle
		}
	}

	log.Println("Predicting angle more precisely...")
	bestArea = 0
	for angle := bestAngle - 0.05; angle < bestAngle+0.05; angle += 0.00025 {
		area := EvaluateAngle(inputMesh, angle, 0.05)
		if area > bestArea {
			bestArea = area
			bestAngle = angle
		}
	}

	log.Println("Final rotation angle is:", bestAngle*180/math.Pi)

	return RotateMesh(inputMesh, bestAngle)
}

func EvaluateAngle(input *model3d.Mesh, angle, threshold float64) float64 {
	return BottomArea(RotateMesh(input, angle), threshold)
}

func BottomArea(m *model3d.Mesh, threshold float64) float64 {
	var result float64

	minZ := m.Min().Z
	m.Iterate(func(t *model3d.Triangle) {
		if t.Max().Z < minZ+threshold {
			result += t.Area()
		}
	})

	return result
}

func RotateMesh(m *model3d.Mesh, angle float64) *model3d.Mesh {
	return m.Rotate(model3d.X(1), angle)
}

// CreateFilterRing creates a piece of plastic that fits
// into the filter section and holds the filter in place.
func CreateFilterRing(m *model3d.Mesh) *model3d.Mesh {
	collider := model3d.MeshToCollider(m)

	// Pick a z-axis where we can slice the ring.
	sliceZ := 2 + m.Min().Z
	// Scale up the bitmap to get accurate resolution.
	const scale = 20

	log.Println("Tracing ring outline...")
	min := m.Min()
	size := m.Max().Sub(min)
	bitmap := model2d.NewBitmap(int(size.X*scale), int(size.Y*scale))
	for y := 0; y < bitmap.Height; y++ {
		for x := 0; x < bitmap.Width; x++ {
			realX := min.X + float64(x)/scale
			realY := min.Y + float64(y)/scale
			numColl := collider.RayCollisions(&model3d.Ray{
				Origin:    model3d.XYZ(realX, realY, sliceZ),
				Direction: model3d.X(1),
			}, nil)
			numColl1 := collider.RayCollisions(&model3d.Ray{
				Origin:    model3d.XYZ(realX, realY, sliceZ),
				Direction: model3d.Coord3D{X: -1},
			}, nil)
			bitmap.Set(x, y, numColl == 2 && numColl1 == 2)
		}
	}

	solid := NewRingSolid(bitmap, scale)
	squeeze := &toolbox3d.AxisSqueeze{
		Axis:  toolbox3d.AxisZ,
		Min:   1,
		Max:   4.5,
		Ratio: 0.1,
	}
	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(model3d.TransformSolid(squeeze, solid), 0.1, 8)
	mesh = mesh.Transform(squeeze.Inverse())
	log.Println("Done creating mesh...")
	mesh = mesh.FlattenBase(0)
	mesh = mesh.EliminateCoplanar(1e-8)
	return mesh
}

type RingSolid struct {
	Collider model2d.Collider
	Scale    float64

	MinVal model3d.Coord3D
	MaxVal model3d.Coord3D
}

func NewRingSolid(bmp *model2d.Bitmap, scale float64) *RingSolid {
	min := model3d.Coord3D{X: math.Inf(1), Y: math.Inf(1)}
	max := min.Scale(-1)
	for y := 0; y < bmp.Height; y++ {
		for x := 0; x < bmp.Width; x++ {
			if !bmp.Get(x, y) {
				continue
			}
			coord := model3d.Coord3D{X: float64(x) / scale, Y: float64(y) / scale}
			min = min.Min(coord)
			max = max.Max(coord)
		}
	}
	return &RingSolid{
		Collider: model2d.MeshToCollider(bmp.Mesh().Blur(0.25)),
		Scale:    scale,
		MinVal:   min.Sub(model3d.Coord3D{X: 1, Y: 1}),
		MaxVal:   max.Add(model3d.XYZ(1, 1, 8)),
	}
}

func (r *RingSolid) Min() model3d.Coord3D {
	return r.MinVal
}

func (r *RingSolid) Max() model3d.Coord3D {
	return r.MaxVal
}

func (r *RingSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(r, c) {
		return false
	}

	c2d := c.XY().Scale(r.Scale)

	innerInset := 2.0

	// Grid on the hole.
	if c.Z < 1.0 && model2d.ColliderContains(r.Collider, c2d, (innerInset-0.01)*r.Scale) {
		mid := r.Max().Mid(r.Min())
		mid1 := r.Max().Mid(mid)
		mid2 := r.Min().Mid(mid)
		for _, mp := range []model3d.Coord3D{mid, mid1, mid2} {
			if math.Abs(c.X-mp.X) < 1.0 || math.Abs(c.Y-mp.Y) < 1.0 {
				return true
			}
		}
	}

	if model2d.ColliderContains(r.Collider, c2d, innerInset*r.Scale) {
		return false
	}

	inset := 0.1
	if c.Z > 7 {
		inset -= c.Z - 7
	}

	if inset > 0 {
		return model2d.ColliderContains(r.Collider, c2d, inset*r.Scale)
	}

	return model2d.ColliderContains(r.Collider, c2d, 0) ||
		r.Collider.CircleCollision(c2d, -inset*r.Scale)
}
