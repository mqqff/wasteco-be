package repository

import (
	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
	"github.com/mqqff/wasteco-be/internal/domain/entity"
	"github.com/mqqff/wasteco-be/pkg/log"
	"gorm.io/gorm"
)

type UserRepoItf interface {
	Show(param dto.UserParam) (entity.User, error)
	Create(user *entity.User) error
	Update(user *entity.User) error
	GetContributionSummary(userID uuid.UUID) (totalPortions int, totalMoney float64, err error)
}

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) UserRepoItf {
	return &UserRepo{db}
}

func (r *UserRepo) Create(user *entity.User) error {
	err := r.db.Create(user).Error
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[UserRepo][Create] failed to create user")
	}

	return err
}

func (r *UserRepo) Update(user *entity.User) error {
	err := r.db.Save(user).Error
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[UserRepo][Update] failed to update user")
	}

	return err
}

func (r *UserRepo) Show(param dto.UserParam) (entity.User, error) {
	user := entity.User{}
	err := r.db.Find(&user, param).Error
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[UserRepo][Show] failed to get user")
		return entity.User{}, err
	}

	return user, err
}

func (r *UserRepo) GetContributionSummary(userID uuid.UUID) (int, float64, error) {
	type Result struct {
		TotalPortions int     `gorm:"column:total_portions"`
		TotalMoney    float64 `gorm:"column:total_money"`
	}
	var res Result
	err := r.db.Model(&entity.ImpactActivity{}).
		Select("COALESCE(SUM(portions_saved), 0) as total_portions, COALESCE(SUM(money_saved), 0) as total_money").
		Where("user_id = ?", userID).
		Scan(&res).Error

	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[UserRepo][GetContributionSummary] failed to get summary")
		return 0, 0, err
	}

	return res.TotalPortions, res.TotalMoney, nil
}
