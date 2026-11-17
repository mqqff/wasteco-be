package usecase

import (
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

func TestConfirmCookSavingsUseUnitPrice(t *testing.T) {
	userID := uuid.New()
	riceID := uuid.New()
	pantry := &fakePantryRepo{items: map[uuid.UUID]entity.PantryIngredient{
		riceID: {ID: riceID, UserID: userID, Quantity: 3, EstimatedPrice: 2500},
	}}
	impact := &fakeImpactRepo{}
	u := NewRecipeUsecase(pantry, impact, nil)

	res, err := u.ConfirmCook(userID, dto.CookConfirmRequest{
		RecipeName: "Nasi Goreng Wortel Telur",
		Portions:   2,
		UsedIngredients: []dto.UsedIngredientInput{
			{IngredientID: riceID.String(), Name: "Nasi Putih", Quantity: 2, Unit: "porsi", EstimatedPrice: 999},
			{Name: "Telur Ayam", Quantity: 2, Unit: "butir", EstimatedPrice: 2000},
			{Name: "Wortel", Quantity: 100, Unit: "gram", EstimatedPrice: 30},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.MoneySaved != 12000 {
		t.Fatalf("money saved = %v, want 12000", res.MoneySaved)
	}
	if impact.created.MoneySaved != 12000 {
		t.Fatalf("logged money saved = %v, want 12000", impact.created.MoneySaved)
	}
	if got := pantry.items[riceID].Quantity; got != 1 {
		t.Fatalf("rice left = %v, want 1", got)
	}
}
