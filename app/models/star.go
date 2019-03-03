package models

import (
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// Star - Star struct
type Star struct {
	gorm.Model
	Name        string
	Size        float64
	X           float64
	Y           float64
	Z           float64
	Temperature float64
	Color       string
}

// Stars - Star slice
type Stars []Star

// Create star
func (s *Star) Create() *Star {
	services.Database.Create(&s)

	return s
}
