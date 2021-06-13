package model3d

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	var buf bytes.Buffer
	WritePLY(&buf, triangles, colorFunc)
	return buf.Bytes()
}

// WritePLY writes the 3D model as a PLY file, including
// colors for every vertex.
//
// The colorFunc maps coordinates to 24-bit RGB colors.
func WritePLY(w io.Writer, triangles []*Triangle, colorFunc func(Coord3D) [3]uint8) error {
	if err := writePLY(bufio.NewWriter(w), triangles, colorFunc); err != nil {
		return errors.Wrap(err, "write PLY")
	}
	return nil
}

func writePLY(w *bufio.Writer, triangles []*Triangle, colorFunc func(Coord3D) [3]uint8) error {
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

	var header strings.Builder
	header.WriteString("ply\nformat ascii 1.0\n")
	header.WriteString(fmt.Sprintf("element vertex %d\n", len(coords)))
	header.WriteString("property float x\n")
	header.WriteString("property float y\n")
	header.WriteString("property float z\n")
	header.WriteString("property uchar red\n")
	header.WriteString("property uchar green\n")
	header.WriteString("property uchar blue\n")
	header.WriteString(fmt.Sprintf("element face %d\n", len(triangles)))
	header.WriteString("property list uchar int vertex_index\n")
	header.WriteString("end_header\n")

	if _, err := w.WriteString(header.String()); err != nil {
		return err
	}

	for _, coord := range coords {
		color := colorFunc(coord)
		coordLine := fmt.Sprintf("%f %f %f %d %d %d\n", coord.X, coord.Y, coord.Z,
			int(color[0]), int(color[1]), int(color[2]))
		if _, err := w.WriteString(coordLine); err != nil {
			return err
		}
	}

	var triangleBuffer strings.Builder
	for _, t := range triangles {
		triangleBuffer.Reset()
		triangleBuffer.WriteString("3")
		for _, p := range t {
			triangleBuffer.WriteByte(' ')
			triangleBuffer.WriteString(strconv.Itoa(coordToIdx[p]))
		}
		triangleBuffer.WriteByte('\n')
		if _, err := w.WriteString(triangleBuffer.String()); err != nil {
			return err
		}
	}

	return w.Flush()
}

// A DirWriter provides an interface for creating multiple
// files inside of an abstract directory.
// Only one file can be written at a time, and each file
// is automatically closed when Create() is called again.
// The final file should be written/closed with Flush().
//
// This is implemented by zip.Writer and *FSDirWriter.
type DirWriter interface {
	Create(name string) (io.Writer, error)
	Flush() error
}

// FSDirWriter implements the DirWriter interface for the
// local filesystem.
type FSDirWriter struct {
	Root string

	curFile *os.File
}

func (f *FSDirWriter) Create(name string) (io.Writer, error) {
	if f.curFile != nil {
		err := f.curFile.Close()
		f.curFile = nil
		if err != nil {
			return nil, err
		}
	}
	var err error
	f.curFile, err = os.Create(filepath.Join(f.Root, name))
	return f.curFile, err
}

func (f *FSDirWriter) Flush() error {
	if f.curFile == nil {
		return nil
	}
	res := f.curFile.Close()
	f.curFile = nil
	return res
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
	var buf bytes.Buffer
	WriteMaterialOBJ(&buf, triangles, colorFunc)
	return buf.Bytes()
}

// WriteMaterialOBJ encodes a 3D model as a zip file
// containing both an OBJ and an MTL file.
//
// The colorFunc maps faces to real-valued RGB colors.
//
// The encoding creates a different material for every
// color, so the resulting file will be much smaller if a
// few identical colors are reused for many triangles.
func WriteMaterialOBJ(w io.Writer, ts []*Triangle, colorFunc func(t *Triangle) [3]float64) error {
	if err := writeMaterialOBJ(w, ts, colorFunc); err != nil {
		return errors.Wrap(err, "write material OBJ")
	}
	return nil
}

func writeMaterialOBJ(w io.Writer, triangles []*Triangle,
	colorFunc func(t *Triangle) [3]float64) error {
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

	zipFile := zip.NewWriter(w)

	fw, err := zipFile.Create("object.obj")
	if err != nil {
		return err
	}

	buf := bufio.NewWriter(fw)
	if _, err := buf.WriteString("mtllib material.mtl\n"); err != nil {
		return err
	}
	for _, c := range coords {
		if _, err := buf.WriteString(fmt.Sprintf("v %f %f %f\n", c.X, c.Y, c.Z)); err != nil {
			return err
		}
	}
	for color, ts := range colorToTriangle {
		matLine := fmt.Sprintf("usemtl mat%d\n", colorToMat[color])
		if _, err := buf.WriteString(matLine); err != nil {
			return err
		}
		for _, t := range ts {
			faceLine := fmt.Sprintf("f %d %d %d\n", coordToIdx[t[0]]+1, coordToIdx[t[1]]+1,
				coordToIdx[t[2]]+1)
			if _, err := buf.WriteString(faceLine); err != nil {
				return err
			}
		}
	}
	if err := buf.Flush(); err != nil {
		return err
	}

	fw, err = zipFile.Create("material.mtl")
	if err != nil {
		return err
	}
	buf = bufio.NewWriter(fw)

	for color, mat := range colorToMat {
		mtlLine := fmt.Sprintf("newmtl mat%d\nillum 1\nKa %f %f %f\nKd %f %f %f\n",
			mat, color[0], color[1], color[2], color[0], color[1], color[2])
		if _, err := buf.WriteString(mtlLine); err != nil {
			return err
		}
	}
	if err := buf.Flush(); err != nil {
		return err
	}

	return zipFile.Close()
}

// An OBJFileFaceGroup is a group of faces with one
// material in a Wavefront obj file.
type OBJFileFaceGroup struct {
	// Material is the material name, or "" by default.
	Material string

	// Each face has three vertices, which itself has a
	// vertex, texture, and normal index.
	// If a texture or normal index is 0, it is omitted.
	Faces [][3][3]int
}

// An OBJFile represents the contents of a Wavefront obj
// file.
type OBJFile struct {
	MaterialFiles []string

	Vertices []Coord3D
	UVs      []Coord2D
	Normals  []Coord3D
	Faces    []*OBJFileFaceGroup
}

// Write encodes the file to w and returns the first write
// error encountered.
func (o *OBJFile) Write(w io.Writer) error {
	buf := bufio.NewWriter(w)
	for _, mtl := range o.MaterialFiles {
		_, err := buf.WriteString("mtllib " + mtl + "\n")
		if err != nil {
			return err
		}
	}
	for _, c := range o.Vertices {
		if _, err := buf.WriteString(o.encode3D("f", c)); err != nil {
			return err
		}
	}
	for _, n := range o.Normals {
		if _, err := buf.WriteString(o.encode3D("vn", n)); err != nil {
			return err
		}
	}
	for _, t := range o.UVs {
		if _, err := buf.WriteString(o.encode2D("vt", t)); err != nil {
			return err
		}
	}
	for _, fg := range o.Faces {
		if fg.Material != "" {
			if _, err := buf.WriteString("usemtl " + fg.Material); err != nil {
				return err
			}
		}
		for _, f := range fg.Faces {
			if _, err := buf.WriteString(o.encodeFace(f)); err != nil {
				return err
			}
		}
	}
	return buf.Flush()
}

func (o *OBJFile) encode2D(name string, c Coord2D) string {
	return name + " " + strconv.FormatFloat(c.X, 'f', -1, 32) +
		" " + strconv.FormatFloat(c.Y, 'f', -1, 32)
}

func (o *OBJFile) encode3D(name string, c Coord3D) string {
	return name + " " + strconv.FormatFloat(c.X, 'f', -1, 32) +
		" " + strconv.FormatFloat(c.Y, 'f', -1, 32) +
		" " + strconv.FormatFloat(c.Z, 'f', -1, 32)
}

func (o *OBJFile) encodeFace(coords [3][3]int) string {
	res := "f"
	for _, c := range coords {
		res += " "
		if c[1] == 0 && c[2] == 0 {
			res += strconv.Itoa(c[0])
		} else if c[1] == 0 && c[2] != 0 {
			res += strconv.Itoa(c[0]) + "//" + strconv.Itoa(c[2])
		} else if c[1] != 0 && c[2] == 0 {
			res += strconv.Itoa(c[0]) + "/" + strconv.Itoa(c[1])
		} else {
			res += strconv.Itoa(c[0]) + "/" + strconv.Itoa(c[1]) + "/" + strconv.Itoa(c[2])
		}
	}
	return res
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
