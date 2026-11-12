package dto

import "github.com/google/uuid"

type UserProfile struct {
	ID                 string  `json:"id"`
	Email              string  `json:"email"`
	Name               string  `json:"name"`
	PhoneNumber        string  `json:"phone_number"`
	AreaName           string  `json:"area_name"`
	Latitude           float64 `json:"latitude"`
	Longitude          float64 `json:"longitude"`
	SharingRadiusKM    int     `json:"sharing_radius_km"`
	TotalPortionsSaved int     `json:"total_portions_saved"`
	TotalMoneySaved    float64 `json:"total_money_saved"`
}

type UpdateProfileRequest struct {
	Name            string  `json:"name" validate:"omitempty,min=2,max=100"`
	PhoneNumber     string  `json:"phone_number"`
	AreaName        string  `json:"area_name"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	SharingRadiusKM int     `json:"sharing_radius_km" validate:"omitempty,min=1,max=50"`
}

type UserParam struct {
	Email string
	ID    uuid.UUID
}
