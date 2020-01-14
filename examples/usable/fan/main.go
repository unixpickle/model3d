package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
)

const OutputDir = "models"

func main() {
	CreateMeshFile("propeller.stl", PropellerMesh)
	CreateMeshFile("spine.stl", SpineMesh)
	CreateMeshFile("small_gear.stl", SmallGearMesh)
	CreateMeshFile("crank_gear.stl", CrankGearMesh)
	CreateMeshFile("crank_bolt.stl", CrankBoltMesh)
}

func CreateMeshFile(name string, f func() *model3d.Mesh) {
	if _, err := os.Stat(OutputDir); os.IsNotExist(err) {
		essentials.Must(os.Mkdir(OutputDir, 0755))
	}
	outPath := filepath.Join(OutputDir, name)
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		log.Println("Creating mesh for", name, "...")
		mesh := f()
		log.Println("Saving mesh for", name, "...")
		mesh.SaveGroupedSTL(outPath)
	}
}
