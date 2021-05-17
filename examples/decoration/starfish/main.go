package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	MaxThickness = 0.2
	MinThickness = 0.02
)

func main() {
	solid := model3d.JoinedSolid{}
	for _, curve := range ArmCurves() {
		solid = append(solid, NewArm(curve))
	}
	solid = append(solid, CreateBase())
	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.01, 16)
	mesh = mesh.Transform(model3d.Rotation(model3d.X(1), math.Pi/2))
	mesh = mesh.FlipDelaunay()
	log.Println("Smoothing...")
	mesh = mesh.SmoothAreas(0.01, 16)

	log.Println("Saving...")
	triColor := model3d.VertexColorsToTriangle(func(c model3d.Coord3D) [3]float64 {
		r, g, b := render3d.RGB(ColorFunc(c, model3d.RayCollision{}))
		return [3]float64{r, g, b}
	})
	mesh.SaveMaterialOBJ("starfish.zip", triColor)

	log.Println("Rendering...")
	render3d.SaveRendering("rendering.png", mesh, model3d.XYZ(0.8, -3.0, 0.8), 500, 500, ColorFunc)
}

func CreateBase() model3d.Solid {
	return model3d.NewRect(model3d.XYZ(-1, -1.01, -0.3), model3d.XYZ(1, -0.9, 0.3))
}

func ColorFunc(c model3d.Coord3D, rc model3d.RayCollision) render3d.Color {
	if c.Z < CreateBase().Max().Y+1e-4 {
		return render3d.NewColorRGB(0xc2/255.0, 0xb2/255.0, 0x80/255.0)
	} else {
		return render3d.NewColorRGB(1, 153.0/255.0, 185.0/255.0)
	}
}

func ArmCurves() []model2d.BezierCurve {
	return []model2d.BezierCurve{
		// Top
		{
			model2d.XY(0, 0),
			model2d.XY(0.1, 0.5),
			model2d.XY(-0.2, 1.0),
		},
		// Right and left.
		{
			model2d.XY(0, 0),
			model2d.XY(0.3, 0.0),
			model2d.XY(0.5, 0.1),
			model2d.XY(1.0, -0.2),
		},
		{
			model2d.XY(0, 0),
			model2d.XY(-0.3, 0.0),
			model2d.XY(-0.5, -0.1),
			model2d.XY(-1.0, 0.2),
		},
		// Bottom right and left
		{
			model2d.XY(0, 0),
			model2d.XY(0.5, -0.4),
			model2d.XY(0.5, -1.0),
		},
		{
			model2d.XY(0, 0),
			model2d.XY(-0.5, -0.4),
			model2d.XY(-0.5, -1.0),
		},
	}
}

type Arm struct {
	Mesh      *model2d.Mesh
	SDF       model2d.FaceSDF
	Dists     map[model2d.Coord]float64
	TotalDist float64
}

func NewArm(curve model2d.BezierCurve) *Arm {
	mesh := model2d.NewMesh()
	dists := map[model2d.Coord]float64{curve.Eval(0): 0.0}
	dist := 0.0
	for t := 0.0; t < 1.0; t += 0.01 {
		c1 := curve.Eval(t)
		c2 := curve.Eval(t + 0.01)
		seg := &model2d.Segment{c1, c2}
		mesh.Add(seg)
		dist += seg.Length()
		dists[c2] = dist
	}
	return &Arm{
		Mesh:      mesh,
		SDF:       model2d.MeshToSDF(mesh),
		Dists:     dists,
		TotalDist: dist,
	}
}

func (t *Arm) Min() model3d.Coord3D {
	min := t.SDF.Min()
	return model3d.XYZ(min.X-MaxThickness, min.Y-MaxThickness, -MaxThickness)
}

func (t *Arm) Max() model3d.Coord3D {
	max := t.SDF.Max()
	return model3d.XYZ(max.X+MaxThickness, max.Y+MaxThickness, MaxThickness)
}

func (t *Arm) Contains(c model3d.Coord3D) bool {
	proj, dist := t.projDist(c.XY())
	thickness := (MaxThickness-MinThickness)*(1-dist/t.TotalDist) + MinThickness
	projDist := proj.Dist(c.XY())
	return math.Abs(projDist)+math.Abs(c.Z) < thickness
}

func (t *Arm) projDist(c model2d.Coord) (model2d.Coord, float64) {
	seg, proj, _ := t.SDF.FaceSDF(c)

	if proj.Norm() < 1e-5 {
		// No true points past the endpoint, which always
		// happens to be at the origin.
		return proj, math.Inf(1)
	}

	return proj, proj.Dist(seg[0]) + t.Dists[seg[0]]
}
