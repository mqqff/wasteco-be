package usecase

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/app/impact/repository"
	pantryRepo "github.com/mqqff/wasteco-be/internal/app/pantry/repository"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
	"github.com/mqqff/wasteco-be/internal/domain/entity"
	"github.com/mqqff/wasteco-be/internal/infra/gemini"
)

type fakePantryRepo struct {
	pantryRepo.PantryRepoItf
	items map[uuid.UUID]entity.PantryIngredient
}

func (f *fakePantryRepo) FindAllByUserID(userID uuid.UUID, _ string) ([]entity.PantryIngredient, error) {
	var res []entity.PantryIngredient
	for _, item := range f.items {
		if item.UserID == userID && (item.Status == "fresh" || item.Status == "near_expiry") {
			res = append(res, item)
		}
	}
	return res, nil
}

func (f *fakePantryRepo) FindByID(id uuid.UUID) (entity.PantryIngredient, error) {
	item, ok := f.items[id]
	if !ok {
		return entity.PantryIngredient{}, errors.New("not found")
	}
	return item, nil
}

func (f *fakePantryRepo) Update(item *entity.PantryIngredient) error {
	f.items[item.ID] = *item
	return nil
}

type fakeImpactRepo struct {
	repository.ImpactRepoItf
	created *entity.ImpactActivity
}

func (f *fakeImpactRepo) Create(a *entity.ImpactActivity) error {
	f.created = a
	return nil
}

type fakeGemini struct {
	gemini.GeminiItf
	lastIngredients []gemini.IngredientInput
	recipes         []gemini.GeneratedRecipe
}

func (f *fakeGemini) GenerateRecipes(ingredients []gemini.IngredientInput) ([]gemini.GeneratedRecipe, error) {
	f.lastIngredients = ingredients
	return f.recipes, nil
}

func TestConfirmCookDeductsPantryAndLogsPortions(t *testing.T) {
	userID := uuid.New()
	riceID := uuid.New()
	eggID := uuid.New()
	pantry := &fakePantryRepo{items: map[uuid.UUID]entity.PantryIngredient{
		riceID: {ID: riceID, UserID: userID, Quantity: 3, Status: "fresh"},
		eggID:  {ID: eggID, UserID: userID, Quantity: 2, Status: "fresh"},
	}}
	impact := &fakeImpactRepo{}
	u := NewRecipeUsecase(pantry, impact, nil)

	res, err := u.ConfirmCook(userID, dto.CookConfirmRequest{
		RecipeName: "Nasi Goreng Wortel Telur",
		Portions:   2,
		UsedIngredients: []dto.UsedIngredientInput{
			{IngredientID: riceID.String(), Name: "Nasi Putih", Quantity: 2, Unit: "porsi"},
			{IngredientID: eggID.String(), Name: "Telur Ayam", Quantity: 2, Unit: "butir"},
			{Name: "Wortel", Quantity: 100, Unit: "gram"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.PortionsSaved != 2 || impact.created.PortionsSaved != 2 {
		t.Fatalf("portions = %d/%d, want 2", res.PortionsSaved, impact.created.PortionsSaved)
	}
	if got := pantry.items[riceID].Quantity; got != 1 {
		t.Fatalf("rice left = %v, want 1", got)
	}
	if got := pantry.items[eggID]; got.Quantity != 0 || got.Status != "used" {
		t.Fatalf("egg = %v/%s, want 0/used", got.Quantity, got.Status)
	}
	if !strings.Contains(impact.created.ItemsBreakdown, `"name":"Wortel"`) || strings.Contains(impact.created.ItemsBreakdown, "price") {
		t.Fatalf("unexpected breakdown: %s", impact.created.ItemsBreakdown)
	}
}

func TestConfirmCookUnitConversionAndSkipsInactive(t *testing.T) {
	userID := uuid.New()
	carrotID := uuid.New()
	usedID := uuid.New()
	spinachID := uuid.New()
	pantry := &fakePantryRepo{items: map[uuid.UUID]entity.PantryIngredient{
		carrotID:  {ID: carrotID, UserID: userID, Name: "Wortel", Quantity: 1, Unit: "kg", Status: "fresh"},
		usedID:    {ID: usedID, UserID: userID, Name: "Nasi", Quantity: 0, Unit: "porsi", Status: "used"},
		spinachID: {ID: spinachID, UserID: userID, Name: "Bayam", Quantity: 2, Unit: "ikat", Status: "near_expiry"},
	}}
	impact := &fakeImpactRepo{}
	u := NewRecipeUsecase(pantry, impact, nil)

	res, err := u.ConfirmCook(userID, dto.CookConfirmRequest{
		RecipeName: "Tumis Wortel Bayam",
		Portions:   3,
		UsedIngredients: []dto.UsedIngredientInput{
			{IngredientID: carrotID.String(), Name: "Wortel", Quantity: 500, Unit: "gram"},
			{IngredientID: usedID.String(), Name: "Nasi", Quantity: 1, Unit: "porsi"},
			{IngredientID: spinachID.String(), Name: "Bayam", Quantity: 100, Unit: "gram"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.CO2PreventedKg != 0.63 {
		t.Fatalf("CO2PreventedKg = %v, want 0.63", res.CO2PreventedKg)
	}
	if got := pantry.items[carrotID].Quantity; got != 0.5 {
		t.Fatalf("carrot left = %v, want 0.5", got)
	}
	if got := pantry.items[usedID].Status; got != "used" {
		t.Fatalf("used item status = %s, want used", got)
	}
	if got := pantry.items[spinachID].Quantity; got != 2 {
		t.Fatalf("incompatible unit modified spinach quantity to %v, want 2", got)
	}
}

func TestGenerateRecipesActiveFilterAndEmptyError(t *testing.T) {
	userID := uuid.New()
	freshID := uuid.New()
	usedID := uuid.New()
	pantry := &fakePantryRepo{items: map[uuid.UUID]entity.PantryIngredient{
		freshID: {ID: freshID, UserID: userID, Name: "Telur", Quantity: 2, Unit: "butir", Status: "fresh"},
		usedID:  {ID: usedID, UserID: userID, Name: "Nasi", Quantity: 0, Unit: "porsi", Status: "used"},
	}}
	ai := &fakeGemini{recipes: []gemini.GeneratedRecipe{{RecipeName: "Orak-Arik Telur"}}}
	u := NewRecipeUsecase(pantry, &fakeImpactRepo{}, ai)

	if _, err := u.GenerateRecipes(userID, dto.GenerateRecipeRequest{
		IngredientIDs: []string{usedID.String()},
	}); !errors.Is(err, ErrNoIngredients) {
		t.Fatalf("used-only ingredient IDs: err = %v, want ErrNoIngredients", err)
	}

	recipes, err := u.GenerateRecipes(userID, dto.GenerateRecipeRequest{
		IngredientIDs:     []string{freshID.String(), usedID.String()},
		CustomIngredients: []string{"Bawang Putih"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(recipes) != 1 || len(ai.lastIngredients) != 2 {
		t.Fatalf("recipes=%d ingredients=%d, want 1 and 2", len(recipes), len(ai.lastIngredients))
	}
}
