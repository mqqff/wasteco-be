package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/domain/entity"
	"github.com/mqqff/wasteco-be/pkg/log"
	"gorm.io/gorm"
)

var ErrStateChanged = errors.New("donation request state changed")

type PantryDeduction struct {
	PantryIngredientID uuid.UUID
	Quantity           float64
}

type DonationRepoItf interface {
	Create(request *entity.DonationRequest) error
	FindOpen(excludeUserID uuid.UUID, search string) ([]entity.DonationRequest, error)
	FindByID(id uuid.UUID) (entity.DonationRequest, error)
	FindByRequesterID(userID uuid.UUID) ([]entity.DonationRequest, error)
	Fulfill(id, donorID uuid.UUID, items []entity.DonationRequestItem) error
	CompleteReceipt(id, requesterID, donorID uuid.UUID, activity *entity.ImpactActivity, deductions []PantryDeduction) error
	Cancel(id, requesterID uuid.UUID) error
}

type DonationRepo struct {
	db *gorm.DB
}

func NewDonationRepo(db *gorm.DB) DonationRepoItf {
	return &DonationRepo{db}
}

func (r *DonationRepo) Create(request *entity.DonationRequest) error {
	err := r.db.Create(request).Error
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[DonationRepo][Create] failed to create donation request")
	}

	return err
}

func (r *DonationRepo) FindOpen(excludeUserID uuid.UUID, search string) ([]entity.DonationRequest, error) {
	var requests []entity.DonationRequest
	query := r.db.Preload("Requester").Preload("Items").
		Where("status = ? AND receive_end > ? AND requester_id <> ?", "open", time.Now(), excludeUserID)

	if search != "" {
		like := "%" + search + "%"
		query = query.Where(
			"title LIKE ? OR EXISTS (SELECT 1 FROM donation_request_items i WHERE i.request_id = donation_requests.id AND i.name LIKE ?)",
			like, like,
		)
	}

	err := query.Order("created_at DESC").Find(&requests).Error
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[DonationRepo][FindOpen] failed to fetch open donation requests")
		return nil, err
	}

	return requests, nil
}

func (r *DonationRepo) FindByID(id uuid.UUID) (entity.DonationRequest, error) {
	var request entity.DonationRequest
	err := r.db.Preload("Requester").Preload("Donor").Preload("Items").Where("id = ?", id).First(&request).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(log.LogInfo{
				"error": err.Error(),
			}, "[DonationRepo][FindByID] failed to find donation request")
		}
		return entity.DonationRequest{}, err
	}

	return request, nil
}

func (r *DonationRepo) FindByRequesterID(userID uuid.UUID) ([]entity.DonationRequest, error) {
	var requests []entity.DonationRequest
	err := r.db.Preload("Donor").Preload("Items").Where("requester_id = ?", userID).Order("created_at DESC").Find(&requests).Error
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[DonationRepo][FindByRequesterID] failed to find user donation requests")
		return nil, err
	}

	return requests, nil
}

func (r *DonationRepo) Fulfill(id, donorID uuid.UUID, items []entity.DonationRequestItem) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&entity.DonationRequest{}).
			Where("id = ? AND status = ?", id, "open").
			Updates(map[string]any{"donor_id": donorID, "status": "handed_over", "handed_over_at": time.Now()})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrStateChanged
		}

		for _, item := range items {
			err := tx.Model(&entity.DonationRequestItem{}).
				Where("id = ? AND request_id = ?", item.ID, id).
				Updates(map[string]any{"selected": true, "pantry_ingredient_id": item.PantryIngredientID, "unit_price": item.UnitPrice}).Error
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil && !errors.Is(err, ErrStateChanged) {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[DonationRepo][Fulfill] failed to fulfill donation request")
	}

	return err
}

func (r *DonationRepo) CompleteReceipt(id, requesterID, donorID uuid.UUID, activity *entity.ImpactActivity, deductions []PantryDeduction) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&entity.DonationRequest{}).
			Where("id = ? AND requester_id = ? AND status = ?", id, requesterID, "handed_over").
			Updates(map[string]any{"status": "received", "received_at": time.Now()})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrStateChanged
		}

		for _, d := range deductions {
			err := tx.Exec(
				`UPDATE pantry_ingredients
				SET status = CASE WHEN quantity <= ? THEN 'donated' ELSE status END,
					quantity = GREATEST(quantity - ?, 0)
				WHERE id = ? AND user_id = ? AND status IN ('fresh', 'near_expiry')`,
				d.Quantity, d.Quantity, d.PantryIngredientID, donorID,
			).Error
			if err != nil {
				return err
			}
		}

		return tx.Create(activity).Error
	})
	if err != nil && !errors.Is(err, ErrStateChanged) {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[DonationRepo][CompleteReceipt] failed to complete donation receipt")
	}

	return err
}

func (r *DonationRepo) Cancel(id, requesterID uuid.UUID) error {
	return r.transition("Cancel", id, requesterID, []string{"open", "handed_over"},
		map[string]any{"status": "cancelled"})
}

func (r *DonationRepo) transition(method string, id, requesterID uuid.UUID, from []string, updates map[string]any) error {
	res := r.db.Model(&entity.DonationRequest{}).
		Where("id = ? AND requester_id = ? AND status IN ?", id, requesterID, from).
		Updates(updates)
	if res.Error != nil {
		log.Error(log.LogInfo{
			"error": res.Error.Error(),
		}, "[DonationRepo]["+method+"] failed to update donation request status")
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrStateChanged
	}

	return nil
}
