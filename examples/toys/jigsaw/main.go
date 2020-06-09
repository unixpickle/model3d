package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		DieUsage()
	}
	switch os.Args[1] {
	case "mesh":
		GenerateMesh()
	case "image":
		if len(os.Args) != 4 {
			fmt.Fprintln(os.Stderr, "the 'image' sub-command expects two arguments")
			fmt.Fprintln(os.Stderr)
			DieUsage()
		}
		GenerateImage(os.Args[2], os.Args[3])
	default:
		fmt.Fprintln(os.Stderr, "unknown sub-command:", os.Args[1])
		fmt.Fprintln(os.Stderr)
		DieUsage()
	}
}

func DieUsage() {
	fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "<command> [args]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "sub-commands are:")
	fmt.Fprintln(os.Stderr, "    mesh                   generate a 3D mesh")
	fmt.Fprintln(os.Stderr, "    image [in] [out.png]   cut slices out of an image")
	fmt.Fprintln(os.Stderr)
	os.Exit(1)
}
