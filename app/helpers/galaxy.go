package helpers

import (
	"bitbucket.org/no-name-game/no-name/app/commands/godeffect/galaxies"
	"bitbucket.org/no-name-game/no-name/app/models"
)

// NewGalaxyChunk - Generate new portion of
func NewGalaxyChunk(deviation int) bool {
	size := 10.0
	sphere := galaxies.Sphere{
		Size:      size,
		Density:   20,
		Deviation: float64(deviation),
	}

	for _, star := range sphere.Generate() {
		newStar := models.Star{
			Name:        star.Name,
			X:           star.Position[0],
			Y:           star.Position[1],
			Z:           star.Position[2],
			Temperature: star.Temperature,
			Color:       star.Color,
		}

		newStar.Create()
	}

	return true
}
