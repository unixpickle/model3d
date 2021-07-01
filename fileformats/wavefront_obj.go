package fileformats

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
)

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

	Vertices   [][3]float64
	UVs        [][2]float64
	Normals    [][3]float64
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

func (o *OBJFile) encode2D(name string, c [2]float64) string {
	return name + " " + strconv.FormatFloat(c[0], 'f', -1, 32) +
		" " + strconv.FormatFloat(c[1], 'f', -1, 32) + "\n"
}

func (o *OBJFile) encode3D(name string, c [3]float64) string {
	return name + " " + strconv.FormatFloat(c[0], 'f', -1, 32) +
		" " + strconv.FormatFloat(c[1], 'f', -1, 32) +
		" " + strconv.FormatFloat(c[2], 'f', -1, 32) + "\n"
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
