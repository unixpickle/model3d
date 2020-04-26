package render3d

import (
	"math"
	"math/rand"
	"sort"

	"github.com/unixpickle/model3d/model3d"
)

// Color is a linear RGB color, where X, Y, and Z store R,
// G, and B respectively.
//
// Note that these colors are NOT sRGB (the standard),
// since sRGB values do not represent linear brightness.
//
// Colors should be positive, but they are not bounded on
// the positive side, since light isn't in the real world.
type Color = model3d.Coord3D

// ClampColor clamps the color into the range [0, 1].
func ClampColor(c Color) Color {
	return c.Max(Color{}).Min(Color{X: 1, Y: 1, Z: 1})
}

// NewColor creates a Color with a given brightness.
func NewColor(b float64) Color {
	return Color{X: b, Y: b, Z: b}
}

// NewColorRGB creates a Color from sRGB values.
func NewColorRGB(r, g, b float64) Color {
	return Color{X: gammaExpand(r), Y: gammaExpand(g), Z: gammaExpand(b)}
}

// RGB gets sRGB values for a Color.
func RGB(c Color) (float64, float64, float64) {
	return gammaCompress(c.X), gammaCompress(c.Y), gammaCompress(c.Z)
}

func gammaCompress(u float64) float64 {
	if u <= 0.0031308 {
		return 12.92 * u
	} else {
		return 1.055*math.Pow(u, 1/2.4) - 0.055
	}
}

func gammaExpand(u float64) float64 {
	if u <= 0.04045 {
		return u / 12.92
	} else {
		return math.Pow((u+0.055)/1.055, 2.4)
	}
}

// A PointLight is a light eminating from a point and
// going in all directions equally.
type PointLight struct {
	Origin model3d.Coord3D
	Color  Color

	// If true, the ray tracer should use an inverse
	// square relation to dim this light as it gets
	// farther from an object.
	QuadDropoff bool
}

// ColorAtDistance gets the Color produced by this light
// at some distance.
func (p *PointLight) ColorAtDistance(distance float64) Color {
	if !p.QuadDropoff {
		return p.Color
	}
	return p.Color.Scale(1 / (distance * distance))
}

// ShadeCollision determines a scaled color for a surface
// light collision.
func (p *PointLight) ShadeCollision(normal, pointToLight model3d.Coord3D) Color {
	dist := pointToLight.Norm()
	color := p.ColorAtDistance(dist)

	// Multiply by a density correction that comes from
	// lambertian shading.
	// In essence, when doing simple ray tracing, we want
	// the brightest part of a lambertian surface to have
	// the same brightness as the point light.
	density := 0.25 * math.Max(0, normal.Dot(pointToLight.Scale(1/dist)))

	return color.Scale(density)
}

// AreaLight is a surface that emits light.
//
// For regular path tracing, AreaLight is not needed,
// since lights are just regular objects.
// For global illumination methods that trace paths from
// lights to the scene, AreaLights are needed to sample
// from the light sources.
//
// An AreaLight should not be reflective in any way;
// its BSDF should be zero everywhere.
type AreaLight interface {
	Object

	// SampleLight samples a point uniformly on the
	// surface of the light and yields both the normal
	// and the emission at that point.
	SampleLight(gen *rand.Rand) (point, normal model3d.Coord3D, emission Color)

	// Area gets the total area of the light.
	Area() float64
}

// MeshAreaLight is an AreaLight for the surface of a
// mesh.
type MeshAreaLight struct {
	Object
	emission  Color
	triangles []*model3d.Triangle
	cumuAreas []float64
	totalArea float64
}

// NewMeshAreaLight creates an efficient area light from
// the triangle mesh.
func NewMeshAreaLight(mesh *model3d.Mesh, emission Color) *MeshAreaLight {
	m := &MeshAreaLight{
		Object: &ColliderObject{
			Collider: model3d.MeshToCollider(mesh),
			Material: &LambertMaterial{EmissionColor: emission},
		},
		emission:  emission,
		triangles: mesh.TriangleSlice(),
	}
	m.cumuAreas = make([]float64, len(m.triangles))
	for i, t := range m.triangles {
		m.totalArea += t.Area()
		m.cumuAreas[i] = m.totalArea
	}
	return m
}

func (m *MeshAreaLight) SampleLight(gen *rand.Rand) (point, normal model3d.Coord3D,
	emission Color) {
	triIdx := sort.SearchFloat64s(m.cumuAreas, rand.Float64()*m.totalArea)
	if triIdx == len(m.cumuAreas) {
		triIdx--
	}

	triangle := m.triangles[triIdx]

	// https://stackoverflow.com/questions/4778147/sample-random-point-in-triangle
	r1 := math.Sqrt(rand.Float64())
	r2 := rand.Float64()
	res := triangle[0].Scale(1 - r1)
	res = res.Add(triangle[1].Scale(r1 * (1 - r2)))
	res = res.Add(triangle[2].Scale(r1 * r2))
	return res, triangle.Normal(), m.emission
}

func (m *MeshAreaLight) Area() float64 {
	return m.totalArea
}

type joinedAreaLight struct {
	JoinedObject
	lights    []AreaLight
	cumuAreas []float64
	totalArea float64
}

// JoinAreaLights creates a larger AreaLight by combining
// smaller AreaLights.
func JoinAreaLights(lights ...AreaLight) AreaLight {
	jo := make(JoinedObject, len(lights))
	for i, l := range lights {
		jo[i] = l
	}
	j := &joinedAreaLight{
		JoinedObject: jo,
		lights:       lights,
		cumuAreas:    make([]float64, len(lights)),
	}
	for i, l := range lights {
		j.totalArea += l.Area()
		j.cumuAreas[i] = j.totalArea
	}
	return j
}

func (j *joinedAreaLight) SampleLight(gen *rand.Rand) (point, normal model3d.Coord3D,
	emission Color) {
	lIdx := sort.SearchFloat64s(j.cumuAreas, rand.Float64()*j.totalArea)
	if lIdx == len(j.cumuAreas) {
		lIdx--
	}
	point, normal, emission = j.lights[lIdx].SampleLight(gen)
	return
}

func (j *joinedAreaLight) Area() float64 {
	return j.totalArea
}
