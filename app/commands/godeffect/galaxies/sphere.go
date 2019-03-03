package galaxies

import (
	"bitbucket.org/no-name-game/no-name/app/commands/godeffect/helpers"
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

			dimension := pos.Len() / s.Size
			mass := dimension*2000 + (1-dimension)*1500
			temperature := helpers.Distribution(4000, mass, 1000, 40000)

			stars = append(stars, Star{
				Name:        helpers.GenerateStarName(),
				Size:        s.Size,
				Position:    pos,
				Temperature: temperature,
			})
		}
	}

	return
}
