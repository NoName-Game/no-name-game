package galaxies

import (
	"fmt"
	"math"
	"strings"

	"bitbucket.org/no-name-game/no-name/app/commands/godeffect/helpers/starnamer"
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/mb-14/gomarkov"
)

//Star -
type Star struct {
	Name        string
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

// GenerateName -
func (s *Star) GenerateName() {
	chain, err := starnamer.StarNameModel()
	if err != nil {
		services.ErrorHandler("Error load star name models", err)
	}

	order := chain.Order
	tokens := make([]string, 0)
	for i := 0; i < order; i++ {
		tokens = append(tokens, gomarkov.StartToken)
	}
	for tokens[len(tokens)-1] != gomarkov.EndToken {
		next, _ := chain.Generate(tokens[(len(tokens) - order):])
		tokens = append(tokens, next)
	}

	s.Name = strings.Join(tokens[order:len(tokens)-1], "")
}

//ConvertTemperature -
func (s *Star) ConvertTemperature() {
	var red, green, blue float64
	temp := s.Temperature / 100

	if temp <= 66 {
		red = 255
		green = temp
		green = 99.4708025861*math.Log(green) - 161.1195681661

		if temp <= 19 {
			blue = 0
		} else {
			blue = temp - 10
			blue = 138.5177312231*math.Log(blue) - 305.0447927307
		}

	} else {
		red = temp - 60
		red = 329.698727446 * math.Pow(red, -0.1332047592)

		green = temp - 60
		green = 288.1221695283 * math.Pow(green, -0.0755148492)

		blue = 255
	}

	s.Color = fmt.Sprintf("#%x%x%x", clamp(red, 0, 255), clamp(green, 0, 255), clamp(blue, 0, 255))
}

func clamp(x, min, max float64) uint8 {

	if x < min {
		return uint8(min)
	}
	if x > max {
		return uint8(max)
	}
	return uint8(x)
}
