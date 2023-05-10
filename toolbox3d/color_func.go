package toolbox3d

import (
	"fmt"
	"log"
	"math"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/numerical"
	"github.com/unixpickle/model3d/render3d"
)

const DefaultTextureImageAntialias = 4

// CoordColorFunc wraps a generic point-to-color function
// and provides methods for various other color-using APIs.
type CoordColorFunc func(c model3d.Coord3D) render3d.Color

// RenderColor is a render3d.ColorFunc wrapper for c.F.
func (c CoordColorFunc) RenderColor(coord model3d.Coord3D, rc model3d.RayCollision) render3d.Color {
	return c(coord)
}

// TriangeColor returns sRGB colors for triangles by
// averaging the sRGB values of each vertex.
func (c CoordColorFunc) TriangleColor(t *model3d.Triangle) [3]float64 {
	sum := [3]float64{}
	for _, coord := range t {
		vertexColor := c(coord)
		r, g, b := render3d.RGB(vertexColor)
		sum[0] += r / 3
		sum[1] += g / 3
		sum[2] += b / 3
	}
	return sum
}

// QuantizedTriangleColor clusters triangle colors and
// returns a mapping from triangles to a finite space of
// colors.
//
// Inputs to the resulting function need not be contained
// in the original mesh. The mesh is only used to obtain a
// dataset for clustering.
//
// It is recommended that you call this on a cached
// CoordColorFunc to avoid re-computing colors at vertices
// shared across triangles.
func (c CoordColorFunc) QuantizedTriangleColor(mesh *model3d.Mesh,
	numColors int) func(t *model3d.Triangle) [3]float64 {
	tris := mesh.TriangleSlice()
	colors := make([]numerical.Vec3, len(tris))
	essentials.ConcurrentMap(0, len(tris), func(i int) {
		colors[i] = c.TriangleColor(tris[i])
	})
	clusters := numerical.NewKMeans(colors, numColors)
	loss := math.Inf(1)
	for i := 0; i < 5; i++ {
		loss1 := clusters.Iterate()
		if loss1 >= loss {
			break
		}
		loss = loss1
	}
	coords := make([]model3d.Coord3D, len(clusters.Centers))
	for i, c := range clusters.Centers {
		coords[i] = model3d.NewCoord3DArray(c)
	}
	table := model3d.NewCoordTree(coords)
	return func(t *model3d.Triangle) [3]float64 {
		return table.NearestNeighbor(model3d.NewCoord3DArray(c.TriangleColor(t))).Array()
	}
}

// Map returns a new CoordColorFunc which applies f to all
// input coordinates before calling c.
func (c CoordColorFunc) Map(f func(model3d.Coord3D) model3d.Coord3D) CoordColorFunc {
	return func(x model3d.Coord3D) render3d.Color {
		return c(f(x))
	}
}

// Cached wraps c in another CoordColorFunc that caches
// colors for coordinates.
//
// The cached function is safe to call concurrently from
// multiple Goroutines at once.
func (c CoordColorFunc) Cached() CoordColorFunc {
	cache := &sync.Map{}
	return func(coord model3d.Coord3D) render3d.Color {
		value, ok := cache.Load(coord)
		if ok {
			return value.(render3d.Color)
		}
		actual := c(coord)
		cache.Store(coord, actual)
		return actual
	}
}

// Transform wraps c in another CoordColorFunc that applies
// the inverse of t to input points.
func (c CoordColorFunc) Transform(t model3d.Transform) CoordColorFunc {
	tInv := t.Inverse()
	return func(coord model3d.Coord3D) render3d.Color {
		return c(tInv.Apply(coord))
	}
}

// ToTexture writes a texture image for a UV map by
// averaging colors for 3D points mapped from each pixel.
//
// This assumes that the UV map is confined to the unit
// square.
//
// The antialias argument specifies the square root of the
// number of points to sample per pixel. A value of 0 will
// default to DefaultTextureImageAntialias.
func (c CoordColorFunc) ToTexture(out *render3d.Image, mapping model3d.MeshUVMap, antialias int,
	verbose bool) {
	if antialias == 0 {
		antialias = DefaultTextureImageAntialias
	}
	mapFn := mapping.MapFn()

	dx := 1 / float64(out.Width*antialias)
	dy := 1 / float64(out.Height*antialias)
	numPixels := out.Width * out.Height
	logInterval := numPixels / 10
	essentials.ConcurrentMap(0, numPixels, func(i int) {
		x := i % out.Width
		y := i / out.Width
		if verbose && i%logInterval == 0 {
			log.Printf("- filled %.02f%% of texture", 100*(float64(i)/float64(numPixels)))
		}
		minY := float64(y) / float64(out.Height)
		minX := float64(x) / float64(out.Width)
		var sum render3d.Color
		var count float64
		for innerY := 0; innerY < antialias; innerY++ {
			finalY := minY + dy/2 + dy*float64(innerY)
			for innerX := 0; innerX < antialias; innerX++ {
				finalX := minX + dx/2 + dx*float64(innerX)
				c3d, _ := mapFn(model2d.XY(finalX, finalY))
				sum = sum.Add(c(c3d))
				count++
			}
		}
		if count > 0 {
			out.Set(x, out.Height-(y+1), sum.Scale(1/count))
		}
	})
	if verbose {
		log.Println("- filled texture")
	}
}

// SaveTexturedMaterialOBJ writes an OBJ zip file with a
// material based on the color function.
//
// The material is encoded as a texture with the given
// resolution as its side length.
//
// It is highly recommended that the texture size is a
// multiple of 2 to work best with edges in an automatic
// UV map.
//
// The provided uvMap is used to map texture coordinates to
// 3D coordinates. If it is nil, one will be built
// automatically.
func (c CoordColorFunc) SaveTexturedMaterialOBJ(path string, mesh *model3d.Mesh,
	uvMap model3d.MeshUVMap, resolution int, verbose bool) error {
	if uvMap == nil {
		uvMap = model3d.BuildAutomaticUVMap(mesh, resolution, verbose)
	}
	tris := mesh.TriangleSlice()
	obj, mtl := model3d.BuildUVMapMaterialOBJ(tris, uvMap)
	img := render3d.NewImage(resolution, resolution)
	if verbose {
		log.Printf("- constructing texture...")
	}
	c.ToTexture(img, uvMap, 0, verbose)
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "save textured material OBJ")
	}
	if err := model3d.WriteTexturedMaterialOBJ(f, obj, mtl, img.RGBA()); err != nil {
		f.Close()
		return errors.Wrap(err, "save textured material OBJ")
	}
	if err := f.Close(); err != nil {
		return errors.Wrap(err, "save textured material OBJ")
	}
	return nil
}

// ChangeFilterFunc creates a filter for mesh decimation
// that avoids decimating vertices near color changes.
//
// In particular, it returns a function that returns true
// for points further than epsilon distance of a mesh
// vertex that is part of a segment that changes color.
func (c CoordColorFunc) ChangeFilterFunc(m *model3d.Mesh,
	epsilon float64) func(c model3d.Coord3D) bool {
	changed := model3d.NewCoordMap[bool]()
	m.Iterate(func(t *model3d.Triangle) {
		for _, seg := range t.Segments() {
			if c(seg[0]) != c(seg[1]) {
				changed.Store(seg[0], true)
				changed.Store(seg[1], true)
			}
		}
	})
	points := make([]model3d.Coord3D, 0, changed.Len())
	changed.KeyRange(func(k model3d.Coord3D) bool {
		points = append(points, k)
		return true
	})
	tree := model3d.NewCoordTree(points)
	return func(c model3d.Coord3D) bool {
		return !tree.SphereCollision(c, epsilon)
	}
}

// ConstantCoordColorFunc creates a CoordColorFunc that
// returns a constant value.
func ConstantCoordColorFunc(c render3d.Color) CoordColorFunc {
	return func(x model3d.Coord3D) render3d.Color {
		return c
	}
}

// JoinedCoordColorFunc creates a CoordColorFunc that
// evaluates separate CoordColorFunc for different objects,
// where the object with maximum SDF is chosen.
//
// Pass a sequence of object, color, object, color, ...
// where objects are *model3d.Mesh or model3d.SDF, and
// colors are render3d.Color or CoordColorFunc.
func JoinedCoordColorFunc(sdfsAndColors ...any) CoordColorFunc {
	if len(sdfsAndColors)%2 != 0 {
		panic("must pass an even number of arguments")
	}
	sdfs := make([]model3d.SDF, 0, len(sdfsAndColors)/2)
	colorFns := make([]CoordColorFunc, 0, len(sdfsAndColors)/2)
	for i := 0; i < len(sdfsAndColors); i += 2 {
		switch obj := sdfsAndColors[i].(type) {
		case model3d.SDF:
			sdfs = append(sdfs, obj)
		case *model3d.Mesh:
			sdfs = append(sdfs, model3d.MeshToSDF(obj))
		default:
			panic(fmt.Sprintf("unknown type for object: %T", obj))
		}
		colorFns = append(colorFns, colorFuncFromObj(sdfsAndColors[i+1]))
	}
	return func(c model3d.Coord3D) render3d.Color {
		maxSDF := math.Inf(-1)
		var maxColorFn CoordColorFunc
		for i, sdf := range sdfs {
			value := sdf.SDF(c)
			if value > maxSDF {
				maxSDF = value
				maxColorFn = colorFns[i]
			}
		}
		return maxColorFn(c)
	}
}

// JoinedMeshCoordColorFunc combines CoordColorFuncs for
// different meshes, using the color function of the mesh
// closest to a given point.
//
// This behaves similarly to JoinedCoordColorFunc, but will
// choose the closest surface rather than the object with
// the overall greatest SDF. This should only affect points
// contained inside of the union of all of the objects.
func JoinedMeshCoordColorFunc(meshToColorFunc map[*model3d.Mesh]any) CoordColorFunc {
	allMeshes := model3d.NewMesh()
	triToColorFunc := map[*model3d.Triangle]CoordColorFunc{}
	for mesh, colorObj := range meshToColorFunc {
		colorFunc := colorFuncFromObj(colorObj)
		mesh.Iterate(func(t *model3d.Triangle) {
			// Note: if a triangle is present in multiple meshes,
			// one mesh's color func will end up owning the triangle.
			triToColorFunc[t] = colorFunc
			allMeshes.Add(t)
		})
	}
	// The mesh may not have a well-defined sign, since the different
	// meshes may overlap and disobey the even-odd rule. The resulting
	// SDF should still produce correct closest points and faces.
	faceSDF := model3d.MeshToSDF(allMeshes)
	return func(c model3d.Coord3D) render3d.Color {
		face, _, _ := faceSDF.FaceSDF(c)
		return triToColorFunc[face](c)
	}
}

// JoinedSolidCoordColorFunc creates a CoordColorFunc that
// returns colors for different solids depending on which
// solid contains a point. If multiple solids contain a
// point, the average of the solids' colors are used.
//
// The points argument is a collection of points that are
// known to be within some solid. It may either be a slice
// of points, a *CoordTree, or a *CoordMap[Coord3D]
// returned by model3d.MarchingCubesInterior.
// It can also be nil, in which case no nearest neighbor
// lookups are performed. Note that the color function will
// panic() if no solid contains a given point or its
// nearest neighbor.
//
// Since the color func must work on all points, not just
// points contained within one of the solids, a separate
// set of interior points should be provided to use for
// nearest neighbor lookup. This is the points argument.
func JoinedSolidCoordColorFunc(points any, solidsAndColors ...any) CoordColorFunc {
	var coordTree *model3d.CoordTree
	if points != nil {
		switch points := points.(type) {
		case *model3d.CoordTree:
			coordTree = points
		case []model3d.Coord3D:
			coordTree = model3d.NewCoordTree(points)
		case *model3d.CoordMap[model3d.Coord3D]:
			cs := make([]model3d.Coord3D, 0, points.Len())
			points.ValueRange(func(c model3d.Coord3D) bool {
				cs = append(cs, c)
				return true
			})
			coordTree = model3d.NewCoordTree(cs)
		default:
			panic(fmt.Sprintf("unknown type for points: %T", points))
		}
	}

	if len(solidsAndColors) == 0 || len(solidsAndColors)%2 != 0 {
		panic("must provide an even, positive number of arguments")
	}
	solids := make([]model3d.Solid, 0, len(solidsAndColors)/2)
	cfs := make([]CoordColorFunc, 0, len(solidsAndColors)/2)
	for i := 0; i < len(solidsAndColors); i += 2 {
		solids = append(solids, solidsAndColors[i].(model3d.Solid))
		cfs = append(cfs, colorFuncFromObj(solidsAndColors[i+1]))
	}

	mux := model3d.NewSolidMux(solids)

	return func(c model3d.Coord3D) render3d.Color {
		// Try without and then with the nearest neighbor to c.
		for try := 0; try < 2; try++ {
			var colorSum render3d.Color
			colorCount := mux.IterContains(c, func(i int) {
				colorSum = colorSum.Add(cfs[i](c))
			})
			if colorCount > 0 {
				return colorSum.Scale(1.0 / float64(colorCount))
			}
			if coordTree == nil {
				break
			}
			c = coordTree.NearestNeighbor(c)
		}
		if coordTree != nil {
			panic("coordinate (or its nearest neighbor) is not within any solid")
		} else {
			panic("coordinate is not within any solid")
		}
	}
}

func colorFuncFromObj(obj any) CoordColorFunc {
	switch colorFn := obj.(type) {
	case CoordColorFunc:
		return colorFn
	case func(c model3d.Coord3D) render3d.Color:
		return colorFn
	case render3d.Color:
		return ConstantCoordColorFunc(colorFn)
	default:
		panic(fmt.Sprintf("unknown type for color: %T", colorFn))
	}
}
