package dto

import "time"

type DonationItemInput struct {
	Name     string  `json:"name" validate:"required,min=2,max=150"`
	Quantity float64 `json:"quantity" validate:"required,gt=0"`
	Unit     string  `json:"unit" validate:"required,max=50"`
}

type CreateDonationRequest struct {
	Title        string              `json:"title" validate:"omitempty,max=150"`
	Note         string              `json:"note" validate:"omitempty,max=500"`
	Items        []DonationItemInput `json:"items" validate:"required,min=1,max=20,dive"`
	AreaName     string              `json:"area_name" validate:"omitempty,max=150"`
	MeetingPoint string              `json:"meeting_point" validate:"required,max=255"`
	Latitude     float64             `json:"latitude" validate:"required,latitude"`
	Longitude    float64             `json:"longitude" validate:"required,longitude"`
	ReceiveStart time.Time           `json:"receive_start" validate:"required"`
	ReceiveEnd   time.Time           `json:"receive_end" validate:"required,gtfield=ReceiveStart,gt"`
}

type FulfillDonationRequest struct {
	ItemIDs []string `json:"item_ids" validate:"required,min=1,dive,uuid"`
}

type DonationUser struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	AreaName       string `json:"area_name"`
	WhatsAppNumber string `json:"whatsapp_number,omitempty"`
	WhatsAppURL    string `json:"whatsapp_url,omitempty"`
}

type DonationItemResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Quantity  float64 `json:"quantity"`
	Unit      string  `json:"unit"`
	Selected  bool    `json:"selected"`
	UnitPrice float64 `json:"unit_price"`
}

type DonationResponse struct {
	ID                string                 `json:"id"`
	Title             string                 `json:"title"`
	Note              string                 `json:"note"`
	AreaName          string                 `json:"area_name"`
	MeetingPoint      string                 `json:"meeting_point"`
	Latitude          float64                `json:"latitude"`
	Longitude         float64                `json:"longitude"`
	ReceiveStart      string                 `json:"receive_start"`
	ReceiveEnd        string                 `json:"receive_end"`
	Status            string                 `json:"status"`
	ItemCount         int                    `json:"item_count"`
	Items             []DonationItemResponse `json:"items"`
	Requester         *DonationUser          `json:"requester,omitempty"`
	Donor             *DonationUser          `json:"donor,omitempty"`
	DistanceKM        *float64               `json:"distance_km,omitempty"`
	FormattedDistance string                 `json:"formatted_distance,omitempty"`
	HandedOverAt      string                 `json:"handed_over_at,omitempty"`
	ReceivedAt        string                 `json:"received_at,omitempty"`
	CreatedAt         string                 `json:"created_at"`
}

type MyDonationsResponse struct {
	Active  []DonationResponse `json:"active"`
	History []DonationResponse `json:"history"`
}
