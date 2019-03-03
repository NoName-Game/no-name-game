package galaxies

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
)

//Star -
type Star struct {
	Name        string
	Size        float64
	Position    mgl64.Vec3
	Temperature float64
	Color       string
}

// Stars - slice of star
type Stars []Star

// Offset -
func (s *Star) Offset(offset mgl64.Vec3) {
	s.Position = s.Position.Add(offset)
}

// Swirl -
func (s *Star) Swirl(axis mgl64.Vec3, amount float64) {
	d := s.Position.Len()
	angle := math.Pow(d, 0.1) * amount
	quaternion := mgl64.QuatRotate(angle, axis)

	s.Position = mgl64.TransformNormal(s.Position, quaternion.Mat4())
}
