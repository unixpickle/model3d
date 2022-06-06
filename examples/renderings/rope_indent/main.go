package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	RopeStops       = 500
	RopeSmoothIters = 5
	RopeSmoothRate  = 0.3
	RopeDelta       = 0.01
	ObjectDelta     = 0.01
	SoftplusSlack   = 0.02

	Frames     = 96
	Antialias  = 8
	Resolution = 512
)

func main() {
	log.Println("Creating model...")
	obj := model3d.NewRect(model3d.Ones(-0.8), model3d.Ones(0.8))
	objRounded := model3d.NewColliderSolidInset(obj, -0.2)
	objMesh := model3d.MarchingCubesSearch(objRounded, ObjectDelta, 8)
	objCollider := model3d.MeshToCollider(objMesh)

	rotationAxis := model3d.XYZ(0.1, 0, 2.0).Normalize()

	os.Mkdir("frames", 0755)
	for i := 0; i < Frames; i++ {
		log.Printf("Generating frame %d/%d...", i+1, Frames)
		theta := 2 * math.Pi * float64(i) / Frames
		xf := model3d.Rotation(rotationAxis, theta)
		mesh, ropeMeshes := WrapRopes(
			objMesh.Transform(xf),
			model3d.TransformCollider(xf, objCollider),
			true,
			&RopeDesc{
				Center: model3d.XYZ(0, 0, 0),
				Normal: model3d.XYZ(0.2, 0, 1).Normalize(),
				Radius: 0.075,
			},
			&RopeDesc{
				Center: model3d.XYZ(0.2, 0, 0),
				Normal: model3d.XZ(1, 0.3).Normalize(),
				Radius: 0.075,
			},
		)

		combined := model3d.NewMesh()
		for _, ropeMesh := range ropeMeshes {
			combined.AddMesh(ropeMesh)
		}
		img := RenderView(mesh, combined)
		img.Save(filepath.Join("frames", fmt.Sprintf("%03d.png", i)))
	}
}

func WrapRopes(objMesh *model3d.Mesh, objCollider model3d.Collider, overlapRopes bool, ropes ...*RopeDesc) (*model3d.Mesh, map[*RopeDesc]*model3d.Mesh) {
	ropeMeshes := map[*RopeDesc]*model3d.Mesh{}
	origMesh := objMesh
	for _, rope := range ropes {
		var c model3d.Collider
		if !overlapRopes {
			fullMesh := model3d.NewMesh()
			fullMesh.AddMesh(origMesh)
			for _, m := range ropeMeshes {
				fullMesh.AddMesh(m)
			}
			c = model3d.MeshToCollider(fullMesh)
		} else {
			c = objCollider
		}
		ropeMesh, ropeSDF := rope.MeshAndSDF(c)

		deform := func(c model3d.Coord3D) model3d.Coord3D {
			offset := Softplus(ropeSDF.SDF(c))
			softRadius := Softplus(rope.Radius)

			// Simulate the displacement if a circular cross-section
			// of the rope was exactly tangent to this surface.
			// This approximation is less accurate for more curved
			// objects and/or larger ropes.
			offset = softRadius * math.Sqrt(math.Abs(1-math.Pow(1-offset/softRadius, 2)))

			innerDir := rope.Center.Sub(c).ProjectOut(rope.Normal)
			disp := innerDir.Scale(offset / innerDir.Norm())
			return c.Add(disp)
		}
		objMesh = ConcurrentMapCoords(objMesh, deform)
		newRopeMeshes := map[*RopeDesc]*model3d.Mesh{}
		for k, v := range ropeMeshes {
			newRopeMeshes[k] = ConcurrentMapCoords(v, deform)
		}
		ropeMeshes = newRopeMeshes
		ropeMeshes[rope] = ropeMesh
	}

	return objMesh, ropeMeshes
}

func ConcurrentMapCoords(m *model3d.Mesh, f func(model3d.Coord3D) model3d.Coord3D) *model3d.Mesh {
	vertices := m.VertexSlice()
	newVertices := make([]model3d.Coord3D, len(vertices))
	essentials.ConcurrentMap(0, len(vertices), func(i int) {
		newVertices[i] = f(vertices[i])
	})

	mapping := model3d.NewCoordToCoord()
	for i, old := range vertices {
		mapping.Store(old, newVertices[i])
	}

	m1 := model3d.NewMesh()
	m.Iterate(func(t *model3d.Triangle) {
		t1 := *t
		for i, p := range t {
			t1[i] = mapping.Value(p)
		}
		m1.Add(&t1)
	})
	return m1
}

func RenderView(obj, ropes *model3d.Mesh) *render3d.Image {
	cameraPos := model3d.XYZ(0, 4, 2)
	renderer := &render3d.RayCaster{
		Camera: render3d.NewCameraAt(cameraPos, model3d.Coord3D{}, 0.8),
		Lights: []*render3d.PointLight{
			{Origin: model3d.XYZ(1, 4, 3), Color: render3d.NewColor(1.0)},
		},
	}
	mainObj := &render3d.ColliderObject{
		Collider: model3d.MeshToCollider(obj),
		Material: &render3d.PhongMaterial{
			Alpha:         5.0,
			SpecularColor: render3d.NewColor(0.2),
			DiffuseColor:  render3d.NewColor(0.6),
		},
	}
	ropeObj := &render3d.ColliderObject{
		Collider: model3d.MeshToCollider(ropes),
		Material: &render3d.LambertMaterial{
			DiffuseColor: render3d.NewColorRGB(0.62, 0.45, 0),
		},
	}
	full := render3d.NewImage(Resolution*Antialias, Resolution*Antialias)
	renderer.Render(full, render3d.JoinedObject{mainObj, ropeObj})
	out := render3d.NewImage(Resolution, Resolution)
	for i := 0; i < out.Height; i++ {
		for j := 0; j < out.Width; j++ {
			var sum render3d.Color
			for k := 0; k < Antialias; k++ {
				for l := 0; l < Antialias; l++ {
					offset := full.Width*(i*Antialias+k) + j*Antialias + l
					sum = sum.Add(full.Data[offset])
				}
			}
			out.Data[i*out.Width+j] = sum.Scale(1.0 / (Antialias * Antialias))
		}
	}
	return out
}

func Softplus(x float64) float64 {
	return SoftplusSlack * math.Log(1+math.Exp(x/SoftplusSlack))
}

type RopeDesc struct {
	Center model3d.Coord3D
	Normal model3d.Coord3D
	Radius float64
}

// MeshAndSDF creates a rope mesh and SDF that ties tightly
// around the given object.
func (r *RopeDesc) MeshAndSDF(obj model3d.Collider) (*model3d.Mesh, model3d.SDF) {
	segs := r.Segments(obj)
	lineMesh := model3d.NewMesh()
	for _, l := range segs {
		lineMesh.Add(&model3d.Triangle{l[0], l[1], l[1]})
	}
	lineSDF := model3d.MeshToSDF(lineMesh)
	sdf := model3d.FuncSDF(
		lineSDF.Min().AddScalar(-r.Radius),
		lineSDF.Max().AddScalar(r.Radius),
		func(c model3d.Coord3D) float64 {
			dist := math.Abs(lineSDF.SDF(c))
			return r.Radius - dist
		},
	)
	solid := toolbox3d.LineJoin(r.Radius, segs...)
	return model3d.MarchingCubesSearch(solid, RopeDelta, 8), sdf
}

// Segments creates a rope as a series of segments on the
// outside of an object.
func (r *RopeDesc) Segments(obj model3d.Collider) []model3d.Segment {
	b1, b2 := r.Normal.OrthoBasis()
	dirs := make([]model3d.Coord3D, RopeStops)
	radii := make([]float64, RopeStops)
	for i := 0; i < RopeStops; i++ {
		theta := 2 * math.Pi * float64(i%RopeStops) / float64(RopeStops)
		dir := b1.Scale(math.Cos(theta)).Add(b2.Scale(math.Sin(theta)))
		dirs[i] = dir

		var furthest float64
		ray := &model3d.Ray{Origin: r.Center, Direction: dir}
		obj.RayCollisions(ray, func(rc model3d.RayCollision) {
			furthest = math.Max(furthest, rc.Scale)
		})
		if furthest == 0 {
			panic("no meaningful ray collisions")
		}
		radii[i] = furthest
	}
	for i := 0; i < RopeSmoothIters; i++ {
		newRadii := make([]float64, RopeStops)
		for i := 0; i < RopeStops; i++ {
			prevRad := radii[(i+(RopeStops-1))%RopeStops]
			rad := radii[i]
			nextRad := radii[(i+1)%RopeStops]
			maxRad := math.Max(prevRad, math.Max(nextRad, rad))
			newRadii[i] = rad + RopeSmoothRate*(maxRad-rad)
		}
		radii = newRadii
	}
	segs := make([]model3d.Segment, RopeStops)
	for i, rad := range radii {
		dir := dirs[i]
		p1 := r.Center.Add(dir.Scale(rad))
		nextRad := radii[(i+1)%RopeStops]
		nextDir := dirs[(i+1)%RopeStops]
		p2 := r.Center.Add(nextDir.Scale(nextRad))
		segs[i] = model3d.NewSegment(p1, p2)
	}
	return segs
}
