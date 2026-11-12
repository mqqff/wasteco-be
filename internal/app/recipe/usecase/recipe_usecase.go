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
		// Take all available pantry ingredients
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
	var calculatedSavings float64

	for _, used := range req.UsedIngredients {
		calculatedSavings += used.EstimatedPrice

		if used.IngredientID != "" {
			id, err := uuid.Parse(used.IngredientID)
			if err == nil {
				item, err := u.pantryRepo.FindByID(id)
				if err == nil && item.UserID == userID {
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
	}

	totalSavings := req.EstimatedSavings
	if totalSavings <= 0 {
		totalSavings = calculatedSavings
	}

	// 1 portion of food waste prevented ~ 0.21 kg CO2e
	co2Prevented := float64(req.Portions) * 0.21

	breakdownJSON, _ := sonic.Marshal(req.UsedIngredients)

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
		Message:        fmt.Sprintf("Hebat! %d porsi makanan dan Rp%.0f berhasil diselamatkan dari potensi terbuang.", req.Portions, totalSavings),
	}, nil
}
