package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/app/pantry/repository"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
	"github.com/mqqff/wasteco-be/internal/domain/entity"
	"gorm.io/gorm"
)

type fakePantryRepo struct {
	repository.PantryRepoItf
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
		return entity.PantryIngredient{}, gorm.ErrRecordNotFound
	}
	return item, nil
}

func (f *fakePantryRepo) Update(item *entity.PantryIngredient) error {
	f.items[item.ID] = *item
	return nil
}

func (f *fakePantryRepo) Delete(id uuid.UUID, userID uuid.UUID) error {
	item, ok := f.items[id]
	if !ok || item.UserID != userID {
		return gorm.ErrRecordNotFound
	}
	delete(f.items, id)
	return nil
}

func TestGetIngredientsDynamicStatusAndFilter(t *testing.T) {
	userID := uuid.New()
	soon := time.Now().AddDate(0, 0, 2)
	later := time.Now().AddDate(0, 0, 10)
	repo := &fakePantryRepo{items: map[uuid.UUID]entity.PantryIngredient{
		uuid.New(): {ID: uuid.New(), UserID: userID, Name: "Wortel", Quantity: 100, Unit: "gram", Status: "fresh", ExpiryDate: &soon},
		uuid.New(): {ID: uuid.New(), UserID: userID, Name: "Nasi", Quantity: 2, Unit: "porsi", Status: "fresh", ExpiryDate: &later},
	}}
	u := NewPantryUsecase(repo, nil)

	near, err := u.GetIngredients(userID, "", "near_expiry")
	if err != nil {
		t.Fatal(err)
	}
	if len(near) != 1 || near[0].Name != "Wortel" || near[0].Status != "near_expiry" {
		t.Fatalf("near_expiry filter got %+v, want Wortel", near)
	}
}

func TestAddInvalidExpiryAndNotFoundErrors(t *testing.T) {
	userID := uuid.New()
	otherID := uuid.New()
	itemID := uuid.New()
	repo := &fakePantryRepo{items: map[uuid.UUID]entity.PantryIngredient{
		itemID: {ID: itemID, UserID: otherID, Name: "Telur", Quantity: 2, Unit: "butir", Status: "fresh"},
	}}
	u := NewPantryUsecase(repo, nil)

	badDate := "25-11-2026"
	if _, err := u.AddIngredient(userID, dto.CreatePantryItemRequest{
		Name: "Susu", Quantity: 1, Unit: "liter", ExpiryDate: &badDate,
	}); !errors.Is(err, ErrInvalidExpiryDate) {
		t.Fatalf("invalid date err = %v, want ErrInvalidExpiryDate", err)
	}

	if _, err := u.UpdateIngredient(userID, uuid.New(), dto.UpdatePantryItemRequest{Name: "X"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing update err = %v, want ErrNotFound", err)
	}
	if _, err := u.UpdateIngredient(userID, itemID, dto.UpdatePantryItemRequest{Name: "X"}); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("other user update err = %v, want ErrUnauthorized", err)
	}
	if err := u.DeleteIngredient(userID, itemID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("other user delete err = %v, want ErrNotFound", err)
	}
}
