package godivers

import (
	"bitbucket.org/no-name-game/no-name/godivers/galaxies"
)

//Galaxy -
type Galaxy struct {
	Stars []Star
}

// GodEffect - GodEffect
func GodEffect() {
	galaxies.Spiral()
}
