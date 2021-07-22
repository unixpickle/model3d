package numerical

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/unixpickle/essentials"
)

// A Polynomial is an equation of the form
//
//     a0 + a1*x + a2*x^2 + a3*x^3 + ...
//
// Here, the polynomial is represented as an array of
// [a0, a1, ...].
type Polynomial []float64

// String returns a string representation of p.
func (p Polynomial) String() string {
	parts := make([]string, len(p))
	for i, x := range p {
		parts[i] = fmt.Sprintf("%f", x)
		if i == 1 {
			parts[i] += "x"
		} else if i > 1 {
			parts[i] += fmt.Sprintf("x^%d", i)
		}
	}
	return strings.Join(parts, " + ")
}

// Eval evaluates the polynomial at the given x value.
func (p Polynomial) Eval(x float64) float64 {
	xP := 1.0
	res := 0.0
	for _, c := range p {
		res += c * xP
		xP *= x
	}
	return res
}

// Derivative computes the derivative of the polynomial.
func (p Polynomial) Derivative() Polynomial {
	if len(p) == 1 {
		return Polynomial{}
	}
	res := make(Polynomial, len(p)-1)
	for i, c := range p[1:] {
		res[i] = c * float64(i+1)
	}
	return res
}

// Add returns the sum of p and p1.
func (p Polynomial) Add(p1 Polynomial) Polynomial {
	res := make(Polynomial, essentials.MaxInt(len(p), len(p1)))
	for i := range res {
		if i < len(p) {
			res[i] += p[i]
		}
		if i < len(p1) {
			res[i] += p1[i]
		}
	}
	// Terms may have cancelled out.
	for len(res) > 0 && res[len(res)-1] == 0 {
		res = res[:len(res)-1]
	}
	return res
}

// Mul computes the product of two polynomials.
func (p Polynomial) Mul(p1 Polynomial) Polynomial {
	if len(p) == 0 || len(p1) == 0 {
		return Polynomial{}
	}
	res := make(Polynomial, len(p)+len(p1)-1)
	for i, x := range p {
		for j, y := range p1 {
			res[i+j] += x * y
		}
	}
	return res
}

// RealRoots computes the real roots of p, i.e. values of
// X such that p(x) = 0. The result may have duplicates
// since roots can be repeated.
//
// If the polynomial has an infinite number of roots, one
// NaN root is returned.
func (p Polynomial) RealRoots() []float64 {
	result := make([]float64, 0, len(p))
	p.IterRealRoots(func(x float64) bool {
		result = append(result, x)
		return true
	})
	return result
}

// IterRealRoots iterates over the real roots of p.
// This is similar to RealRoots(), but allows the caller
// to stop iteration early be returning false.
func (p Polynomial) IterRealRoots(f func(x float64) bool) {
	if len(p) == 0 {
		f(math.NaN())
		return
	} else if p[len(p)-1] == 0 {
		p[:len(p)-1].IterRealRoots(f)
		return
	} else if len(p) == 1 {
		return
	} else if len(p) == 2 {
		// Inverse of linear equation.
		f(-p[0] / p[1])
		return
	} else if len(p) == 3 {
		// Quadratic formula.
		a, b, c := p[2], p[1], p[0]
		sqrtMe := b*b - 4*a*c
		if sqrtMe < 0 {
			return
		}
		sqrt := math.Sqrt(sqrtMe)
		root1 := (-b - sqrt) / (2 * a)
		root2 := (-b + sqrt) / (2 * a)
		if root1 > root2 {
			root1, root2 = root2, root1
		}
		if f(root1) {
			f(root2)
		}
		return
	}

	// Cauchy's bound for real roots.
	absBound := 0.0
	for _, x := range p[:len(p)-1] {
		absBound = math.Max(absBound, math.Abs(x/p[len(p)-1]))
	}
	absBound += 1.0

	y1 := p.Eval(-absBound)
	y2 := p.Eval(absBound)

	// Fast path to avoid finding roots of derivative.
	if y1 == 0 {
		if f(-absBound) {
			p.divideRoot(-absBound).IterRealRoots(f)
		}
		return
	} else if y2 == 0 {
		if f(absBound) {
			p.divideRoot(absBound).IterRealRoots(f)
		}
		return
	} else if (y1 < 0) != (y2 < 0) {
		r := p.searchRoot(-absBound, absBound)
		if f(r) {
			p.divideRoot(r).IterRealRoots(f)
		}
		return
	}

	extrema := make([]float64, 2, len(p)+2)
	extremaY := make([]float64, 2, len(p)+2)
	extrema[0] = -absBound
	extrema[1] = absBound
	for i, x := range extrema {
		extremaY[i] = p.Eval(x)
	}

	var root float64
	var foundRoot bool

	p.Derivative().IterRealRoots(func(x float64) bool {
		idx := sort.SearchFloat64s(extrema, x)
		extrema = append(extrema, 0)
		extremaY = append(extremaY, 0)
		copy(extrema[idx+1:], extrema[idx:])
		copy(extremaY[idx+1:], extremaY[idx:])
		extrema[idx] = x
		extremaY[idx] = p.Eval(x)
		for i := idx; i < idx+2; i++ {
			if i == 0 || i >= len(extrema) {
				continue
			}
			x1 := extrema[i-1]
			x2 := extrema[i]
			y1 := extremaY[i-1]
			y2 := extremaY[i]
			if y1 == 0 {
				foundRoot = true
				root = x1
			} else if y2 == 0 {
				foundRoot = true
				root = x2
			} else if (y1 < 0) != (y2 < 0) {
				root = p.searchRoot(x1, x2)
				foundRoot = true
			}
			if foundRoot {
				break
			}
		}
		return !foundRoot
	})
	if foundRoot {
		if f(root) {
			p.divideRoot(root).IterRealRoots(f)
		}
	}
}

// divideRoot returns the polynomial resulting from
// dividing the root x out of p.
// This will yield invalid results if x is not a root.
func (p Polynomial) divideRoot(x float64) Polynomial {
	if len(p) < 2 {
		panic("cannot divide root out of constant equation")
	} else if len(p) == 2 {
		// Assume that the root is correct.
		return Polynomial{1}
	}
	res := make(Polynomial, len(p)-1)
	temp := append(Polynomial{}, p...)
	for i := len(p) - 1; i > 0; i-- {
		res[i-1] = temp[i]
		temp[i-1] += x * temp[i]
	}
	return res
}

// searchRoot finds a root between two x values with
// different signs for p(x).
func (p Polynomial) searchRoot(x1, x2 float64) float64 {
	y1 := p.Eval(x1)
	diff := math.Abs(x1 - x2)
	for {
		mx := (x1 + x2) / 2
		my := p.Eval(mx)
		if my == 0 {
			return mx
		} else if (my < 0) == (y1 < 0) {
			x1 = mx
			y1 = my
		} else {
			x2 = mx
		}
		diff1 := math.Abs(x1 - x2)
		if diff1 >= diff {
			// Our bounds are as close as they can be in
			// this numeric format.
			break
		}
		diff = diff1
	}
	return (x1 + x2) / 2
}
