// Command pan_pointcloud inputs a colored point cloud from
// a PLY file, and render a pan around it.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/fileformats"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	var inputPath string
	var outputPath string
	var pointRadius float64
	var fov float64
	var cameraDist float64
	var frames int
	var resolution int
	flag.StringVar(&inputPath, "input-path", "", "input path to PLY")
	flag.StringVar(&outputPath, "output-path", "", "output path to PLY")
	flag.Float64Var(&pointRadius, "radius", 0.01, "radius of each point")
	flag.Float64Var(&fov, "fov", render3d.DefaultFieldOfView, "field of view")
	flag.Float64Var(&cameraDist, "camera-dist", -1.0, "distance of camera from center")
	flag.IntVar(&frames, "frames", 20, "number of frames to animate")
	flag.IntVar(&resolution, "resolution", 256, "size of images")
	flag.Parse()

	if inputPath == "" || outputPath == "" {
		essentials.Die("Must provide -input-path and -output-path. See -help.")
	}

	os.MkdirAll(outputPath, 0755)

	f, err := os.Open(inputPath)
	essentials.Must(err)
	defer f.Close()
	reader, err := fileformats.NewPLYReader(f)
	essentials.Must(err)

	var objects []render3d.Object
	for {
		values, element, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		essentials.Must(err)
		var red, green, blue float64
		var x, y, z float32
		for i, value := range values {
			prop := element.Properties[i]
			name := prop.Name
			switch name {
			case "red":
				red = float64(value.(fileformats.PLYValueUint8).Value) / 255
			case "green":
				green = float64(value.(fileformats.PLYValueUint8).Value) / 255
			case "blue":
				blue = float64(value.(fileformats.PLYValueUint8).Value) / 255
			case "x":
				x = value.(fileformats.PLYValueFloat32).Value
			case "y":
				y = value.(fileformats.PLYValueFloat32).Value
			case "z":
				z = value.(fileformats.PLYValueFloat32).Value
			}
		}
		objects = append(objects, &render3d.ColliderObject{
			Collider: &model3d.Sphere{
				Center: model3d.XYZ(float64(x), float64(y), float64(z)),
				Radius: pointRadius,
			},
			Material: &render3d.PhongMaterial{
				Alpha:         10.0,
				SpecularColor: render3d.NewColor(0.1),
				DiffuseColor:  render3d.NewColorRGB(red, green, blue).Scale(0.7),
				AmbientColor:  render3d.NewColorRGB(red, green, blue).Scale(0.1),
			},
		})
	}

	joined := render3d.BVHToObject(model3d.NewBVHAreaDensity(objects))
	center := joined.Min().Mid(joined.Max())
	radius := joined.Min().Dist(joined.Max())

	if cameraDist == -1 {
		cameraDist = radius
	}

	renderer := &render3d.RayCaster{
		Lights: []*render3d.PointLight{
			{
				Origin: model3d.XYZ(center.X+radius, center.Y+radius, center.Z+radius/2),
				Color:  render3d.NewColor(0.5),
			},
			{
				Origin: model3d.XYZ(center.X-radius, center.Y-radius, center.Z+radius/2),
				Color:  render3d.NewColor(0.5),
			},
		},
	}

	for i := 0; i < frames; i++ {
		log.Println("rendering view", i, "of", frames, "...")
		theta := 2 * math.Pi * float64(i) / float64(frames)
		offset := model3d.XYZ(math.Cos(theta), math.Sin(theta), 0.25).Scale(cameraDist)
		renderer.Camera = render3d.NewCameraAt(center.Add(offset), center, fov)
		out := render3d.NewImage(resolution*4, resolution*4)
		out.SetAll(render3d.NewColor(1.0))
		renderer.Render(out, joined)
		out.Downsample(4).Save(filepath.Join(outputPath, fmt.Sprintf("%05d.png", i)))
	}
}
