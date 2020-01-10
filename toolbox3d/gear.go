package toolbox3d

import (
	"math"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/model2d"
)

type SpurGear struct {
	P1      model3d.Coord3D
	P2      model3d.Coord3D
	Profile GearProfile
}

func (s *SpurGear) Min() model3d.Coord3D {
	return s.boundingCylinder().Min()
}

func (s *SpurGear) Max() model3d.Coord3D {
	return s.boundingCylinder().Max()
}

func (s *SpurGear) Contains(c model3d.Coord3D) bool {
	if !model3d.InSolidBounds(s, c) {
		return false
	}
	v1, v2 := s.P2.Sub(s.P1).OrthoBasis()
	return s.Profile.Contains(model2d.Coord{
		X: v1.Dot(c),
		Y: v2.Dot(c),
	})
}

func (s *SpurGear) boundingCylinder() *model3d.CylinderSolid {
	return &model3d.CylinderSolid{
		P1:     s.P1,
		P2:     s.P2,
		Radius: s.Profile.Max().X,
	}
}

type HelicalGear struct {
	P1      model3d.Coord3D
	P2      model3d.Coord3D
	Profile GearProfile
	Angle   float64
}

func (h *HelicalGear) Min() model3d.Coord3D {
	return h.boundingCylinder().Min()
}

func (h *HelicalGear) Max() model3d.Coord3D {
	return h.boundingCylinder().Max()
}

func (h *HelicalGear) Contains(c model3d.Coord3D) bool {
	if !model3d.InSolidBounds(h, c) {
		return false
	}
	axis := h.P2.Sub(h.P1)
	v1, v2 := axis.OrthoBasis()
	c2 := model2d.Coord{
		X: v1.Dot(c),
		Y: v2.Dot(c),
	}

	distUp := axis.Normalize().Dot(c.Sub(h.P1))
	radius := h.boundingCylinder().Radius
	theta := math.Tan(h.Angle) * distUp / radius

	c2 = model2d.NewMatrix2Rotation(theta).MulColumn(c2)

	return h.Profile.Contains(c2)
}

func (h *HelicalGear) boundingCylinder() *model3d.CylinderSolid {
	return &model3d.CylinderSolid{
		P1:     h.P1,
		P2:     h.P2,
		Radius: h.Profile.Max().X,
	}
}

type GearProfile interface {
	model2d.Solid
}

type involuteGearProfile struct {
	rootRadius   float64
	baseRadius   float64
	outerRadius  float64
	toothTheta   float64
	reflectTheta float64
}

// InvoluteGearProfile creates a GearProfile for a
// standard involute gear with the given specs.
func InvoluteGearProfile(pressureAngle, module, clearance float64, numTeeth int) GearProfile {
	radius := module * float64(numTeeth) / 2
	baseRadius := math.Cos(pressureAngle) * radius

	tForR := math.Sqrt(math.Pow(radius/baseRadius, 2) - 1)
	x, y := involuteCoords(tForR)

	toothTheta := math.Pi * 2 / float64(numTeeth)
	reflectTheta := toothTheta/2 + 2*math.Atan2(y, x)

	return &involuteGearProfile{
		rootRadius:   baseRadius - clearance,
		baseRadius:   baseRadius,
		outerRadius:  radius*2 - baseRadius,
		toothTheta:   toothTheta,
		reflectTheta: reflectTheta,
	}
}

func (i *involuteGearProfile) Min() model2d.Coord {
	return model2d.Coord{X: -i.outerRadius, Y: -i.outerRadius}
}

func (i *involuteGearProfile) Max() model2d.Coord {
	return i.Min().Scale(-1)
}

func (i *involuteGearProfile) Contains(c model2d.Coord) bool {
	if !model2d.InSolidBounds(i, c) {
		return false
	}
	r := c.Norm()
	if r < i.rootRadius {
		return true
	} else if r > i.outerRadius {
		return false
	}
	theta := math.Atan2(c.Y, c.X)

	// Move theta into first tooth.
	if theta < 0 {
		theta += math.Pi * 2
	}
	_, frac := math.Modf(theta / i.toothTheta)
	theta = frac * i.toothTheta

	if r < i.baseRadius {
		return theta < i.reflectTheta
	}

	tForR := math.Sqrt(math.Pow(r/i.baseRadius, 2) - 1)
	x, y := involuteCoords(tForR)
	thetaBound := math.Atan2(y, x)
	if theta < thetaBound || i.reflectTheta-theta < thetaBound {
		return false
	}

	return true
}

func involuteCoords(t float64) (float64, float64) {
	return math.Cos(t) + t*math.Sin(t), math.Sin(t) - t*math.Cos(t)
}
