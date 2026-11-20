package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	donationRepo "github.com/mqqff/wasteco-be/internal/app/donation/repository"
	pantryRepo "github.com/mqqff/wasteco-be/internal/app/pantry/repository"
	userRepo "github.com/mqqff/wasteco-be/internal/app/user/repository"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
	"github.com/mqqff/wasteco-be/internal/domain/entity"
	"github.com/mqqff/wasteco-be/pkg/unit"
)

type fakeDonationRepo struct {
	donationRepo.DonationRepoItf
	request    entity.DonationRequest
	activities []entity.ImpactActivity
	deductions []donationRepo.PantryDeduction
}

func (f *fakeDonationRepo) FindByID(uuid.UUID) (entity.DonationRequest, error) {
	return f.request, nil
}

func (f *fakeDonationRepo) Fulfill(_, donorID uuid.UUID, items []entity.DonationRequestItem) error {
	if f.request.Status != "open" {
		return donationRepo.ErrStateChanged
	}
	f.request.Status = "handed_over"
	f.request.DonorID = &donorID
	for _, sel := range items {
		for i := range f.request.Items {
			if f.request.Items[i].ID == sel.ID {
				f.request.Items[i] = sel
				f.request.Items[i].Selected = true
			}
		}
	}
	return nil
}

func (f *fakeDonationRepo) CompleteReceipt(_, _, _ uuid.UUID, activity *entity.ImpactActivity, deductions []donationRepo.PantryDeduction) error {
	if f.request.Status != "handed_over" {
		return donationRepo.ErrStateChanged
	}
	f.request.Status = "received"
	f.activities = append(f.activities, *activity)
	f.deductions = append(f.deductions, deductions...)
	return nil
}

type fakePantryRepo struct {
	pantryRepo.PantryRepoItf
	items map[uuid.UUID]entity.PantryIngredient
}

func (f *fakePantryRepo) FindAllByUserID(userID uuid.UUID, _ string) ([]entity.PantryIngredient, error) {
	var res []entity.PantryIngredient
	for _, item := range f.items {
		if item.UserID == userID {
			res = append(res, item)
		}
	}
	return res, nil
}

type fakeUserRepo struct {
	userRepo.UserRepoItf
}

func (f *fakeUserRepo) Show(param dto.UserParam) (entity.User, error) {
	return entity.User{ID: param.ID, Name: "User", PhoneNumber: "081234567890"}, nil
}

func TestFulfillThenConfirmReceivedLogsDonorImpact(t *testing.T) {
	requesterID, donorID := uuid.New(), uuid.New()
	carrotID, eggID, riceID := uuid.New(), uuid.New(), uuid.New()
	pantryCarrot, pantryEgg := uuid.New(), uuid.New()

	donations := &fakeDonationRepo{request: entity.DonationRequest{
		ID:          uuid.New(),
		RequesterID: requesterID,
		Title:       "Wortel & Telur",
		Status:      "open",
		ReceiveEnd:  time.Now().Add(time.Hour),
		Items: []entity.DonationRequestItem{
			{ID: carrotID, Name: "Wortel", Quantity: 500, Unit: "gram"},
			{ID: eggID, Name: "Telur", Quantity: 4, Unit: "butir"},
			{ID: riceID, Name: "Nasi Putih", Quantity: 2, Unit: "porsi"},
		},
	}}
	pantry := &fakePantryRepo{items: map[uuid.UUID]entity.PantryIngredient{
		pantryCarrot: {ID: pantryCarrot, UserID: donorID, Name: "Wortel", Quantity: 600, Status: "fresh"},
		pantryEgg:    {ID: pantryEgg, UserID: donorID, Name: "Telur Ayam", Quantity: 3, Status: "fresh"},
	}}
	u := NewDonationUsecase(donations, pantry, &fakeUserRepo{})
	id := donations.request.ID
	fulfill := dto.FulfillDonationRequest{ItemIDs: []string{carrotID.String(), eggID.String()}}

	if _, err := u.Fulfill(requesterID, id, fulfill); !errors.Is(err, ErrForbidden) {
		t.Fatalf("requester fulfilling own request: err = %v, want ErrForbidden", err)
	}
	if _, err := u.ConfirmReceived(requesterID, id); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("confirm before fulfil: err = %v, want ErrInvalidState", err)
	}
	if _, err := u.Fulfill(donorID, id, fulfill); err != nil {
		t.Fatal(err)
	}
	if _, err := u.Fulfill(uuid.New(), id, fulfill); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("second donor: err = %v, want ErrInvalidState", err)
	}
	if _, err := u.ConfirmReceived(donorID, id); !errors.Is(err, ErrForbidden) {
		t.Fatalf("donor confirming receipt: err = %v, want ErrForbidden", err)
	}
	if _, err := u.ConfirmReceived(requesterID, id); err != nil {
		t.Fatal(err)
	}

	if len(donations.activities) != 1 {
		t.Fatalf("impact activities = %d, want 1", len(donations.activities))
	}
	got := donations.activities[0]
	if got.UserID != donorID || got.ActivityType != "donation" {
		t.Fatalf("impact logged for %v (%s), want donor donation", got.UserID, got.ActivityType)
	}
	if got.PortionsSaved != 2 {
		t.Fatalf("portions = %d, want 2", got.PortionsSaved)
	}
	want := map[uuid.UUID]float64{pantryCarrot: 500, pantryEgg: 4}
	if len(donations.deductions) != len(want) {
		t.Fatalf("deductions = %v, want %v", donations.deductions, want)
	}
	for _, d := range donations.deductions {
		if want[d.PantryIngredientID] != d.Quantity {
			t.Fatalf("deduction %v = %v, want %v", d.PantryIngredientID, d.Quantity, want[d.PantryIngredientID])
		}
	}
	if _, err := u.ConfirmReceived(requesterID, id); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("double confirm: err = %v, want ErrInvalidState", err)
	}
}

func TestWhatsAppURL(t *testing.T) {
	for in, want := range map[string]string{
		"+62 812-3456-7890": "https://wa.me/6281234567890",
		"081234567890":      "https://wa.me/6281234567890",
	} {
		if got := whatsAppURL(in); got != want {
			t.Fatalf("whatsAppURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMatchPantryItemAndUnitConversion(t *testing.T) {
	sauceID := uuid.New()
	carrotKgID := uuid.New()
	carrotGramID := uuid.New()
	pantry := []entity.PantryIngredient{
		{ID: sauceID, Name: "Saus Tomat", Quantity: 1, Unit: "botol"},
		{ID: carrotKgID, Name: "Wortel Segar", Quantity: 1, Unit: "kg"},
		{ID: carrotGramID, Name: "Wortel", Quantity: 200, Unit: "gram"},
	}

	if got := matchPantryItem("Tomat", "buah", pantry); got != nil {
		t.Fatalf("Tomat unexpectedly matched %q", got.Name)
	}
	if got := matchPantryItem("Wortel", "gr", pantry); got == nil || got.ID != carrotGramID {
		t.Fatalf("Wortel matched %v, want exact Wortel (%v)", got, carrotGramID)
	}
	if qty, ok := unit.ConvertQuantity(500, "gram", "kg"); !ok || qty != 0.5 {
		t.Fatalf("500 gram to kg = (%v, %v), want (0.5, true)", qty, ok)
	}
	if _, ok := unit.ConvertQuantity(1, "ikat", "gram"); ok {
		t.Fatal("ikat to gram should not be compatible")
	}
}
