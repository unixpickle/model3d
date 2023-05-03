// Generated from templates/util_test.template

package model2d

import (
	"fmt"

	"github.com/pkg/errors"
)

type Failer interface {
	Fatal(args ...any)
}

// ValidateMesh checks if m is manifold and has correct normals.
//
// If checkExtra is true, then normals are checked universally.
// Otherwise, it is ensured that orientation is correct, but
// normals could be flipped.
func ValidateMesh(m *Mesh, checkExtra bool) error {
	if !m.Manifold() {
		return errors.New("mesh is non-manifold")
	}
	if checkExtra {
		if _, n := m.RepairNormals(1e-8); n != 0 {
			return fmt.Errorf("mesh has %d flipped normals", n)
		}
	} else {
		if n := len(m.InconsistentVertices()); n != 0 {
			return fmt.Errorf("mesh has %d inconsistent vertices", n)
		}
	}

	return nil
}

func MustValidateMesh(f Failer, m *Mesh, checkExtra bool) {
	if err := ValidateMesh(m, checkExtra); err != nil {
		f.Fatal(err)
	}
}
