package usecase

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/app/impact/repository"
	pantryRepo "github.com/mqqff/wasteco-be/internal/app/pantry/repository"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
	"github.com/mqqff/wasteco-be/internal/domain/entity"
)

type fakePantryRepo struct {
	pantryRepo.PantryRepoItf
	items map[uuid.UUID]entity.PantryIngredient
}

func (f *fakePantryRepo) FindByID(id uuid.UUID) (entity.PantryIngredient, error) {
	return f.items[id], nil
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
