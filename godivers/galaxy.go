package godivers

import (
	"bitbucket.org/no-name-game/no-name/godivers/galaxies"
)

//Galaxy -
type Galaxy struct {
	Stars []Star
}

//Star -
type Star struct {
	Name        string
	Size        float32
	Position    []float32
	Temperature float32
}

// GodEffect - GodEffect
func GodEffect() {
	galaxies.Spiral()
}
