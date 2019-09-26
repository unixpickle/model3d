package main

import (
	"io/ioutil"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
)

func main() {
	if len(os.Args) != 3 {
		essentials.Die("Usage: off_to_stl <file.off> <file.stl>")
	}
	offFile := os.Args[1]
	stlFile := os.Args[2]

	f, err := os.Open(offFile)
	essentials.Must(err)
	defer f.Close()

	mesh, err := model3d.ReadOFF(f)
	essentials.Must(err)

	essentials.Must(ioutil.WriteFile(stlFile, mesh.EncodeSTL(), 0755))
}
