// Command inch_to_mm converts an STL file from using
// inches to using millimeters.
package main

import (
	"io/ioutil"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
)

func main() {
	if len(os.Args) != 3 {
		essentials.Die("Usage: inch_to_mm <input.stl> <output.stl>")
	}
	r, err := os.Open(os.Args[1])
	essentials.Must(err)
	triangles, err := model3d.ReadSTL(r)
	r.Close()
	essentials.Must(err)

	for _, t := range triangles {
		for i := range t {
			t[i] = t[i].Scale(25.4)
		}
	}

	err = ioutil.WriteFile(os.Args[2], model3d.EncodeSTL(triangles), 0755)
	essentials.Must(err)
}
