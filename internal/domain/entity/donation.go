package entity

import (
	"time"

	"github.com/google/uuid"
)

type Donation struct {
	ID                uuid.UUID  `gorm:"type:char(36);primaryKey"`
	UserID            uuid.UUID  `gorm:"type:char(36);not null;index"`
	Title             string     `gorm:"type:varchar(150);not null"`
	Category          string     `gorm:"type:varchar(50);not null"`
	Quantity          float64    `gorm:"type:decimal(10,2);not null;default:0"`
	Unit              string     `gorm:"type:varchar(50);not null"`
	PhotoURLs         string     `gorm:"type:json"`
	ConditionNote     string     `gorm:"type:text;not null"`
	StorageMethod     string     `gorm:"type:varchar(100)"`
	ExpiryDate        *time.Time `gorm:"type:date"`
	PickupDeadline    *time.Time `gorm:"type:datetime"`
	AreaName          string     `gorm:"type:varchar(150);not null"`
	MeetingPoint      string     `gorm:"type:text;not null"`
	Latitude          float64    `gorm:"type:decimal(10,8);default:0"`
	Longitude         float64    `gorm:"type:decimal(11,8);default:0"`
	WhatsAppNumber    string     `gorm:"column:whatsapp_number;type:varchar(30);not null"`
	EstimatedPrice    float64    `gorm:"type:decimal(12,2);default:0"`
	EstimatedPortions int        `gorm:"type:int;default:0"`
	Status            string     `gorm:"type:enum('available','in_progress','completed','cancelled');default:'available';not null;index"`
	CreatedAt         time.Time  `gorm:"type:timestamp;autoCreateTime"`
	UpdatedAt         time.Time  `gorm:"type:timestamp;autoUpdateTime"`

	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}
