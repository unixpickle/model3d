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

	Vertices   []Coord3D
	UVs        []Coord2D
	Normals    []Coord3D
	FaceGroups []*OBJFileFaceGroup
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
		if _, err := buf.WriteString(o.encode3D("v", c)); err != nil {
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
	for _, fg := range o.FaceGroups {
		if fg.Material != "" {
			if _, err := buf.WriteString("usemtl " + fg.Material + "\n"); err != nil {
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
		" " + strconv.FormatFloat(c.Y, 'f', -1, 32) + "\n"
}

func (o *OBJFile) encode3D(name string, c Coord3D) string {
	return name + " " + strconv.FormatFloat(c.X, 'f', -1, 32) +
		" " + strconv.FormatFloat(c.Y, 'f', -1, 32) +
		" " + strconv.FormatFloat(c.Z, 'f', -1, 32) + "\n"
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
	return res + "\n"
}

// MTLFileTextureMap is a configured texture map for an
// MTLFileMaterial.
type MTLFileTextureMap struct {
	Filename string

	// May be nil.
	Options map[string]string
}

// MTLFileMaterial is a single material in an MTLFile.
type MTLFileMaterial struct {
	Name             string
	Ambient          [3]float32
	Diffuse          [3]float32
	Specular         [3]float32
	SpecularExponent float32

	// Texture maps
	AmbientMap   *MTLFileTextureMap
	DiffuseMap   *MTLFileTextureMap
	SpecularMap  *MTLFileTextureMap
	HighlightMap *MTLFileTextureMap
}

// Write encodes m to a writer w.
func (m *MTLFileMaterial) Write(w io.Writer) error {
	data := bytes.NewBuffer(nil)
	data.WriteString("newmtl ")
	data.WriteString(m.Name)
	data.WriteByte('\n')

	colorNames := [3]string{"Ka", "Kd", "Ks"}
	colors := [3][3]float32{m.Ambient, m.Diffuse, m.Specular}
	for i, color := range colors {
		name := colorNames[i]
		data.WriteString(name)
		for _, c := range color {
			data.WriteByte(' ')
			data.WriteString(strconv.FormatFloat(float64(c), 'f', 4, 32))
		}
		data.WriteByte('\n')
	}
	if m.Specular != [3]float32{} {
		data.WriteString("Ns ")
		data.WriteString(strconv.FormatFloat(float64(m.SpecularExponent), 'f', -1, 32))
		data.WriteByte('\n')
	}
	textures := []*MTLFileTextureMap{m.AmbientMap, m.DiffuseMap, m.SpecularMap, m.HighlightMap}
	textureNames := []string{"map_Ka", "map_Kd", "map_Ks", "map_Ns"}
	for i, tex := range textures {
		if tex == nil {
			continue
		}
		data.WriteString(textureNames[i])
		if tex.Options != nil {
			for name, value := range tex.Options {
				data.WriteString(" -")
				data.WriteString(name)
				data.WriteByte(' ')
				data.WriteString(value)
			}
		}
		data.WriteByte(' ')
		data.WriteString(tex.Filename)
		data.WriteByte('\n')
	}
	_, err := w.Write(data.Bytes())
	return err
}

// MTLFile represents the contents of a Wavefront mtl
// file, which is a companion of an obj file.
type MTLFile struct {
	Materials []*MTLFileMaterial
}

func (m *MTLFile) Write(w io.Writer) error {
	buf := bufio.NewWriter(w)
	for _, mat := range m.Materials {
		if err := mat.Write(buf); err != nil {
			return err
		}
	}
	return buf.Flush()
}

// BuildMaterialOBJ constructs obj and mtl files from a
// triangle mesh where each triangle's color is determined
// by a function c.
//
// Since the obj file must reference the mtl file, it does
// so by the name "material.mtl". Change o.MaterialFiles
// if this is not desired.
func BuildMaterialOBJ(t []*Triangle, c func(t *Triangle) [3]float64) (o *OBJFile, m *MTLFile) {
	o = &OBJFile{
		MaterialFiles: []string{"material.mtl"},
	}
	m = &MTLFile{}

	colorToMat := map[[3]float32]int{}
	coordToIdx := NewCoordToInt()
	for _, tri := range t {
		color64 := c(tri)
		color32 := [3]float32{float32(color64[0]), float32(color64[1]), float32(color64[2])}
		matIdx, ok := colorToMat[color32]
		var group *OBJFileFaceGroup
		if !ok {
			matIdx = len(colorToMat)
			colorToMat[color32] = matIdx
			matName := "mat" + strconv.Itoa(matIdx)
			m.Materials = append(m.Materials, &MTLFileMaterial{
				Name:    matName,
				Ambient: color32,
				Diffuse: color32,
			})
			group = &OBJFileFaceGroup{Material: matName}
			o.FaceGroups = append(o.FaceGroups, group)
		} else {
			group = o.FaceGroups[matIdx]
		}
		face := [3][3]int{}
		for i, p := range tri {
			if idx, ok := coordToIdx.Load(p); !ok {
				idx = coordToIdx.Len()
				coordToIdx.Store(p, idx)
				o.Vertices = append(o.Vertices, p)
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
