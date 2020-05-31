package main

import "github.com/unixpickle/model3d/model3d"

// A BallState represents the constant-motion state of a
// particle.
type BallState struct {
	Radius float64

	Position model3d.Coord3D
	Velocity model3d.Coord3D
}

// StepWorld performs a small step of a differential
// equation describing the forces acting on balls.
func StepWorld(start []BallState, stepSize float64, forceField ForceField) []BallState {
	// RK4 method.
	grad := func(state []BallState) []BallState {
		return scaleBallState(stateGradient(state, forceField), stepSize)
	}
	k1 := grad(start)
	k2 := grad(addBallStates(start, scaleBallState(k1, 0.5)))
	k3 := grad(addBallStates(start, scaleBallState(k2, 0.5)))
	k4 := grad(addBallStates(start, k3))
	return addBallStates(
		start,
		addBallStates(
			addBallStates(
				scaleBallState(k1, 1.0/6),
				scaleBallState(k2, 1.0/3),
			),
			addBallStates(
				scaleBallState(k3, 1.0/3),
				scaleBallState(k4, 1.0/6),
			),
		),
	)
}

func stateGradient(state []BallState, forceField ForceField) []BallState {
	forces := forceField.Forces(state)
	res := make([]BallState, len(state))
	for i, x := range state {
		f := forces[i]
		res[i] = BallState{
			Position: x.Velocity,
			Velocity: f,
		}
	}
	return res
}

func addBallStates(a, b []BallState) []BallState {
	res := make([]BallState, len(a))
	for i, x := range a {
		y := b[i]
		res[i] = BallState{
			Radius:   x.Radius + y.Radius,
			Position: x.Position.Add(y.Position),
			Velocity: x.Velocity.Add(y.Velocity),
		}
	}
	return res
}

func scaleBallState(a []BallState, s float64) []BallState {
	res := make([]BallState, len(a))
	for i, x := range a {
		res[i] = BallState{
			Radius:   x.Radius * s,
			Position: x.Position.Scale(s),
			Velocity: x.Velocity.Scale(s),
		}
	}
	return res
}
