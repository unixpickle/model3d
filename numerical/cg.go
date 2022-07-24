package numerical

import (
	"math"
)

// A LargeLinearSolver provides numerical solutions to
// linear systems that are implemented as functions on
// vectors.
type LargeLinearSolver interface {
	// SolveLinearSystem applies the inverse of a linear
	// operator to the vector b.
	SolveLinearSystem(op func(v Vec) Vec, b Vec) Vec
}

// BiCGSTABSolver implements LargeLinearSolver using the
// BiCGSTAB algorithm with a specified stopping condition.
//
// The solver is configured using a tolerance on either
// MSE, MAE, iterations, or a combination thereof.
// At least one stopping criterion must be provided.
type BiCGSTABSolver struct {
	MaxIters int

	// If either MSE or MAE go below these values, the
	// optimization will terminate early.
	MSETolerance float64
	MAETolerance float64
}

// SolveLinearSystem iteratively runs BiCGSTAB until a
// stopping criterion is met.
func (b *BiCGSTABSolver) SolveLinearSystem(op func(v Vec) Vec, x Vec) Vec {
	if len(x) == 0 {
		return x.Zeros()
	}
	if b.MaxIters == 0 && b.MAETolerance <= 0 && b.MSETolerance <= 0 {
		panic("no stopping criteria provided")
	}
	solver := NewBiCGSTAB(op, x, nil)
	var solution Vec
	for i := 0; b.MaxIters == 0 || i < b.MaxIters; i++ {
		solution = solver.Iter()
		if b.MSETolerance != 0 || b.MAETolerance != 0 {
			sqErr := 0.0
			absErr := 0.0
			for _, x := range op(solution).Sub(x) {
				sqErr += x * x
				absErr += math.Abs(x)
			}
			if sqErr < b.MSETolerance*float64(len(x)) || absErr < b.MAETolerance*float64(len(x)) {
				break
			}
		}
	}
	return solution
}

// BiCGSTAB implements the biconjugate gradient stabilized
// method for inverting a specific large asymmetric matrix.
//
// This can be instantiated with NewBiCGSTAB() and then
// iterated via the Iter() method, which returns the
// current solution.
type BiCGSTAB struct {
	Op func(v Vec) Vec
	B  Vec

	x        Vec
	r        Vec
	rHatZero Vec
	rho      float64
	alpha    float64
	w        float64
	v        Vec
	p        Vec

	// flag to disable more updates, since they might
	// cause NaN.
	terminate bool
}

// NewBiCGSTAB initializes the BiCGSTAB solver for the
// given linear system, as implement by op.
//
// If initGuess is non-nil, it is the initial solution.
func NewBiCGSTAB(op func(v Vec) Vec, b, initGuess Vec) *BiCGSTAB {
	if initGuess == nil {
		initGuess = b.Zeros()
	}
	r := b.Sub(op(initGuess))
	return &BiCGSTAB{
		Op: op,
		B:  b,

		x:        initGuess,
		r:        r,
		rHatZero: r,
		rho:      1,
		alpha:    1,
		w:        1,
		v:        b.Zeros(),
		p:        b.Zeros(),
	}
}

// Iter performs an iteration of the method and returns the
// current estimated solution.
func (b *BiCGSTAB) Iter() Vec {
	if b.terminate {
		return b.x
	}
	// Unpreconditioned BiCGSTAB: https://en.wikipedia.org/wiki/Biconjugate_gradient_stabilized_method
	rhoI := b.rHatZero.Dot(b.r)
	beta := (rhoI / b.rho) * (b.alpha / b.w)
	pI := b.r.Add((b.p.Sub(b.v.Scale(b.w))).Scale(beta))
	vI := b.Op(pI)
	b.alpha = rhoI / b.rHatZero.Dot(vI)
	h := b.x.Add(pI.Scale(b.alpha))
	s := b.r.Sub(vI.Scale(b.alpha))
	t := b.Op(s)

	if t.Norm() == 0 {
		// Prevent NaN due to division-by-zero
		// because this is an exact solution.
		b.terminate = true
		b.x = h
		return b.x
	}

	wI := t.Dot(s) / t.Dot(t)
	xI := h.Add(s.Scale(wI))
	rI := s.Sub(t.Scale(wI))

	b.rho = rhoI
	b.p = pI
	b.v = vI
	b.w = wI
	b.x = xI
	b.r = rI

	return b.x
}
