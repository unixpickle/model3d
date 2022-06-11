package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	// Create a rough sphere mesh with only 320 triangles.
	mesh := model3d.NewMeshIcosphere(model3d.Origin, 1.0, 4)

	// Create a material that will show off specular highlights
	// and look really bad without smooth shading.
	material := &render3d.PhongMaterial{
		Alpha:         10,
		DiffuseColor:  render3d.NewColor(0.75),
		SpecularColor: render3d.NewColor(0.15),
		AmbientColor:  render3d.NewColor(0.1),
	}

	obj1 := &render3d.ColliderObject{
		// model3d.MeshToCollider() uses flat shading, returning the
		// normals according to each individual triangle.
		Collider: model3d.MeshToCollider(mesh.Translate(model3d.X(-1.3))),
		Material: material,
	}
	text1 := LabelObject("labels/flat.png", false, model3d.XZ(-1.3, -1.7))

	obj2 := &render3d.ColliderObject{
		// model3d.MeshToInterpNormalCollider() uses Phong shading,
		// first computig mesh vertex normals and then interpolating
		// between them at every ray-triangle collision.
		Collider: model3d.MeshToInterpNormalCollider(mesh.Translate(model3d.X(1.3))),
		Material: material,
	}
	text2 := LabelObject("labels/smooth.png", true, model3d.XZ(1.3, -1.7))

	// Create a joined object for our scene.
	joined := &render3d.JoinedObject{obj1, text1, obj2, text2}

	// Render the scene from a distance, and shift the
	// camera slightly down to show the text and spheres.
	renderer := &render3d.RayCaster{
		Camera: render3d.NewCameraAt(model3d.YZ(-8.0, -0.6), model3d.Z(-0.6), 0.8),
		Lights: []*render3d.PointLight{
			{Origin: model3d.XYZ(2.0, -10.0, 4.0), Color: render3d.NewColor(1.0)},
		},
	}
	img := render3d.NewImage(768*4, 432*4)
	renderer.Render(img, joined)
	img.Downsample(4).Save("rendering.png")
}

func LabelObject(labelPath string, smooth bool, center model3d.Coord3D) render3d.Object {
	// Turn the 2D profile into a slightly rounded 3D mesh.
	profile := model2d.MustReadBitmap(labelPath, nil).Mesh().SmoothSq(5).Scale(2.0 / 512)
	profileCollider := model3d.ProfileCollider(model2d.MeshToCollider(profile), -0.05, 0.05)
	solid := model3d.NewColliderSolidInset(profileCollider, -0.01)
	mesh := model3d.MarchingCubesSearch(solid, 0.03, 8)

	// Orient & position the final mesh.
	mesh = mesh.Rotate(model3d.X(1), -math.Pi/2)
	mesh = mesh.Translate(center.Sub(mesh.Min().Mid(mesh.Max())))

	var collider model3d.Collider
	if smooth {
		collider = model3d.MeshToInterpNormalCollider(mesh)
	} else {
		collider = model3d.MeshToCollider(mesh)
	}
	return &render3d.ColliderObject{
		Collider: collider,
		Material: &render3d.PhongMaterial{
			Alpha: 10,
			// Slightly darker than the spheres.
			DiffuseColor:  render3d.NewColor(0.5),
			SpecularColor: render3d.NewColor(0.15),
			AmbientColor:  render3d.NewColor(0.1),
		},
	}
}
