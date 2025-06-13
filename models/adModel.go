package models

import (
	"time"
)

type Ad struct {
	ID uint `gorm:"primaryKey"`

	Title     string `gorm:"not null" json:"title" binding:"required"`
	ImageUrl  string `gorm:"not null" json:"image_url" binding:"required,url"`
	Placement string `gorm:"not null" json:"placement" binding:"required"`
	Status    string `gorm:"not null" json:"status"`

	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DeactivatedAt time.Time `gorm:"index" json:"deactivated_at"`
}
