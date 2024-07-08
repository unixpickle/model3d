package main

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model2d"
)

func main() {
	path := LoadPath()
	path = path.Decimate(100)
	path.SmoothSq(20)
	path.SavePathSVG("path.svg")
}

func LoadPath() *model2d.Mesh {
	f, err := os.Open("path.json.gz")
	essentials.Must(err)
	defer f.Close()
	r, err := gzip.NewReader(f)
	essentials.Must(err)
	data, err := io.ReadAll(r)
	essentials.Must(err)

	var path []struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}
	essentials.Must(json.Unmarshal(data, &path))

	res := model2d.NewMesh()
	for i, point := range path {
		if i > 0 {
			res.Add(&model2d.Segment{
				model2d.XY(path[i-1].X, path[i-1].Y),
				model2d.XY(point.X, point.Y),
			})
		}
	}
	return res
}
