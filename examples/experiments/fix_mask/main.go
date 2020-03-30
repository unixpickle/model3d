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
	"github.com/unixpickle/model3d"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Usage: fix_mask <mask.stl> <output.stl>")
		os.Exit(1)
	}
	inputPath := os.Args[1]
	outputPath := os.Args[2]

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
	rotation := &model3d.Matrix3Transform{
		Matrix: model3d.NewMatrix3Rotation(model3d.Coord3D{X: 1}, angle),
	}
	return m.MapCoords(rotation.Apply)
}
