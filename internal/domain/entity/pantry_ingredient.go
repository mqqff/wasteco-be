package entity

import (
	"time"

	"github.com/google/uuid"
)

type PantryIngredient struct {
	ID              uuid.UUID  `gorm:"type:char(36);primaryKey"`
	UserID          uuid.UUID  `gorm:"type:char(36);not null;index"`
	Name            string     `gorm:"type:varchar(150);not null"`
	Quantity        float64    `gorm:"type:decimal(10,2);not null;default:0"`
	Unit            string     `gorm:"type:varchar(50);not null"`
	Category        string     `gorm:"type:varchar(50);not null"`
	StorageLocation string     `gorm:"type:varchar(100)"`
	ExpiryDate      *time.Time `gorm:"type:date"`
	Status          string     `gorm:"type:enum('fresh','near_expiry','used','donated');default:'fresh';not null"`
	EstimatedPrice  float64    `gorm:"type:decimal(12,2);default:0"`
	CreatedAt       time.Time  `gorm:"type:timestamp;autoCreateTime"`
	UpdatedAt       time.Time  `gorm:"type:timestamp;autoUpdateTime"`

	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}
