package entity

import (
	"time"

	"github.com/google/uuid"
)

type ImpactActivity struct {
	ID             uuid.UUID  `gorm:"type:char(36);primaryKey"`
	UserID         uuid.UUID  `gorm:"type:char(36);not null;index"`
	ActivityType   string     `gorm:"type:enum('cooking','donation');not null;index"`
	ReferenceID    *uuid.UUID `gorm:"type:char(36)"`
	Title          string     `gorm:"type:varchar(200);not null"`
	PortionsSaved  int        `gorm:"type:int;not null;default:0"`
	MoneySaved     float64    `gorm:"type:decimal(12,2);not null;default:0"`
	CO2PreventedKg float64    `gorm:"type:decimal(6,2);not null;default:0"`
	ItemsBreakdown string     `gorm:"type:json"`
	CreatedAt      time.Time  `gorm:"type:timestamp;autoCreateTime;index"`

	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}
