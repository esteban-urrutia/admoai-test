package models

import (
    "time"
    "gorm.io/gorm"
    "gorm.io/plugin/soft_delete"
)

type Ad struct {
    ID        uint                  `gorm:"primaryKey"`
    
	Title     string                `gorm:"not null" json:"title" binding:"required"`
    ImageUrl  string                `gorm:"not null" json:"image_url" binding:"required,url"`
    Placement string                `gorm:"not null" json:"placement" binding:"required"`
    
	Active    soft_delete.DeletedAt `gorm:"softDelete:flag,DeletedAtField:DeletedAt;default:1" json:"active"`
    
	CreatedAt time.Time             `json:"created_at"`
    UpdatedAt time.Time             `json:"updated_at"`
    DeletedAt gorm.DeletedAt        `gorm:"index" json:"deleted_at"`
}