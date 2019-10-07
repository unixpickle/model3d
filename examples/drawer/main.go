package main

import (
	"io/ioutil"
	"log"

	"github.com/unixpickle/model3d"
)

const (
	DrawerThickness = 0.2
	DrawerHeight    = 1.0
	DrawerWidth     = 2.0
	DrawerDepth     = 2.0

	HandleLength     = 0.5
	HandlePoleRadius = 0.2
	HandleRadius     = 0.3

	ContainerThickness = 0.3
	GrooveSize         = 0.1
	PartSpacing        = 0.03
)

func main() {
	centerX := DrawerWidth/2 + GrooveSize + DrawerThickness
	centerZ := (DrawerHeight + DrawerThickness) / 2
	drawerWithHandle := model3d.JoinedSolid{
		DrawerSolid{},
		&model3d.CylinderSolid{
			P1: model3d.Coord3D{
				X: centerX,
				Y: 0,
				Z: centerZ,
			},
			P2: model3d.Coord3D{
				X: centerX,
				Y: -HandleLength,
				Z: centerZ,
			},
			Radius: HandlePoleRadius,
		},
		&model3d.SphereSolid{
			Center: model3d.Coord3D{
				X: centerX,
				Y: -HandleLength,
				Z: centerZ,
			},
			Radius: HandleRadius,
		},
	}
	log.Println("Creating drawer mesh...")
	drawer := model3d.SolidToMesh(drawerWithHandle, PartSpacing, 2, 0.8, 3)
	log.Println("Simplifying drawer mesh...")
	drawer = drawer.EliminateCoplanar(1e-8)

	log.Println("Creating container mesh...")
	container := model3d.SolidToMesh(OuterSolid{}, PartSpacing, 2, 0.8, 3)
	log.Println("Simplifying container mesh...")
	container = container.EliminateCoplanar(1e-8)

	log.Println("Combining and saving...")
	drawer.AddMesh(container)
	ioutil.WriteFile("drawer.stl", drawer.EncodeSTL(), 0755)
}

type DrawerSolid struct{}

func (d DrawerSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{}
}

func (d DrawerSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{
		X: DrawerWidth + DrawerThickness*2 + GrooveSize*2,
		Y: DrawerDepth + DrawerThickness*2,
		Z: DrawerHeight + DrawerThickness,
	}
}

func (d DrawerSolid) Contains(c model3d.Coord3D) bool {
	min := d.Min()
	max := d.Max()
	if min.Min(c) != min || max.Max(c) != max {
		return false
	}

	if c.X >= GrooveSize && c.X <= DrawerThickness+GrooveSize {
		return true
	} else if c.X >= max.X-DrawerThickness-GrooveSize && c.X <= max.X-GrooveSize {
		return true
	} else if c.X <= GrooveSize || c.X >= max.X-GrooveSize {
		center := max.Z / 2
		return c.Z >= center-GrooveSize/2 && c.Z <= center+GrooveSize/2
	} else if c.Z <= DrawerThickness {
		return true
	} else if c.Y <= DrawerThickness || c.Y >= max.Y-DrawerThickness {
		return true
	}

	return false
}

type OuterSolid struct{}

func (o OuterSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{
		X: GrooveSize - (PartSpacing + ContainerThickness),
		Y: 0,
		Z: -(PartSpacing + ContainerThickness),
	}
}

func (o OuterSolid) Max() model3d.Coord3D {
	delta := PartSpacing + ContainerThickness
	mp := (DrawerSolid{}).Max()
	mp.X += delta - GrooveSize
	mp.Y += delta
	mp.Z += delta
	return mp
}

func (o OuterSolid) Contains(c model3d.Coord3D) bool {
	min := o.Min()
	max := o.Max()

	if min.Min(c) != min || max.Max(c) != max {
		return false
	}

	if c.Z <= min.Z+ContainerThickness || c.Z >= max.Z-ContainerThickness {
		return true
	} else if c.Y >= max.Y-ContainerThickness {
		return true
	} else if c.X <= min.X+ContainerThickness || c.X >= max.X-ContainerThickness {
		center := (min.Z + max.Z) / 2
		thickness := GrooveSize/2 + PartSpacing
		return c.Z <= center-thickness || c.Z >= center+thickness ||
			c.X <= min.X+ContainerThickness-GrooveSize ||
			c.X >= max.X-ContainerThickness+GrooveSize
	}

	return false
}
