// Command inch_to_mm converts an STL file from using
// inches to using millimeters.
package main

import (
	"bufio"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model3d"
)

func main() {
	if len(os.Args) < 2 {
		essentials.Die("Usage: inch_to_mm input1.stl [input2.stl ...]")
	}

	for _, inputFile := range os.Args[1:] {
		Convert(inputFile, inputFile)
	}
}

func Convert(inputFile, outputFile string) {
	r, err := os.Open(inputFile)
	essentials.Must(err)
	triangles, err := model3d.ReadSTL(r)
	r.Close()
	essentials.Must(err)

	for _, t := range triangles {
		for i := range t {
			t[i] = t[i].Scale(25.4)
		}
	}

	w, err := os.Create(outputFile)
	essentials.Must(err)
	defer w.Close()
	bufW := bufio.NewWriter(w)
	err = model3d.WriteSTL(bufW, triangles)
	essentials.Must(err)
	essentials.Must(bufW.Flush())
}
