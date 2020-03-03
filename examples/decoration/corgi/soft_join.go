package main

import (
	"math"

	"github.com/unixpickle/model3d"
)

const (
	sdfEpsilon = 0.02
)

type SoftJoin struct {
	min     model3d.Coord3D
	max     model3d.Coord3D
	solids  []model3d.Solid
	stddevs []float64

	rbfs []*rbfSolid
}

func NewSoftJoin() *SoftJoin {
	return &SoftJoin{}
}

func (s *SoftJoin) Add(solid model3d.Solid, stddev float64) {
	s.solids = append(s.solids, solid)
	s.stddevs = append(s.stddevs, stddev)
}

func (s *SoftJoin) Build() {
	s.rbfs = nil
	for i, solid := range s.solids {
		stddev := s.stddevs[i]
		others := model3d.JoinedSolid{}
		for j, solid1 := range s.solids {
			if i != j {
				others = append(others, solid1)
			}
		}
		subSolid := &model3d.SubtractedSolid{Positive: solid, Negative: others}
		mesh := model3d.SolidToMesh(subSolid, sdfEpsilon, 0, -1, 5)
		delta := model3d.Coord3D{X: 1, Y: 1, Z: 1}.Scale(stddev * 4)
		s.rbfs = append(s.rbfs, &rbfSolid{
			Solid:  solid,
			SDF:    model3d.MeshToSDF(mesh),
			Stddev: stddev,
			Min:    solid.Min().Sub(delta),
			Max:    solid.Max().Add(delta),
		})
		if i == 0 {
			s.min = solid.Min()
			s.max = solid.Max()
		} else {
			s.min = s.min.Min(solid.Min())
			s.max = s.max.Max(solid.Max())
		}
	}
}

func (s *SoftJoin) Min() model3d.Coord3D {
	return s.min
}

func (s *SoftJoin) Max() model3d.Coord3D {
	return s.max
}

func (s *SoftJoin) Contains(c model3d.Coord3D) bool {
	if !model3d.InSolidBounds(s, c) {
		return false
	}
	var energy float64
	for _, r := range s.rbfs {
		if r.Min.Min(c) != r.Min || r.Max.Max(c) != r.Max {
			continue
		}
		energy += r.Energy(c)
		if energy >= 1 {
			return true
		}
	}
	return energy >= 1
}

type rbfSolid struct {
	Solid  model3d.Solid
	SDF    model3d.SDF
	Stddev float64
	Min    model3d.Coord3D
	Max    model3d.Coord3D
}

func (r *rbfSolid) Energy(c model3d.Coord3D) float64 {
	if r.Solid.Contains(c) {
		return 1
	}
	return math.Exp(-(math.Pow(r.SDF.SDF(c)/r.Stddev, 2) + sdfEpsilon))
}
