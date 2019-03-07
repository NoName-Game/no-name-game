package helpers

import (
	"bitbucket.org/no-name-game/no-name/app/commands/godeffect/galaxies"
	"bitbucket.org/no-name-game/no-name/app/models"
)

// NewGalaxyChunk - Generate new portion of galaxy
func NewGalaxyChunk(deviation int) models.Star {
	size := 10.0
	sphere := galaxies.Sphere{
		Size:      size,
		Density:   10,
		Deviation: float64(deviation),
	}

	var newStar models.Star
	for _, star := range sphere.Generate() {
		newStar = models.Star{
			Name:        star.Name,
			X:           star.Position[0],
			Y:           star.Position[1],
			Z:           star.Position[2],
			Temperature: star.Temperature,
			Color:       star.Color,
		}

		newStar.Create()
	}

	// Return last star
	return newStar
}
