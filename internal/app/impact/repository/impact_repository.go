package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/domain/entity"
	"github.com/mqqff/wasteco-be/pkg/log"
	"gorm.io/gorm"
)

type MonthlySummary struct {
	TotalPortions    int     `gorm:"column:total_portions"`
	TotalCO2         float64 `gorm:"column:total_co2"`
	CookingPortions  int     `gorm:"column:cooking_portions"`
	DonationPortions int     `gorm:"column:donation_portions"`
}

type ImpactRepoItf interface {
	Create(activity *entity.ImpactActivity) error
	FindAllByUserID(userID uuid.UUID, page, limit int) ([]entity.ImpactActivity, int64, error)
	FindByID(id uuid.UUID) (entity.ImpactActivity, error)
	GetMonthlySummary(userID uuid.UUID, month, year int) (MonthlySummary, error)
}

type ImpactRepo struct {
	db *gorm.DB
}

func NewImpactRepo(db *gorm.DB) ImpactRepoItf {
	return &ImpactRepo{db}
}

func (r *ImpactRepo) Create(activity *entity.ImpactActivity) error {
	err := r.db.Create(activity).Error
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[ImpactRepo][Create] failed to create impact activity")
	}

	return err
}

func (r *ImpactRepo) FindAllByUserID(userID uuid.UUID, page, limit int) ([]entity.ImpactActivity, int64, error) {
	if limit < 1 {
		limit = 10
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	var total int64
	var items []entity.ImpactActivity

	err := r.db.Model(&entity.ImpactActivity{}).
		Where("user_id = ?", userID).
		Count(&total).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&items).Error

	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[ImpactRepo][FindAllByUserID] failed to retrieve impact activities")
		return nil, 0, err
	}

	return items, total, nil
}

func (r *ImpactRepo) FindByID(id uuid.UUID) (entity.ImpactActivity, error) {
	var item entity.ImpactActivity
	err := r.db.Where("id = ?", id).First(&item).Error
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[ImpactRepo][FindByID] failed to find impact activity")
		return entity.ImpactActivity{}, err
	}

	return item, nil
}

func (r *ImpactRepo) GetMonthlySummary(userID uuid.UUID, month, year int) (MonthlySummary, error) {
	if year == 0 {
		year = time.Now().Year()
	}
	if month == 0 {
		month = int(time.Now().Month())
	}

	startDate := fmt.Sprintf("%04d-%02d-01 00:00:00", year, month)
	nextYear := year
	nextMonth := month + 1
	if nextMonth > 12 {
		nextMonth = 1
		nextYear++
	}
	endDate := fmt.Sprintf("%04d-%02d-01 00:00:00", nextYear, nextMonth)

	var summary MonthlySummary
	query := `
		SELECT 
			COALESCE(SUM(portions_saved), 0) AS total_portions,
			COALESCE(SUM(co2_prevented_kg), 0) AS total_co2,
			COALESCE(SUM(CASE WHEN activity_type = 'cooking' THEN portions_saved ELSE 0 END), 0) AS cooking_portions,
			COALESCE(SUM(CASE WHEN activity_type = 'donation' THEN portions_saved ELSE 0 END), 0) AS donation_portions
		FROM impact_activities
		WHERE user_id = ? AND created_at >= ? AND created_at < ?
	`

	err := r.db.Raw(query, userID, startDate, endDate).Scan(&summary).Error
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[ImpactRepo][GetMonthlySummary] failed to get monthly summary")
		return MonthlySummary{}, err
	}

	return summary, nil
}
