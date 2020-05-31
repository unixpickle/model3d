package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	RoomWidth     = 10.0
	RoomHeight    = 10.0
	RoomDepth     = 10.0
	RoomThickness = 1.0

	LightThickness  = 0.2
	LightWidth      = 4.0
	LightDepth      = 2.0
	LightBrightness = 10.0

	Gravity    = 10.0
	Bounciness = 0.3
)

// Scene stores the state of the rendered scene.
type Scene struct {
	ballStates  []BallState
	ballColors  []render3d.Color
	staticScene render3d.Object
	light       render3d.AreaLight
	field       ForceField
}

// NewScene creates a new scene.
func NewScene() *Scene {
	sdf, light, scene := createStaticScene()
	return &Scene{
		ballStates: []BallState{
			{
				Radius:   1.0,
				Position: model3d.Coord3D{X: 1, Y: 3, Z: 3},
				Velocity: model3d.Coord3D{X: -0.2},
			},
			{
				Radius:   1.0,
				Position: model3d.Coord3D{X: -2, Y: 2.8, Z: 4.0},
				Velocity: model3d.Coord3D{X: 0.3},
			},
			{
				Radius:   1.0,
				Position: model3d.Coord3D{X: 0, Y: 5.0, Z: 5},
				Velocity: model3d.Coord3D{Y: -0.1},
			},
		},
		ballColors: []render3d.Color{
			render3d.NewColorRGB(1, 0, 0),
			render3d.NewColorRGB(0, 1, 0),
			render3d.NewColorRGB(00.2, 0.2, 1),
		},
		staticScene: scene,
		light:       light,
		field: JoinedField{
			&ConstantField{Force: model3d.Coord3D{Z: -Gravity}},
			&CollisionField{
				Model:           sdf,
				ReboundFraction: Bounciness,
				Force:           100.0,
			},
		},
	}
}

// Scene creates a renderable scene for the current state.
func (s *Scene) Scene() (render3d.Object, render3d.AreaLight) {
	result := render3d.JoinedObject{s.staticScene}
	for i, state := range s.ballStates {
		color := s.ballColors[i]
		result = append(result, &render3d.ColliderObject{
			Collider: &model3d.Sphere{
				Center: state.Position,
				Radius: state.Radius,
			},
			Material: &render3d.PhongMaterial{
				Alpha:         50.0,
				SpecularColor: render3d.NewColor(0.1),
				DiffuseColor:  color.Scale(0.9),
			},
		})
	}
	return result, s.light
}

// Step advances the physics simulation.
func (s *Scene) Step() {
	for i := 0; i < 10; i++ {
		s.ballStates = StepWorld(s.ballStates, 0.01, s.field)
	}
}

func createStaticScene() (model3d.PointSDF, render3d.AreaLight, render3d.Object) {
	roomMesh, lightMesh := createSceneMeshes()
	roomObject := &render3d.ColliderObject{
		Collider: model3d.MeshToCollider(roomMesh),
		Material: &render3d.LambertMaterial{
			DiffuseColor: render3d.NewColor(0.5),
		},
	}
	lightObject := render3d.NewMeshAreaLight(lightMesh, render3d.NewColor(10.0))

	fullObject := model3d.JoinedSolid{
		model3d.NewColliderSolid(model3d.MeshToCollider(roomMesh)),
		model3d.NewColliderSolid(model3d.MeshToCollider(lightMesh)),
	}
	fullMesh := model3d.MarchingCubesSearch(fullObject, 0.1, 8)
	fullSDF := model3d.MeshToSDF(fullMesh)
	fullMesh.SaveGroupedSTL("/home/alex/Desktop/collider.stl")

	return fullSDF, lightObject, render3d.JoinedObject{roomObject, lightObject}
}

func createSceneMeshes() (room, light *model3d.Mesh) {
	roomMin := model3d.Coord3D{
		X: -RoomWidth / 2,
		Y: -RoomDepth,
		Z: 0,
	}
	roomMax := model3d.Coord3D{
		X: RoomWidth / 2,
		Y: RoomDepth,
		Z: RoomHeight,
	}
	thickness := model3d.Coord3D{X: 1, Y: 1, Z: 1}.Scale(RoomThickness)

	room = model3d.NewMeshRect(roomMin.Sub(thickness), roomMax.Add(thickness))
	room.AddMesh(model3d.NewMeshRect(roomMin, roomMax))
	room, _ = room.RepairNormals(1e-8)

	light = model3d.NewMeshRect(
		model3d.Coord3D{X: -LightWidth / 2, Y: (RoomDepth / 2), Z: RoomHeight - LightThickness},
		model3d.Coord3D{X: LightWidth / 2, Y: (RoomDepth / 2) + LightDepth, Z: RoomHeight},
	)

	return
}
