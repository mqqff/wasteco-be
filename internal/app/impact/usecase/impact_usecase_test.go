package usecase

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/app/impact/repository"
	"github.com/mqqff/wasteco-be/internal/domain/entity"
	"gorm.io/gorm"
)

type fakeImpactRepo struct {
	repository.ImpactRepoItf
	items map[uuid.UUID]entity.ImpactActivity
}

func (f *fakeImpactRepo) GetMonthlySummary(uuid.UUID, int, int) (repository.MonthlySummary, error) {
	return repository.MonthlySummary{}, nil
}

func (f *fakeImpactRepo) FindAllByUserID(uuid.UUID, int, int) ([]entity.ImpactActivity, int64, error) {
	return nil, 0, nil
}

func (f *fakeImpactRepo) FindByID(id uuid.UUID) (entity.ImpactActivity, error) {
	item, ok := f.items[id]
	if !ok {
		return entity.ImpactActivity{}, gorm.ErrRecordNotFound
	}
	return item, nil
}

func TestImpactEmptyListsAreNonNilAndDetailErrors(t *testing.T) {
	userID := uuid.New()
	otherID := uuid.New()
	actID := uuid.New()
	repo := &fakeImpactRepo{items: map[uuid.UUID]entity.ImpactActivity{
		actID: {ID: actID, UserID: otherID, ActivityType: "cooking", Title: "Nasi Goreng", PortionsSaved: 2},
	}}
	u := NewImpactUsecase(repo)

	sum, err := u.GetMonthlySummary(userID, 11, 2026)
	if err != nil {
		t.Fatal(err)
	}
	if sum.RecentActivities == nil || len(sum.RecentActivities) != 0 {
		t.Fatalf("RecentActivities = %#v, want non-nil empty slice", sum.RecentActivities)
	}

	hist, err := u.GetHistory(userID, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if hist.Activities == nil || len(hist.Activities) != 0 {
		t.Fatalf("Activities = %#v, want non-nil empty slice", hist.Activities)
	}

	if _, err := u.GetDetail(userID, uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing detail err = %v, want ErrNotFound", err)
	}
	if _, err := u.GetDetail(userID, actID); !errors.Is(err, ErrForbidden) {
		t.Fatalf("other user detail err = %v, want ErrForbidden", err)
	}
}
