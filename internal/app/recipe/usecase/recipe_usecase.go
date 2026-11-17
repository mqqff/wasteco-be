package usecase

import (
	"errors"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	impactRepo "github.com/mqqff/wasteco-be/internal/app/impact/repository"
	pantryRepo "github.com/mqqff/wasteco-be/internal/app/pantry/repository"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
	"github.com/mqqff/wasteco-be/internal/domain/entity"
	"github.com/mqqff/wasteco-be/internal/infra/gemini"
)

type RecipeUsecaseItf interface {
	GenerateRecipes(userID uuid.UUID, req dto.GenerateRecipeRequest) ([]gemini.GeneratedRecipe, error)
	ConfirmCook(userID uuid.UUID, req dto.CookConfirmRequest) (dto.CookConfirmResponse, error)
}

type RecipeUsecase struct {
	pantryRepo pantryRepo.PantryRepoItf
	impactRepo impactRepo.ImpactRepoItf
	gemini     gemini.GeminiItf
}

func NewRecipeUsecase(p pantryRepo.PantryRepoItf, i impactRepo.ImpactRepoItf, g gemini.GeminiItf) RecipeUsecaseItf {
	return &RecipeUsecase{
		pantryRepo: p,
		impactRepo: i,
		gemini:     g,
	}
}

func (u *RecipeUsecase) GenerateRecipes(userID uuid.UUID, req dto.GenerateRecipeRequest) ([]gemini.GeneratedRecipe, error) {
	var ingredients []gemini.IngredientInput

	if len(req.IngredientIDs) > 0 {
		for _, idStr := range req.IngredientIDs {
			id, err := uuid.Parse(idStr)
			if err != nil {
				continue
			}
			item, err := u.pantryRepo.FindByID(id)
			if err == nil && item.UserID == userID {
				ingredients = append(ingredients, gemini.IngredientInput{
					Name:     item.Name,
					Quantity: item.Quantity,
					Unit:     item.Unit,
				})
			}
		}
	} else if len(req.CustomIngredients) == 0 {
		items, err := u.pantryRepo.FindAllByUserID(userID, "")
		if err == nil {
			for _, item := range items {
				ingredients = append(ingredients, gemini.IngredientInput{
					Name:     item.Name,
					Quantity: item.Quantity,
					Unit:     item.Unit,
				})
			}
		}
	}

	for _, custom := range req.CustomIngredients {
		if custom != "" {
			ingredients = append(ingredients, gemini.IngredientInput{
				Name:     custom,
				Quantity: 1,
				Unit:     "secukupnya",
			})
		}
	}

	if len(ingredients) == 0 {
		return nil, errors.New("no ingredients provided or found in pantry")
	}

	return u.gemini.GenerateRecipes(ingredients)
}

func (u *RecipeUsecase) ConfirmCook(userID uuid.UUID, req dto.CookConfirmRequest) (dto.CookConfirmResponse, error) {
	var totalSavings float64
	breakdown := make([]dto.ImpactItemBreakdown, 0, len(req.UsedIngredients))

	for _, used := range req.UsedIngredients {
		unitPrice := used.EstimatedPrice

		if used.IngredientID != "" {
			id, err := uuid.Parse(used.IngredientID)
			if err == nil {
				item, err := u.pantryRepo.FindByID(id)
				if err == nil && item.UserID == userID {
					if item.EstimatedPrice > 0 {
						unitPrice = item.EstimatedPrice
					}
					if item.Quantity <= used.Quantity {
						item.Quantity = 0
						item.Status = "used"
					} else {
						item.Quantity -= used.Quantity
					}
					_ = u.pantryRepo.Update(&item)
				}
			}
		}

		subtotal := used.Quantity * unitPrice
		totalSavings += subtotal
		breakdown = append(breakdown, dto.ImpactItemBreakdown{
			Name:           used.Name,
			Quantity:       used.Quantity,
			Unit:           used.Unit,
			EstimatedPrice: unitPrice,
			Subtotal:       subtotal,
		})
	}

	co2Prevented := float64(req.Portions) * 0.21

	breakdownJSON, _ := sonic.Marshal(breakdown)

	activityID := uuid.New()
	activity := entity.ImpactActivity{
		ID:             activityID,
		UserID:         userID,
		ActivityType:   "cooking",
		Title:          req.RecipeName,
		PortionsSaved:  req.Portions,
		MoneySaved:     totalSavings,
		CO2PreventedKg: co2Prevented,
		ItemsBreakdown: string(breakdownJSON),
	}

	err := u.impactRepo.Create(&activity)
	if err != nil {
		return dto.CookConfirmResponse{}, err
	}

	return dto.CookConfirmResponse{
		ActivityID:     activityID.String(),
		RecipeName:     req.RecipeName,
		PortionsSaved:  req.Portions,
		MoneySaved:     totalSavings,
		CO2PreventedKg: co2Prevented,
		Message:        fmt.Sprintf("Great job! %d food portions and Rp%.0f successfully prevented from being wasted.", req.Portions, totalSavings),
	}, nil
}
