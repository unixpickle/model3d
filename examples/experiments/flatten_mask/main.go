// Command flatten_mask rotates a face mask 3D model to be
// flat against the build plate.
// For some reason, the model doesn't come as such,
// requiring more supports than necessary.
//
// Download the file WASP_MYFACE_MASK.stl from
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
		fmt.Fprintln(os.Stderr, "Usage: flatten_mask <mask.stl> <output.stl>")
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

	inputMesh := model3d.NewMeshTriangles(triangles)

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
		area := EvaluateAngle(inputMesh, angle, 0.1)
		if area > bestArea {
			bestArea = area
			bestAngle = angle
		}
	}

	log.Println("Angle is:", bestAngle*180/math.Pi)

	outputMesh := RotateMesh(inputMesh, bestAngle)
	outputMesh.SaveGroupedSTL(outputPath)
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
