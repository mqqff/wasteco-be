package entity

import (
	"time"

	"github.com/google/uuid"
)

type DonationRequest struct {
	ID           uuid.UUID  `gorm:"type:char(36);primaryKey"`
	RequesterID  uuid.UUID  `gorm:"type:char(36);not null;index"`
	DonorID      *uuid.UUID `gorm:"type:char(36);index"`
	Title        string     `gorm:"type:varchar(150);not null"`
	Note         string     `gorm:"type:text"`
	AreaName     string     `gorm:"type:varchar(150)"`
	MeetingPoint string     `gorm:"type:varchar(255);not null"`
	Latitude     float64    `gorm:"type:decimal(10,8);not null"`
	Longitude    float64    `gorm:"type:decimal(11,8);not null"`
	ReceiveStart time.Time  `gorm:"type:datetime;not null"`
	ReceiveEnd   time.Time  `gorm:"type:datetime;not null"`
	Status       string     `gorm:"type:enum('open','handed_over','received','cancelled');default:'open';not null;index"`
	HandedOverAt *time.Time `gorm:"type:datetime"`
	ReceivedAt   *time.Time `gorm:"type:datetime"`
	CreatedAt    time.Time  `gorm:"type:timestamp;autoCreateTime"`
	UpdatedAt    time.Time  `gorm:"type:timestamp;autoUpdateTime"`

	Requester User                  `gorm:"foreignKey:RequesterID;references:ID;constraint:OnDelete:CASCADE"`
	Donor     *User                 `gorm:"foreignKey:DonorID;references:ID;constraint:OnDelete:SET NULL"`
	Items     []DonationRequestItem `gorm:"foreignKey:RequestID;references:ID;constraint:OnDelete:CASCADE"`
}

type DonationRequestItem struct {
	ID                 uuid.UUID  `gorm:"type:char(36);primaryKey"`
	RequestID          uuid.UUID  `gorm:"type:char(36);not null;index"`
	Name               string     `gorm:"type:varchar(150);not null"`
	Quantity           float64    `gorm:"type:decimal(10,2);not null"`
	Unit               string     `gorm:"type:varchar(50);not null"`
	Selected           bool       `gorm:"not null;default:false"`
	PantryIngredientID *uuid.UUID `gorm:"type:char(36)"`
	UnitPrice          float64    `gorm:"type:decimal(12,2);not null;default:0"`
}
