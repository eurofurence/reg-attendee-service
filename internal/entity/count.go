package entity

import (
	"time"
)

type Count struct {
	Area      string `gorm:"primaryKey"`
	Name      string `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Pending   int `gorm:"NOT NULL"`
	Attending int `gorm:"NOT NULL"`
}

const (
	CountAreaPackage = "package"
)
