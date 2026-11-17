package dto

type ImpactItemBreakdown struct {
	Name           string  `json:"name"`
	Quantity       float64 `json:"quantity"`
	Unit           string  `json:"unit"`
	EstimatedPrice float64 `json:"estimated_price"`
	Subtotal       float64 `json:"subtotal"`
}

type ImpactActivityResponse struct {
	ID             string                `json:"id"`
	ActivityType   string                `json:"activity_type"`
	Title          string                `json:"title"`
	PortionsSaved  int                   `json:"portions_saved"`
	MoneySaved     float64               `json:"money_saved"`
	CO2PreventedKg float64               `json:"co2_prevented_kg"`
	ItemsBreakdown []ImpactItemBreakdown `json:"items_breakdown,omitempty"`
	FormattedTime  string                `json:"formatted_time"`
	CreatedAt      string                `json:"created_at"`
}

type ImpactComposition struct {
	CookingPortions    int     `json:"cooking_portions"`
	CookingMoneySaved  float64 `json:"cooking_money_saved"`
	CookingPercentage  float64 `json:"cooking_percentage"`
	DonationPortions   int     `json:"donation_portions"`
	DonationMoneySaved float64 `json:"donation_money_saved"`
	DonationPercentage float64 `json:"donation_percentage"`
}

type ImpactSummaryResponse struct {
	Month               int                      `json:"month"`
	Year                int                      `json:"year"`
	MonthName           string                   `json:"month_name"`
	TotalPortionsSaved  int                      `json:"total_portions_saved"`
	TotalMoneySaved     float64                  `json:"total_money_saved"`
	TotalCO2PreventedKg float64                  `json:"total_co2_prevented_kg"`
	Composition         ImpactComposition        `json:"composition"`
	RecentActivities    []ImpactActivityResponse `json:"recent_activities"`
}

type ImpactHistoryResponse struct {
	Page       int                      `json:"page"`
	Limit      int                      `json:"limit"`
	TotalItems int64                    `json:"total_items"`
	TotalPages int                      `json:"total_pages"`
	Activities []ImpactActivityResponse `json:"activities"`
}
