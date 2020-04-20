package model3d

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// EncodeSTL encodes a list of triangles in the binary STL
// format for use in 3D printing.
func EncodeSTL(triangles []*Triangle) []byte {
	var buf bytes.Buffer
	WriteSTL(&buf, triangles)
	return buf.Bytes()
}

// WriteSTL writes a list of triangles in the binary STL
// format to w.
func WriteSTL(w io.Writer, triangles []*Triangle) error {
	if err := writeSTL(w, triangles); err != nil {
		return errors.Wrap(err, "write STL")
	}
	return nil
}

func writeSTL(w io.Writer, triangles []*Triangle) error {
	bw := bufio.NewWriter(w)
	if _, err := bw.Write(make([]byte, 80)); err != nil {
		return err
	}
	if err := binary.Write(bw, binary.LittleEndian, uint32(len(triangles))); err != nil {
		return err
	}
	floats := make([]float32, 3*4)
	for _, t := range triangles {
		castVector32(floats, t.Normal())
		castVector32(floats[3:], t[0])
		castVector32(floats[3*2:], t[1])
		castVector32(floats[3*3:], t[2])
		if err := binary.Write(bw, binary.LittleEndian, floats); err != nil {
			return err
		}
		if _, err := bw.Write([]byte{0, 0}); err != nil {
			return err
		}
	}
	return bw.Flush()
}

func castVector32(dest []float32, v Coord3D) {
	for i, x := range v.Array() {
		dest[i] = float32(x)
	}
}

// EncodePLY encodes a 3D model as a PLY file, including
// colors for every vertex.
//
// The colorFunc maps coordinates to 24-bit RGB colors.
func EncodePLY(triangles []*Triangle, colorFunc func(Coord3D) [3]uint8) []byte {
	coords := []Coord3D{}
	coordToIdx := map[Coord3D]int{}
	for _, t := range triangles {
		for _, p := range t {
			if _, ok := coordToIdx[p]; !ok {
				coordToIdx[p] = len(coords)
				coords = append(coords, p)
			}
		}
	}

	var buffer strings.Builder
	buffer.WriteString("ply\nformat ascii 1.0\n")
	buffer.WriteString(fmt.Sprintf("element vertex %d\n", len(coords)))
	buffer.WriteString("property float x\n")
	buffer.WriteString("property float y\n")
	buffer.WriteString("property float z\n")
	buffer.WriteString("property uchar red\n")
	buffer.WriteString("property uchar green\n")
	buffer.WriteString("property uchar blue\n")
	buffer.WriteString(fmt.Sprintf("element face %d\n", len(triangles)))
	buffer.WriteString("property list uchar int vertex_index\n")
	buffer.WriteString("end_header\n")
	for _, coord := range coords {
		color := colorFunc(coord)
		buffer.WriteString(fmt.Sprintf("%f %f %f %d %d %d\n", coord.X, coord.Y, coord.Z,
			int(color[0]), int(color[1]), int(color[2])))
	}
	for _, t := range triangles {
		buffer.WriteString("3")
		for _, p := range t {
			buffer.WriteByte(' ')
			buffer.WriteString(strconv.Itoa(coordToIdx[p]))
		}
		buffer.WriteByte('\n')
	}
	return []byte(buffer.String())
}

// EncodeMaterialOBJ encodes a 3D model as a zip file
// containing both an OBJ and an MTL file.
//
// The colorFunc maps faces to real-valued RGB colors.
//
// The encoding creates a different material for every
// color, so the resulting file will be much smaller if a
// few identical colors are reused for many triangles.
func EncodeMaterialOBJ(triangles []*Triangle, colorFunc func(t *Triangle) [3]float64) []byte {
	colorToMat := map[[3]float64]int{}
	colorToTriangle := map[[3]float64][]*Triangle{}
	coords := []Coord3D{}
	coordToIdx := map[Coord3D]int{}
	for _, t := range triangles {
		c := colorFunc(t)
		if _, ok := colorToMat[c]; !ok {
			colorToMat[c] = len(colorToMat)
		}
		colorToTriangle[c] = append(colorToTriangle[c], t)
		for _, p := range t {
			if _, ok := coordToIdx[p]; !ok {
				coordToIdx[p] = len(coords)
				coords = append(coords, p)
			}
		}
	}

	var objBuffer strings.Builder
	objBuffer.WriteString("mtllib material.mtl\n")
	for _, c := range coords {
		objBuffer.WriteString(fmt.Sprintf("v %f %f %f\n", c.X, c.Y, c.Z))
	}
	for color, ts := range colorToTriangle {
		objBuffer.WriteString(fmt.Sprintf("usemtl mat%d\n", colorToMat[color]))
		for _, t := range ts {
			objBuffer.WriteString(fmt.Sprintf("f %d %d %d\n", coordToIdx[t[0]]+1,
				coordToIdx[t[1]]+1, coordToIdx[t[2]]+1))
		}
	}

	var mtlBuffer strings.Builder
	for color, mat := range colorToMat {
		mtlBuffer.WriteString(fmt.Sprintf("newmtl mat%d\nillum 1\nKa %f %f %f\nKd %f %f %f\n",
			mat, color[0], color[1], color[2], color[0], color[1], color[2]))
	}

	var fullBuffer bytes.Buffer
	writer := zip.NewWriter(&fullBuffer)
	w, _ := writer.Create("object.obj")
	io.Copy(w, bytes.NewReader([]byte(objBuffer.String())))
	w, _ = writer.Create("material.mtl")
	io.Copy(w, bytes.NewReader([]byte(mtlBuffer.String())))
	writer.Close()
	return fullBuffer.Bytes()
}

// VertexColorsToTriangle creates a per-triangle color
// function that averages the colors at each of the
// vertices.
func VertexColorsToTriangle(f func(c Coord3D) [3]float64) func(t *Triangle) [3]float64 {
	return func(t *Triangle) [3]float64 {
		var sum [3]float64
		for _, c := range t {
			color := f(c)
			for i, x := range color {
				sum[i] += x / 3
			}
		}
		return sum
	}
}
