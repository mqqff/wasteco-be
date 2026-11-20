package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/domain/entity"
	"github.com/mqqff/wasteco-be/pkg/log"
	"gorm.io/gorm"
)

type PantryRepoItf interface {
	Create(item *entity.PantryIngredient) error
	CreateMany(items []entity.PantryIngredient) error
	FindAllByUserID(userID uuid.UUID, category string) ([]entity.PantryIngredient, error)
	FindByID(id uuid.UUID) (entity.PantryIngredient, error)
	Update(item *entity.PantryIngredient) error
	Delete(id uuid.UUID, userID uuid.UUID) error
}

type PantryRepo struct {
	db *gorm.DB
}

func NewPantryRepo(db *gorm.DB) PantryRepoItf {
	return &PantryRepo{db}
}

func (r *PantryRepo) Create(item *entity.PantryIngredient) error {
	err := r.db.Create(item).Error
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[PantryRepo][Create] failed to create pantry ingredient")
	}

	return err
}

func (r *PantryRepo) CreateMany(items []entity.PantryIngredient) error {
	err := r.db.Create(&items).Error
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[PantryRepo][CreateMany] failed to create pantry ingredients")
	}

	return err
}

func (r *PantryRepo) FindAllByUserID(userID uuid.UUID, category string) ([]entity.PantryIngredient, error) {
	var items []entity.PantryIngredient
	query := r.db.Where("user_id = ? AND status IN ('fresh', 'near_expiry')", userID)

	if category != "" && category != "All" && category != "all" && category != "Semua" {
		query = query.Where("category = ?", category)
	}

	err := query.Order("CASE WHEN expiry_date IS NULL THEN 1 ELSE 0 END, expiry_date ASC, created_at DESC").Find(&items).Error
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[PantryRepo][FindAllByUserID] failed to retrieve pantry ingredients")
		return nil, err
	}

	return items, nil
}

func (r *PantryRepo) FindByID(id uuid.UUID) (entity.PantryIngredient, error) {
	var item entity.PantryIngredient
	err := r.db.Where("id = ?", id).First(&item).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(log.LogInfo{
				"error": err.Error(),
			}, "[PantryRepo][FindByID] failed to find pantry ingredient")
		}
		return entity.PantryIngredient{}, err
	}

	return item, nil
}

func (r *PantryRepo) Update(item *entity.PantryIngredient) error {
	err := r.db.Save(item).Error
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[PantryRepo][Update] failed to update pantry ingredient")
	}

	return err
}

func (r *PantryRepo) Delete(id uuid.UUID, userID uuid.UUID) error {
	res := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&entity.PantryIngredient{})
	if res.Error != nil {
		log.Error(log.LogInfo{
			"error": res.Error.Error(),
		}, "[PantryRepo][Delete] failed to delete pantry ingredient")
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
