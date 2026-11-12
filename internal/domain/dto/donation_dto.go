package dto

type CreateDonationRequest struct {
	Title             string  `form:"title" json:"title" validate:"required,min=3,max=150"`
	Category          string  `form:"category" json:"category" validate:"required"`
	Quantity          float64 `form:"quantity" json:"quantity" validate:"required,gt=0"`
	Unit              string  `form:"unit" json:"unit" validate:"required"`
	ConditionNote     string  `form:"condition_note" json:"condition_note" validate:"required"`
	StorageMethod     string  `form:"storage_method" json:"storage_method"`
	ExpiryDate        *string `form:"expiry_date" json:"expiry_date"`
	PickupDeadline    string  `form:"pickup_deadline" json:"pickup_deadline" validate:"required"`
	AreaName          string  `form:"area_name" json:"area_name" validate:"required"`
	MeetingPoint      string  `form:"meeting_point" json:"meeting_point" validate:"required"`
	Latitude          float64 `form:"latitude" json:"latitude"`
	Longitude         float64 `form:"longitude" json:"longitude"`
	WhatsAppNumber    string  `form:"whatsapp_number" json:"whatsapp_number" validate:"required"`
	EstimatedPrice    float64 `form:"estimated_price" json:"estimated_price"`
	EstimatedPortions int     `form:"estimated_portions" json:"estimated_portions"`
}

type UpdateDonationStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=available in_progress completed cancelled"`
}

type DonorInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	AreaName    string `json:"area_name"`
	PhoneNumber string `json:"phone_number"`
}

type DonationResponse struct {
	ID                string     `json:"id"`
	UserID            string     `json:"user_id"`
	Donor             *DonorInfo `json:"donor,omitempty"`
	Title             string     `json:"title"`
	Category          string     `json:"category"`
	Quantity          float64    `json:"quantity"`
	Unit              string     `json:"unit"`
	PhotoURLs         []string   `json:"photo_urls"`
	ConditionNote     string     `json:"condition_note"`
	StorageMethod     string     `json:"storage_method"`
	ExpiryDate        string     `json:"expiry_date,omitempty"`
	PickupDeadline    string     `json:"pickup_deadline"`
	AreaName          string     `json:"area_name"`
	MeetingPoint      string     `json:"meeting_point"`
	Latitude          float64    `json:"latitude"`
	Longitude         float64    `json:"longitude"`
	WhatsAppNumber    string     `json:"whatsapp_number"`
	EstimatedPrice    float64    `json:"estimated_price"`
	EstimatedPortions int        `json:"estimated_portions"`
	Status            string     `json:"status"`
	DistanceKM        *float64   `json:"distance_km,omitempty"`
	FormattedDistance string     `json:"formatted_distance,omitempty"`
	CreatedAt         string     `json:"created_at"`
}

type MyDonationsResponse struct {
	Active  []DonationResponse `json:"active"`
	History []DonationResponse `json:"history"`
}
