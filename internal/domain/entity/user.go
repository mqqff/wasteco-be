package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID `gorm:"type:char(36);primaryKey"`
	Email           string    `gorm:"type:varchar(100);not null;index:unique"`
	Name            string    `gorm:"type:varchar(255);not null"`
	PasswordHash    string    `gorm:"type:varchar(255)"`
	PhoneNumber     string    `gorm:"type:varchar(30)"`
	AreaName        string    `gorm:"type:varchar(150)"`
	Latitude        float64   `gorm:"type:decimal(10,8);default:0"`
	Longitude       float64   `gorm:"type:decimal(11,8);default:0"`
	SharingRadiusKM int       `gorm:"type:int;default:2"`
	CreatedAt       time.Time `gorm:"type:timestamp;autoCreateTime"`
	UpdatedAt       time.Time `gorm:"type:timestamp;autoUpdateTime"`
}
