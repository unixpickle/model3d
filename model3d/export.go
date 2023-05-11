package model3d

import (
	"archive/zip"
	"bufio"
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"strconv"

	"github.com/pkg/errors"
	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/fileformats"
	"github.com/unixpickle/model3d/numerical"
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
	if int(uint32(len(triangles))) != len(triangles) {
		return errors.New("too many triangles for STL format")
	}
	bw := bufio.NewWriter(w)
	writer, err := fileformats.NewSTLWriter(bw, uint32(len(triangles)))
	if err != nil {
		return err
	}

	for _, t := range triangles {
		verts := [3][3]float32{
			castVector32(t[0]),
			castVector32(t[1]),
			castVector32(t[2]),
		}
		if err := writer.WriteTriangle(castVector32(t.Normal()), verts); err != nil {
			return err
		}
	}
	return bw.Flush()
}

func castVector32(v Coord3D) [3]float32 {
	var res [3]float32
	for i, x := range v.Array() {
		res[i] = float32(x)
	}
	return res
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
	coords := [][3]float64{}
	colors := [][3]uint8{}
	coordToIdx := NewCoordMap[int]()
	for _, t := range triangles {
		for _, p := range t {
			if _, ok := coordToIdx.Load(p); !ok {
				coordToIdx.Store(p, len(coords))
				coords = append(coords, p.Array())
				colors = append(colors, colorFunc(p))
			}
		}
	}

	p, err := fileformats.NewPLYMeshWriter(w, len(coords), len(triangles))
	if err != nil {
		return err
	}

	for i, c := range coords {
		if err := p.WriteCoord(c, colors[i]); err != nil {
			return err
		}
	}
	for _, t := range triangles {
		idxs := [3]int{
			coordToIdx.Value(t[0]),
			coordToIdx.Value(t[1]),
			coordToIdx.Value(t[2]),
		}
		if err := p.WriteTriangle(idxs); err != nil {
			return err
		}
	}

	return nil
}

// WriteVertexColorOBJ encodes a 3D model as an OBJ file
// with vertex colors.
//
// The colorFunc maps vertices to real-valued RGB colors.
func WriteVertexColorOBJ(w io.Writer, ts []*Triangle, colorFunc func(Coord3D) [3]float64) error {
	obj := BuildVertexColorOBJ(ts, colorFunc)
	if err := obj.Write(w); err != nil {
		return errors.Wrap(err, "write vertex color OBJ")
	}
	return nil
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

// WriteQuantizedMaterialOBJ is like WriteMaterialOBJ, but
// uses a fixed-size texture image to store face colors.
func WriteQuantizedMaterialOBJ(w io.Writer, ts []*Triangle, textureSize int,
	colorFunc func(t *Triangle) [3]float64) error {
	if err := writeQuantizedMaterialOBJ(w, ts, textureSize, colorFunc); err != nil {
		return errors.Wrap(err, "write quantized material OBJ")
	}
	return nil
}

func writeQuantizedMaterialOBJ(w io.Writer, triangles []*Triangle, textureSize int,
	colorFunc func(t *Triangle) [3]float64) error {
	obj, mtl, texture := BuildQuantizedMaterialOBJ(triangles, textureSize, colorFunc)
	return WriteTexturedMaterialOBJ(w, obj, mtl, texture)
}

func WriteTexturedMaterialOBJ(w io.Writer, obj *fileformats.OBJFile, mtl *fileformats.MTLFile,
	texture image.Image) error {
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

	fw, err = zipFile.Create("texture.png")
	if err != nil {
		return err
	}
	if err := png.Encode(fw, texture); err != nil {
		return err
	}

	return zipFile.Close()
}

// BuildVertexColorOBJ constructs an obj file with vertex
// colors.
func BuildVertexColorOBJ(t []*Triangle, c func(Coord3D) [3]float64) *fileformats.OBJFile {
	o := &fileformats.OBJFile{}

	group := &fileformats.OBJFileFaceGroup{}
	coordToIdx := NewCoordToNumber[int]()
	for _, tri := range t {
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
	o.FaceGroups = append(o.FaceGroups, group)

	o.VertexColors = make([][3]float64, len(o.Vertices))
	essentials.ConcurrentMap(0, len(o.Vertices), func(i int) {
		o.VertexColors[i] = c(NewCoord3DArray(o.Vertices[i]))
	})

	return o
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

	triColors := make([][3]float32, len(t))
	essentials.ConcurrentMap(0, len(t), func(i int) {
		tri := t[i]
		color64 := c(tri)
		triColors[i] = [3]float32{float32(color64[0]), float32(color64[1]), float32(color64[2])}
	})

	colorToMat := map[[3]float32]int{}
	coordToIdx := NewCoordToNumber[int]()
	for i, tri := range t {
		color32 := triColors[i]
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

// BuildUVMapMaterialOBJ is like BuildMaterialOBJ, but
// writes texture coordinates based on a UV map.
//
// The generated material file should be saved to the name
// "material.mtl", and references a texture "texture.png"
// that should be the target of the UV map.
func BuildUVMapMaterialOBJ(t []*Triangle, uvMap MeshUVMap) (*fileformats.OBJFile, *fileformats.MTLFile) {
	uvs := [][2]float64{}
	uvToIndex := map[[2]float64]int{}
	coords := [][3]float64{}
	triIndices := [][3]int{}
	triUVIndices := [][3]int{}
	coordToIndex := map[[3]float64]int{}
	for _, tri := range t {
		var inds [3]int
		for i, c := range tri {
			vec3 := c.Array()
			if _, ok := coordToIndex[vec3]; !ok {
				coordToIndex[vec3] = len(coords)
				coords = append(coords, vec3)
			}
			inds[i] = coordToIndex[vec3] + 1
		}
		triIndices = append(triIndices, inds)

		var uvInds [3]int
		for i, c := range uvMap[tri] {
			vec2 := c.Array()
			if _, ok := uvToIndex[vec2]; !ok {
				uvToIndex[vec2] = len(uvs)
				uvs = append(uvs, vec2)
			}
			uvInds[i] = uvToIndex[vec2] + 1
		}
		triUVIndices = append(triUVIndices, uvInds)
	}

	material := &fileformats.MTLFileMaterial{
		Name:             "material",
		Ambient:          [3]float32{0.0, 0.0, 0.0},
		Diffuse:          [3]float32{1.0, 1.0, 1.0},
		SpecularExponent: 1.0,
		DiffuseMap:       &fileformats.MTLFileTextureMap{Filename: "texture.png"},
	}
	mtl := &fileformats.MTLFile{Materials: []*fileformats.MTLFileMaterial{material}}
	group := &fileformats.OBJFileFaceGroup{Material: "material"}

	for i, vertIndices := range triIndices {
		uvIndices := triUVIndices[i]
		group.Faces = append(group.Faces, [3][3]int{
			{vertIndices[0], uvIndices[0], 0},
			{vertIndices[1], uvIndices[1], 0},
			{vertIndices[2], uvIndices[2], 0},
		})
	}

	obj := &fileformats.OBJFile{
		MaterialFiles: []string{"material.mtl"},
		Vertices:      coords,
		UVs:           uvs,
		FaceGroups:    []*fileformats.OBJFileFaceGroup{group},
	}

	return obj, mtl
}

// BuildQuantizedMaterialOBJ is like BuildMaterialOBJ, but
// quantizes the triangle colors to fit in a single texture
// image, where each pixel is a different color.
//
// Returns the texture image as well as the obj file.
// The texture should be called "texture.png", and the
// material "material.mtl".
func BuildQuantizedMaterialOBJ(t []*Triangle, textureSize int,
	c func(t *Triangle) [3]float64) (*fileformats.OBJFile, *fileformats.MTLFile, *image.RGBA) {
	coords := [][3]float64{}
	triIndices := [][3]int{}
	coordToIndex := map[[3]float64]int{}
	for _, tri := range t {
		var inds [3]int
		for i, c := range tri {
			vec3 := c.Array()
			if _, ok := coordToIndex[vec3]; !ok {
				coordToIndex[vec3] = len(coords)
				coords = append(coords, vec3)
			}
			inds[i] = coordToIndex[vec3] + 1
		}
		triIndices = append(triIndices, inds)
	}

	triColors := make([][3]float32, len(t))
	essentials.ConcurrentMap(0, len(t), func(i int) {
		tri := t[i]
		color64 := c(tri)
		triColors[i] = [3]float32{float32(color64[0]), float32(color64[1]), float32(color64[2])}
	})
	texture, uvs, texIndices := buildPaletteTexture(triColors, textureSize)

	material := &fileformats.MTLFileMaterial{
		Name:             "material",
		Ambient:          [3]float32{0.0, 0.0, 0.0},
		Diffuse:          [3]float32{1.0, 1.0, 1.0},
		SpecularExponent: 1.0,
		DiffuseMap:       &fileformats.MTLFileTextureMap{Filename: "texture.png"},
	}
	mtl := &fileformats.MTLFile{Materials: []*fileformats.MTLFileMaterial{material}}
	group := &fileformats.OBJFileFaceGroup{Material: "material"}

	for i, vertIndices := range triIndices {
		texIndex := texIndices[i] + 1
		group.Faces = append(group.Faces, [3][3]int{
			{vertIndices[0], texIndex, 0},
			{vertIndices[1], texIndex, 0},
			{vertIndices[2], texIndex, 0},
		})
	}

	obj := &fileformats.OBJFile{
		MaterialFiles: []string{"material.mtl"},
		Vertices:      coords,
		UVs:           uvs,
		FaceGroups:    []*fileformats.OBJFileFaceGroup{group},
	}

	return obj, mtl, texture
}

func buildPaletteTexture(colors [][3]float32, imageSize int) (*image.RGBA, [][2]float64, []int) {
	allVecs := make([]numerical.Vec3, len(colors))
	for i, c := range colors {
		allVecs[i] = numerical.Vec3{float64(c[0]), float64(c[1]), float64(c[2])}
	}
	numCenters := imageSize * imageSize
	clusters := numerical.NewKMeans(allVecs, numCenters)
	loss := math.Inf(1)
	for i := 0; i < 5; i++ {
		loss1 := clusters.Iterate()
		if loss1 >= loss {
			break
		}
		loss = loss1
	}

	img := image.NewRGBA(image.Rect(0, 0, imageSize, imageSize))
	uvs := make([][2]float64, len(clusters.Centers))
	for i, c := range clusters.Centers {
		y := i / imageSize
		x := i % imageSize
		rgba := color.RGBA{
			R: uint8(0xff * c[0]),
			G: uint8(0xff * c[1]),
			B: uint8(0xff * c[2]),
			A: 0xff,
		}
		img.SetRGBA(x, y, rgba)
		uvs[i] = [2]float64{
			(0.5 + float64(x)) / float64(imageSize),
			1 - (0.5+float64(y))/float64(imageSize),
		}
	}

	assignments := clusters.Assign(allVecs)

	return img, uvs, assignments
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
