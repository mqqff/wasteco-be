package entity

import (
	"github.com/google/uuid"
)

type User struct {
	ID    uuid.UUID `gorm:"type:char(36);primaryKey"`
	Email string    `gorm:"type:varchar(100);not null;index:unique"`
	Name  string    `gorm:"type:varchar(255);not null"`
}
