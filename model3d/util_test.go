// Generated from templates/util_test.template

package model3d

import (
	"fmt"
	"math"

	"github.com/pkg/errors"
)

type Failer interface {
	Fatal(args ...any)
}

// ValidateMesh checks if m is manifold and has correct normals.
//
// If checkExtra is true, then self-intersections are tested.
func ValidateMesh(m *Mesh, checkExtra bool) error {
	if m.NeedsRepair() {
		return errors.New("mesh needs repair")
	}
	if n := len(m.SingularVertices()); n > 0 {
		return fmt.Errorf("mesh has %d singular vertices", n)
	}
	if _, n := m.RepairNormals(1e-8); n != 0 {
		return fmt.Errorf("mesh has %d flipped normals", n)
	}
	if checkExtra {
		if n := m.SelfIntersections(); n != 0 {
			return fmt.Errorf("mesh has %d self-intersections", n)
		}
	}
	volume := m.Volume()
	if math.IsNaN(volume) || math.IsInf(volume, 0) {
		return fmt.Errorf("volume is %f", volume)
	}
	return nil
}

func MustValidateMesh(f Failer, m *Mesh, checkExtra bool) {
	if err := ValidateMesh(m, checkExtra); err != nil {
		f.Fatal(err)
	}
}
