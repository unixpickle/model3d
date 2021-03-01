// Generated from templates/util_test.template

package model2d

import (
	"fmt"

	"github.com/pkg/errors"
)

type Failer interface {
	Fatal(args ...interface{})
}

// ValidateMesh checks if m is manifold and has correct normals.
func ValidateMesh(m *Mesh) error {
	if !m.Manifold() {
		return errors.New("mesh is non-manifold")
	}
	if _, n := m.RepairNormals(1e-8); n != 0 {
		return fmt.Errorf("mesh has %d flipped normals", n)
	}
	return nil
}

func MustValidateMesh(f Failer, m *Mesh) {
	if err := ValidateMesh(m); err != nil {
		f.Fatal(err)
	}
}
