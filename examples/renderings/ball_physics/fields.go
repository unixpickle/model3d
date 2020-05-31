package main

import (
	"github.com/unixpickle/model3d/model3d"
)

// A ForceField determines forces applied to moving balls
// in a 3D scene.
type ForceField interface {
	Forces([]BallState) []model3d.Coord3D
}

// A JoinedField is a ForceField that adds ForceFields.
type JoinedField []ForceField

// Forces computes the combined forces on the state.
func (j JoinedField) Forces(state []BallState) []model3d.Coord3D {
	res := j[0].Forces(state)
	for _, f := range j[1:] {
		for i, x := range f.Forces(state) {
			res[i] = res[i].Add(x)
		}
	}
	return res
}

// A CollisionField is a force field that
type CollisionField struct {
	// Model is a 3D model that collisions occur with.
	Model model3d.PointSDF

	// ReboundFraction determines how bouncy the model
	// surface is.
	//
	// A value of 1 completely conserves energy.
	// Values less than 1 dampen the energy after every
	// collision.
	//
	// This does not affect the perfectly elastic sphere
	// to sphere collisions.
	ReboundFraction float64

	// Force is the amount of force applied during the
	// incoming part of a collision.
	//
	// Should be a large number to prevent too much
	// overlap.
	Force float64
}

// Forces computes the collision forces on each particle.
func (c *CollisionField) Forces(state []BallState) []model3d.Coord3D {
	forces := make([]model3d.Coord3D, len(state))
	addSDF := func(ps model3d.PointSDF, ignore int) {
		for i, ball := range state {
			if i == ignore {
				continue
			}
			closestPoint, sdf := ps.PointSDF(ball.Position)
			if -sdf > ball.Radius {
				// No collision is taking place.
				continue
			}
			forceDirection := ball.Position.Sub(closestPoint).Normalize()
			if sdf > 0 {
				// Center of ball is inside the surface.
				forceDirection = forceDirection.Scale(-1)
			}
			if ignore == -1 {
				// Only use rebound fraction for collisions with the model.
				if ball.Velocity.Dot(forceDirection) > 0 {
					forceDirection = forceDirection.Scale(c.ReboundFraction)
				}
			}
			forces[i] = forces[i].Add(forceDirection.Scale(c.Force))
		}
	}

	addSDF(c.Model, -1)
	for i, b := range state {
		addSDF(&model3d.Sphere{
			Center: b.Position,
			Radius: b.Radius,
		}, i)
	}

	return forces
}

// A ConstantField is a force field with a constant force
// in some direction (e.g. gravity).
type ConstantField struct {
	Force model3d.Coord3D
}

// Forces returns the same force for every particle.
func (c *ConstantField) Forces(state []BallState) []model3d.Coord3D {
	forces := make([]model3d.Coord3D, len(state))
	for i := range forces {
		forces[i] = c.Force
	}
	return forces
}
