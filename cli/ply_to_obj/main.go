// Command ply_to_obj loads a colored mesh from a PLY file
// and exports it as a textured OBJ file.
package main

import (
	"errors"
	"flag"
	"io"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/fileformats"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func main() {
	var inputPath string
	var outputPath string
	var textureSize int
	flag.StringVar(&inputPath, "input-path", "", "path to PLY file")
	flag.StringVar(&outputPath, "output-path", "", "path to output OBJ file")
	flag.IntVar(&textureSize, "texture-size", 32, "resolution of texture image")
	flag.Parse()

	if inputPath == "" || outputPath == "" {
		essentials.Die("Need -input-path and -output-path. See -help for required flags.")
	}

	f, err := os.Open(inputPath)
	essentials.Must(err)
	defer f.Close()
	reader, err := fileformats.NewPLYReader(f)
	essentials.Must(err)

	var foundFaces, foundVertices bool
	for _, element := range reader.Header().Elements {
		if element.Name == "vertex" {
			if !element.IsStandardVertex() {
				essentials.Die("unexpected vertex element:", element.Encode())
			}
			foundVertices = true
		} else if element.Name == "face" {
			if !element.IsStandardFace() {
				essentials.Die("unexpected vertex element:", element.Encode())
			}
			foundFaces = true
		}
	}
	if !foundFaces {
		essentials.Die("missing 'face' element")
	}
	if !foundVertices {
		essentials.Die("missing 'vertex' element")
	}

	var vertices []model3d.Coord3D
	var triangles [][3]int
	colors := map[model3d.Coord3D]render3d.Color{}

	for {
		values, element, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if element.Name == "face" {
			val := values[0].(fileformats.PLYValueList)
			if val.Length.(fileformats.PLYValueUint8).Value != 3 {
				essentials.Die("expected triangles but got face of count:", val.Length)
			}
			triangles = append(triangles, [3]int{
				int(val.Values[0].(fileformats.PLYValueInt32).Value),
				int(val.Values[1].(fileformats.PLYValueInt32).Value),
				int(val.Values[2].(fileformats.PLYValueInt32).Value),
			})
		} else {
			var r, g, b float64
			var x, y, z float64
			for i, value := range values {
				switch element.Properties[i].Name {
				case "red":
					r = float64(value.(fileformats.PLYValueUint8).Value) / 255.0
				case "green":
					g = float64(value.(fileformats.PLYValueUint8).Value) / 255.0
				case "blue":
					b = float64(value.(fileformats.PLYValueUint8).Value) / 255.0
				case "x":
					x = float64(value.(fileformats.PLYValueFloat32).Value)
				case "y":
					y = float64(value.(fileformats.PLYValueFloat32).Value)
				case "z":
					z = float64(value.(fileformats.PLYValueFloat32).Value)
				}
			}
			vertex := model3d.XYZ(x, y, z)
			vertices = append(vertices, vertex)
			colors[vertex] = render3d.NewColorRGB(r, g, b)
		}
	}

	mesh := model3d.NewMesh()
	for _, tri := range triangles {
		for _, v := range tri {
			if v >= len(vertices) {
				essentials.Die("vertex out of bounds")
			}
		}
		mesh.Add(&model3d.Triangle{
			vertices[tri[0]],
			vertices[tri[1]],
			vertices[tri[2]],
		})
	}

	cf := toolbox3d.CoordColorFunc(func(c model3d.Coord3D) render3d.Color {
		return colors[c]
	})
	essentials.Must(mesh.SaveQuantizedMaterialOBJ(outputPath, textureSize, cf.TriangleColor))
}
