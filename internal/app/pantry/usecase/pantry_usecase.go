package usecase

import (
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/app/pantry/repository"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
	"github.com/mqqff/wasteco-be/internal/domain/entity"
	"github.com/mqqff/wasteco-be/internal/infra/gemini"
	"gorm.io/gorm"
)

var (
	ErrNotFound          = errors.New("pantry ingredient not found")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrInvalidExpiryDate = errors.New("expiry_date must be YYYY-MM-DD")
)

type PantryUsecaseItf interface {
	GetIngredients(userID uuid.UUID, category, status string) ([]dto.PantryItemResponse, error)
	AddIngredient(userID uuid.UUID, req dto.CreatePantryItemRequest) (dto.PantryItemResponse, error)
	AddIngredients(userID uuid.UUID, req dto.BulkCreatePantryItemRequest) ([]dto.PantryItemResponse, error)
	ScanIngredients(image []byte, mimeType string) ([]gemini.DetectedIngredient, error)
	UpdateIngredient(userID uuid.UUID, id uuid.UUID, req dto.UpdatePantryItemRequest) (dto.PantryItemResponse, error)
	DeleteIngredient(userID uuid.UUID, id uuid.UUID) error
}

type PantryUsecase struct {
	pantryRepo repository.PantryRepoItf
	gemini     gemini.GeminiItf
}

func NewPantryUsecase(r repository.PantryRepoItf, g gemini.GeminiItf) PantryUsecaseItf {
	return &PantryUsecase{
		pantryRepo: r,
		gemini:     g,
	}
}

func (u *PantryUsecase) GetIngredients(userID uuid.UUID, category, status string) ([]dto.PantryItemResponse, error) {
	items, err := u.pantryRepo.FindAllByUserID(userID, category)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.PantryItemResponse, 0, len(items))
	now := time.Now()

	for _, item := range items {
		resp := u.mapToResponse(item, now)
		if status != "" && resp.Status != status {
			continue
		}
		responses = append(responses, resp)
	}

	return responses, nil
}

func (u *PantryUsecase) AddIngredient(userID uuid.UUID, req dto.CreatePantryItemRequest) (dto.PantryItemResponse, error) {
	item, err := newPantryItem(userID, req)
	if err != nil {
		return dto.PantryItemResponse{}, err
	}

	if err := u.pantryRepo.Create(&item); err != nil {
		return dto.PantryItemResponse{}, err
	}

	return u.mapToResponse(item, time.Now()), nil
}

func (u *PantryUsecase) AddIngredients(userID uuid.UUID, req dto.BulkCreatePantryItemRequest) ([]dto.PantryItemResponse, error) {
	items := make([]entity.PantryIngredient, 0, len(req.Items))
	for _, r := range req.Items {
		item, err := newPantryItem(userID, r)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := u.pantryRepo.CreateMany(items); err != nil {
		return nil, err
	}

	now := time.Now()
	responses := make([]dto.PantryItemResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, u.mapToResponse(item, now))
	}

	return responses, nil
}

func (u *PantryUsecase) ScanIngredients(image []byte, mimeType string) ([]gemini.DetectedIngredient, error) {
	items, err := u.gemini.DetectIngredients(image, mimeType)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []gemini.DetectedIngredient{}
	}

	return items, nil
}

func parseExpiryDate(raw string) (*time.Time, error) {
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, ErrInvalidExpiryDate
	}
	return &parsed, nil
}

func resolveActiveStatus(status string, expiryDate *time.Time, now time.Time) string {
	if status == "used" || status == "donated" {
		return status
	}
	if expiryDate != nil && daysUntil(*expiryDate, now) <= 3 {
		return "near_expiry"
	}
	return "fresh"
}

func daysUntil(expiry, now time.Time) int {
	return int(math.Ceil(expiry.Sub(now).Hours() / 24))
}

func newPantryItem(userID uuid.UUID, req dto.CreatePantryItemRequest) (entity.PantryIngredient, error) {
	var expiryDate *time.Time
	now := time.Now()

	if req.ExpiryDate != nil && *req.ExpiryDate != "" {
		parsed, err := parseExpiryDate(*req.ExpiryDate)
		if err != nil {
			return entity.PantryIngredient{}, err
		}
		expiryDate = parsed
	}

	return entity.PantryIngredient{
		ID:              uuid.New(),
		UserID:          userID,
		Name:            req.Name,
		Quantity:        req.Quantity,
		Unit:            req.Unit,
		Category:        req.Category,
		StorageLocation: req.StorageLocation,
		ExpiryDate:      expiryDate,
		Status:          resolveActiveStatus("fresh", expiryDate, now),
	}, nil
}

func (u *PantryUsecase) UpdateIngredient(userID uuid.UUID, id uuid.UUID, req dto.UpdatePantryItemRequest) (dto.PantryItemResponse, error) {
	item, err := u.pantryRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.PantryItemResponse{}, ErrNotFound
		}
		return dto.PantryItemResponse{}, err
	}

	if item.UserID != userID {
		return dto.PantryItemResponse{}, ErrUnauthorized
	}

	if req.Name != "" {
		item.Name = req.Name
	}
	if req.Quantity != nil {
		item.Quantity = *req.Quantity
		if item.Quantity == 0 {
			item.Status = "used"
		}
	}
	if req.Unit != "" {
		item.Unit = req.Unit
	}
	if req.Category != "" {
		item.Category = req.Category
	}
	if req.StorageLocation != "" {
		item.StorageLocation = req.StorageLocation
	}
	if req.ExpiryDate != nil {
		if *req.ExpiryDate != "" {
			parsed, err := parseExpiryDate(*req.ExpiryDate)
			if err != nil {
				return dto.PantryItemResponse{}, err
			}
			item.ExpiryDate = parsed
		} else {
			item.ExpiryDate = nil
		}
	}
	if req.Status != "" {
		item.Status = req.Status
	} else if req.ExpiryDate != nil {
		item.Status = resolveActiveStatus(item.Status, item.ExpiryDate, time.Now())
	}

	err = u.pantryRepo.Update(&item)
	if err != nil {
		return dto.PantryItemResponse{}, err
	}

	return u.mapToResponse(item, time.Now()), nil
}

func (u *PantryUsecase) DeleteIngredient(userID uuid.UUID, id uuid.UUID) error {
	err := u.pantryRepo.Delete(id, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func (u *PantryUsecase) mapToResponse(item entity.PantryIngredient, now time.Time) dto.PantryItemResponse {
	resp := dto.PantryItemResponse{
		ID:              item.ID.String(),
		Name:            item.Name,
		Quantity:        item.Quantity,
		Unit:            item.Unit,
		Category:        item.Category,
		StorageLocation: item.StorageLocation,
		Status:          resolveActiveStatus(item.Status, item.ExpiryDate, now),
		CreatedAt:       item.CreatedAt.Format(time.RFC3339),
	}

	if item.ExpiryDate != nil {
		resp.ExpiryDate = item.ExpiryDate.Format("2006-01-02")
		days := daysUntil(*item.ExpiryDate, now)
		resp.DaysUntilExpiry = &days
	}

	return resp
}
