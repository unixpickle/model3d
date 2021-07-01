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
	"github.com/unixpickle/model3d/fileformats"
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
	obj, mtl := BuildMaterialOBJ(triangles, colorFunc)

	zipFile := zip.NewWriter(w)

	fw, err := zipFile.Create("object.obj")
	if err != nil {
		return err
	}
	if err := obj.Write(fw); err != nil {
		return err
	}

	fw, err = zipFile.Create("material.mtl")
	if err != nil {
		return err
	}
	if err := mtl.Write(fw); err != nil {
		return err
	}

	return zipFile.Close()
}

// BuildMaterialOBJ constructs obj and mtl files from a
// triangle mesh where each triangle's color is determined
// by a function c.
//
// Since the obj file must reference the mtl file, it does
// so by the name "material.mtl". Change o.MaterialFiles
// if this is not desired.
func BuildMaterialOBJ(t []*Triangle, c func(t *Triangle) [3]float64) (o *fileformats.OBJFile,
	m *fileformats.MTLFile) {
	o = &fileformats.OBJFile{
		MaterialFiles: []string{"material.mtl"},
	}
	m = &fileformats.MTLFile{}

	colorToMat := map[[3]float32]int{}
	coordToIdx := NewCoordToInt()
	for _, tri := range t {
		color64 := c(tri)
		color32 := [3]float32{float32(color64[0]), float32(color64[1]), float32(color64[2])}
		matIdx, ok := colorToMat[color32]
		var group *fileformats.OBJFileFaceGroup
		if !ok {
			matIdx = len(colorToMat)
			colorToMat[color32] = matIdx
			matName := "mat" + strconv.Itoa(matIdx)
			m.Materials = append(m.Materials, &fileformats.MTLFileMaterial{
				Name:    matName,
				Ambient: color32,
				Diffuse: color32,
			})
			group = &fileformats.OBJFileFaceGroup{Material: matName}
			o.FaceGroups = append(o.FaceGroups, group)
		} else {
			group = o.FaceGroups[matIdx]
		}
		face := [3][3]int{}
		for i, p := range tri {
			if idx, ok := coordToIdx.Load(p); !ok {
				idx = coordToIdx.Len()
				coordToIdx.Store(p, idx)
				o.Vertices = append(o.Vertices, p.Array())
				face[i][0] = idx + 1
			} else {
				face[i][0] = idx + 1
			}
		}
		group.Faces = append(group.Faces, face)
	}

	return
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
