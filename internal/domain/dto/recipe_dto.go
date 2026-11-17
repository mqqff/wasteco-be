package dto

type GenerateRecipeRequest struct {
	IngredientIDs     []string `json:"ingredient_ids"`
	CustomIngredients []string `json:"custom_ingredients"`
}

type UsedIngredientInput struct {
	IngredientID string  `json:"ingredient_id"`
	Name         string  `json:"name" validate:"required"`
	Quantity     float64 `json:"quantity" validate:"required,gt=0"`
	Unit         string  `json:"unit" validate:"required"`
}

type CookConfirmRequest struct {
	RecipeName      string                `json:"recipe_name" validate:"required"`
	Portions        int                   `json:"portions" validate:"required,min=1"`
	UsedIngredients []UsedIngredientInput `json:"used_ingredients" validate:"required,min=1"`
}

type CookConfirmResponse struct {
	ActivityID     string  `json:"activity_id"`
	RecipeName     string  `json:"recipe_name"`
	PortionsSaved  int     `json:"portions_saved"`
	CO2PreventedKg float64 `json:"co2_prevented_kg"`
	Message        string  `json:"message"`
}
