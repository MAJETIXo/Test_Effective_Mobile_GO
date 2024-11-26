package models

type Group struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"type:text;not null"`
}
