package helpers

import (
	"bitbucket.org/no-name-game/no-name/app/acme/namer"
	"bitbucket.org/no-name-game/no-name/app/models"
)

// NewEnemy - Generate enemy
func NewEnemy() (enemy models.Enemy) {
	enemy = models.Enemy{
		Name: namer.GenerateName("resources/namer/enemies/model.json"),
		Life: 100,
	}

	enemy.Create()

	return
}
