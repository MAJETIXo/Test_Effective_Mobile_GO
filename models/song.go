package models

import (
	"time"
)
type Song struct {
	ID uint `gorm:"primaryKey"`
	Name string `gorm:"type:text;not null"`
	ReleaseDate time.Time `gorm:"type:date;not null"`
	Text string `gorm:"type:text;not null"`
	GroupID uint `gorm:"not null"`
	Group Group `gorm:"foreignKey:GroupID"`
}
