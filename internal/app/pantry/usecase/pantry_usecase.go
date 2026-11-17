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
)

type PantryUsecaseItf interface {
	GetIngredients(userID uuid.UUID, category string) ([]dto.PantryItemResponse, error)
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

func (u *PantryUsecase) GetIngredients(userID uuid.UUID, category string) ([]dto.PantryItemResponse, error) {
	items, err := u.pantryRepo.FindAllByUserID(userID, category)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.PantryItemResponse, 0, len(items))
	now := time.Now()

	for _, item := range items {
		resp := u.mapToResponse(item, now)
		responses = append(responses, resp)
	}

	return responses, nil
}

func (u *PantryUsecase) AddIngredient(userID uuid.UUID, req dto.CreatePantryItemRequest) (dto.PantryItemResponse, error) {
	item := newPantryItem(userID, req)

	err := u.pantryRepo.Create(&item)
	if err != nil {
		return dto.PantryItemResponse{}, err
	}

	return u.mapToResponse(item, time.Now()), nil
}

func (u *PantryUsecase) AddIngredients(userID uuid.UUID, req dto.BulkCreatePantryItemRequest) ([]dto.PantryItemResponse, error) {
	items := make([]entity.PantryIngredient, 0, len(req.Items))
	for _, r := range req.Items {
		items = append(items, newPantryItem(userID, r))
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

func newPantryItem(userID uuid.UUID, req dto.CreatePantryItemRequest) entity.PantryIngredient {
	var expiryDate *time.Time
	status := "fresh"

	if req.ExpiryDate != nil && *req.ExpiryDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.ExpiryDate)
		if err == nil {
			expiryDate = &parsed
			days := int(math.Ceil(time.Until(parsed).Hours() / 24))
			if days <= 3 {
				status = "near_expiry"
			}
		}
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
		Status:          status,
		EstimatedPrice:  req.EstimatedPrice,
	}
}

func (u *PantryUsecase) UpdateIngredient(userID uuid.UUID, id uuid.UUID, req dto.UpdatePantryItemRequest) (dto.PantryItemResponse, error) {
	item, err := u.pantryRepo.FindByID(id)
	if err != nil {
		return dto.PantryItemResponse{}, err
	}

	if item.UserID != userID {
		return dto.PantryItemResponse{}, errors.New("unauthorized")
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
	if req.Status != "" {
		item.Status = req.Status
	}
	if req.EstimatedPrice != nil {
		item.EstimatedPrice = *req.EstimatedPrice
	}
	if req.ExpiryDate != nil {
		if *req.ExpiryDate != "" {
			parsed, err := time.Parse("2006-01-02", *req.ExpiryDate)
			if err == nil {
				item.ExpiryDate = &parsed
			}
		} else {
			item.ExpiryDate = nil
		}
	}

	err = u.pantryRepo.Update(&item)
	if err != nil {
		return dto.PantryItemResponse{}, err
	}

	return u.mapToResponse(item, time.Now()), nil
}

func (u *PantryUsecase) DeleteIngredient(userID uuid.UUID, id uuid.UUID) error {
	return u.pantryRepo.Delete(id, userID)
}

func (u *PantryUsecase) mapToResponse(item entity.PantryIngredient, now time.Time) dto.PantryItemResponse {
	resp := dto.PantryItemResponse{
		ID:              item.ID.String(),
		Name:            item.Name,
		Quantity:        item.Quantity,
		Unit:            item.Unit,
		Category:        item.Category,
		StorageLocation: item.StorageLocation,
		Status:          item.Status,
		EstimatedPrice:  item.EstimatedPrice,
		CreatedAt:       item.CreatedAt.Format(time.RFC3339),
	}

	if item.ExpiryDate != nil {
		resp.ExpiryDate = item.ExpiryDate.Format("2006-01-02")
		days := int(math.Ceil(item.ExpiryDate.Sub(now).Hours() / 24))
		resp.DaysUntilExpiry = &days
	}

	return resp
}
