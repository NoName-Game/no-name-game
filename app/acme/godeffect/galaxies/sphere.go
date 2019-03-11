package galaxies

import (
	"math"

	"bitbucket.org/no-name-game/no-name/app/acme/godeffect/helpers"
	"github.com/go-gl/mathgl/mgl64"
)

// Sphere - Sphere
type Sphere struct {
	Size      float64
	Density   int
	Deviation float64
}

// Generate - generate sphere
func (s Sphere) Generate() (stars Stars) {
	if s.Density > 0 {
		for i := 0; i < s.Density; i++ {
			pos := mgl64.Vec3{
				helpers.NormallyDistributedSingle(s.Deviation*s.Size, 0),
				helpers.NormallyDistributedSingle(s.Deviation*s.Size, 0),
				helpers.NormallyDistributedSingle(s.Deviation*s.Size, 0),
			}

			d := pos.Len() / s.Size
			m := d*2000 + (1-d)*1500
			temperature := helpers.Distribution(4000, m, 1000, 40000)
			swilr := (math.Pi * 4) * 5

			newStar := Star{
				Position:    pos,
				Temperature: temperature,
			}

			newStar.GenerateName()
			newStar.ConvertTemperature()
			newStar.Swirl(mgl64.Vec3{0, 1, 0}, swilr)

			stars = append(stars, newStar)
		}
	}

	return
}
