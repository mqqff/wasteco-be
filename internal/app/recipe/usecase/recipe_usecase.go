package usecase

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	impactRepo "github.com/mqqff/wasteco-be/internal/app/impact/repository"
	pantryRepo "github.com/mqqff/wasteco-be/internal/app/pantry/repository"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
	"github.com/mqqff/wasteco-be/internal/domain/entity"
	"github.com/mqqff/wasteco-be/internal/infra/gemini"
	"github.com/mqqff/wasteco-be/pkg/unit"
)

var ErrNoIngredients = errors.New("no ingredients provided or found in pantry")

type RecipeUsecaseItf interface {
	GenerateRecipes(userID uuid.UUID, req dto.GenerateRecipeRequest) ([]gemini.GeneratedRecipe, error)
	GenerateRecipesFromImage(userID uuid.UUID, image []byte, mimeType string, req dto.GenerateRecipeRequest) (gemini.ImageRecipeResponse, error)
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
	ingredients, err := u.collectIngredients(userID, req, true)
	if err != nil {
		return nil, err
	}
	if len(ingredients) == 0 {
		return nil, ErrNoIngredients
	}

	return u.gemini.GenerateRecipes(ingredients)
}

func (u *RecipeUsecase) GenerateRecipesFromImage(userID uuid.UUID, image []byte, mimeType string, req dto.GenerateRecipeRequest) (gemini.ImageRecipeResponse, error) {
	extra, err := u.collectIngredients(userID, req, false)
	if err != nil {
		return gemini.ImageRecipeResponse{}, err
	}

	res, err := u.gemini.GenerateRecipesFromImage(image, mimeType, extra)
	if err != nil {
		return gemini.ImageRecipeResponse{}, err
	}
	if len(res.DetectedIngredients) == 0 && len(res.Recipes) == 0 {
		return gemini.ImageRecipeResponse{}, ErrNoIngredients
	}

	return res, nil
}

func (u *RecipeUsecase) collectIngredients(userID uuid.UUID, req dto.GenerateRecipeRequest, fallbackToPantry bool) ([]gemini.IngredientInput, error) {
	var ingredients []gemini.IngredientInput

	if len(req.IngredientIDs) > 0 {
		for _, idStr := range req.IngredientIDs {
			id, err := uuid.Parse(idStr)
			if err != nil {
				continue
			}
			item, err := u.pantryRepo.FindByID(id)
			if err == nil && item.UserID == userID && isActivePantryStatus(item.Status) {
				ingredients = append(ingredients, gemini.IngredientInput{
					Name:     item.Name,
					Quantity: item.Quantity,
					Unit:     item.Unit,
				})
			}
		}
	} else if fallbackToPantry && len(req.CustomIngredients) == 0 {
		items, err := u.pantryRepo.FindAllByUserID(userID, "")
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			ingredients = append(ingredients, gemini.IngredientInput{
				Name:     item.Name,
				Quantity: item.Quantity,
				Unit:     item.Unit,
			})
		}
	}

	for _, custom := range req.CustomIngredients {
		trimmed := strings.TrimSpace(custom)
		if trimmed != "" {
			ingredients = append(ingredients, gemini.IngredientInput{
				Name:     trimmed,
				Quantity: 1,
				Unit:     "secukupnya",
			})
		}
	}

	return ingredients, nil
}

func (u *RecipeUsecase) ConfirmCook(userID uuid.UUID, req dto.CookConfirmRequest) (dto.CookConfirmResponse, error) {
	breakdown := make([]dto.ImpactItemBreakdown, 0, len(req.UsedIngredients))
	deductions := make([]impactRepo.PantryDeduction, 0, len(req.UsedIngredients))

	for _, used := range req.UsedIngredients {
		if used.IngredientID != "" {
			id, err := uuid.Parse(used.IngredientID)
			if err == nil {
				item, err := u.pantryRepo.FindByID(id)
				if err == nil && item.UserID == userID && isActivePantryStatus(item.Status) {
					if deductQty, ok := unit.ConvertQuantity(used.Quantity, used.Unit, item.Unit); ok {
						deductions = append(deductions, impactRepo.PantryDeduction{
							PantryIngredientID: item.ID,
							Quantity:           deductQty,
						})
					}
				}
			}
		}

		breakdown = append(breakdown, dto.ImpactItemBreakdown{
			Name:     used.Name,
			Quantity: used.Quantity,
			Unit:     used.Unit,
		})
	}

	co2Prevented := math.Round(float64(req.Portions)*0.21*100) / 100

	breakdownJSON, _ := sonic.Marshal(breakdown)

	activityID := uuid.New()
	activity := entity.ImpactActivity{
		ID:             activityID,
		UserID:         userID,
		ActivityType:   "cooking",
		Title:          req.RecipeName,
		PortionsSaved:  req.Portions,
		CO2PreventedKg: co2Prevented,
		ItemsBreakdown: string(breakdownJSON),
	}

	err := u.impactRepo.RecordCooking(userID, deductions, &activity)
	if err != nil {
		return dto.CookConfirmResponse{}, err
	}

	return dto.CookConfirmResponse{
		ActivityID:     activityID.String(),
		RecipeName:     req.RecipeName,
		PortionsSaved:  req.Portions,
		CO2PreventedKg: co2Prevented,
		Message:        fmt.Sprintf("Great job! %d food portions successfully prevented from being wasted.", req.Portions),
	}, nil
}

func isActivePantryStatus(status string) bool {
	return status == "fresh" || status == "near_expiry"
}
