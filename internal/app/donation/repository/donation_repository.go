package repository

import (
	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/domain/entity"
	"github.com/mqqff/wasteco-be/pkg/log"
	"gorm.io/gorm"
)

type DonationRepoItf interface {
	Create(donation *entity.Donation) error
	FindAllAvailable(category, search string) ([]entity.Donation, error)
	FindByID(id uuid.UUID) (entity.Donation, error)
	FindByUserID(userID uuid.UUID) ([]entity.Donation, error)
	Update(donation *entity.Donation) error
	Delete(id uuid.UUID, userID uuid.UUID) error
}

type DonationRepo struct {
	db *gorm.DB
}

func NewDonationRepo(db *gorm.DB) DonationRepoItf {
	return &DonationRepo{db}
}

func (r *DonationRepo) Create(donation *entity.Donation) error {
	err := r.db.Create(donation).Error
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[DonationRepo][Create] failed to create donation")
	}

	return err
}

func (r *DonationRepo) FindAllAvailable(category, search string) ([]entity.Donation, error) {
	var donations []entity.Donation
	query := r.db.Preload("User").Where("status = ?", "available")

	if category != "" && category != "Semua" && category != "Semua Kategori" {
		query = query.Where("category = ?", category)
	}

	if search != "" {
		query = query.Where("title LIKE ? OR condition_note LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	err := query.Order("created_at DESC").Find(&donations).Error
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[DonationRepo][FindAllAvailable] failed to fetch available donations")
		return nil, err
	}

	return donations, nil
}

func (r *DonationRepo) FindByID(id uuid.UUID) (entity.Donation, error) {
	var donation entity.Donation
	err := r.db.Preload("User").Where("id = ?", id).First(&donation).Error
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[DonationRepo][FindByID] failed to find donation")
		return entity.Donation{}, err
	}

	return donation, nil
}

func (r *DonationRepo) FindByUserID(userID uuid.UUID) ([]entity.Donation, error) {
	var donations []entity.Donation
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&donations).Error
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[DonationRepo][FindByUserID] failed to find user donations")
		return nil, err
	}

	return donations, nil
}

func (r *DonationRepo) Update(donation *entity.Donation) error {
	err := r.db.Save(donation).Error
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[DonationRepo][Update] failed to update donation")
	}

	return err
}

func (r *DonationRepo) Delete(id uuid.UUID, userID uuid.UUID) error {
	err := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&entity.Donation{}).Error
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[DonationRepo][Delete] failed to delete donation")
	}

	return err
}
