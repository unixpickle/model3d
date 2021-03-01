// Generated from templates/util_test.template

package model3d

import (
	"fmt"

	"github.com/pkg/errors"
)

type Failer interface {
	Fatal(args ...interface{})
}

// ValidateMesh checks if m is manifold and has correct normals.
func ValidateMesh(m *Mesh, checkIntersections bool) error {
	if m.NeedsRepair() {
		return errors.New("mesh needs repair")
	}
	if n := len(m.SingularVertices()); n > 0 {
		return fmt.Errorf("mesh has %d singular vertices", n)
	}
	if _, n := m.RepairNormals(1e-8); n != 0 {
		return fmt.Errorf("mesh has %d flipped normals", n)
	}
	if checkIntersections {
		if n := m.SelfIntersections(); n != 0 {
			return fmt.Errorf("mesh has %d self-intersections", n)
		}
	}
	return nil
}

func MustValidateMesh(f Failer, m *Mesh, checkIntersections bool) {
	if err := ValidateMesh(m, checkIntersections); err != nil {
		f.Fatal(err)
	}
}
