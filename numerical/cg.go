package numerical

// BiCGSTAB implements the biconjugate gradient stabilized
// method for inverting large asymmetric matrices.
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
	// Unpreconditioned BiCGSTAB: https://en.wikipedia.org/wiki/Biconjugate_gradient_stabilized_method
	rhoI := b.rHatZero.Dot(b.r)
	beta := (rhoI / b.rho) * (b.alpha / b.w)
	pI := b.r.Add((b.p.Sub(b.v.Scale(b.w))).Scale(beta))
	vI := b.Op(pI)
	b.alpha = rhoI / b.rHatZero.Dot(vI)
	h := b.x.Add(pI.Scale(b.alpha))
	s := b.r.Sub(vI.Scale(b.alpha))
	t := b.Op(s)
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
