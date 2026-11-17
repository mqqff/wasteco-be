package dto

type CreatePantryItemRequest struct {
	Name            string  `json:"name" validate:"required,min=2,max=100"`
	Quantity        float64 `json:"quantity" validate:"required,gt=0"`
	Unit            string  `json:"unit" validate:"required"`
	Category        string  `json:"category"`
	StorageLocation string  `json:"storage_location"`
	ExpiryDate      *string `json:"expiry_date"`
}

type BulkCreatePantryItemRequest struct {
	Items []CreatePantryItemRequest `json:"items" validate:"required,min=1,max=50,dive"`
}

type UpdatePantryItemRequest struct {
	Name            string   `json:"name" validate:"omitempty,min=2,max=100"`
	Quantity        *float64 `json:"quantity" validate:"omitempty,gte=0"`
	Unit            string   `json:"unit"`
	Category        string   `json:"category"`
	StorageLocation string   `json:"storage_location"`
	ExpiryDate      *string  `json:"expiry_date"`
	Status          string   `json:"status" validate:"omitempty,oneof=fresh near_expiry used donated"`
}

type PantryItemResponse struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Quantity        float64 `json:"quantity"`
	Unit            string  `json:"unit"`
	Category        string  `json:"category"`
	StorageLocation string  `json:"storage_location"`
	ExpiryDate      string  `json:"expiry_date,omitempty"`
	Status          string  `json:"status"`
	DaysUntilExpiry *int    `json:"days_until_expiry,omitempty"`
	CreatedAt       string  `json:"created_at"`
}
