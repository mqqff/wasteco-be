package repository

import (
	"github.com/mqqff/wasteco-be/internal/domain/entity"
	"github.com/mqqff/wasteco-be/pkg/log"
	"gorm.io/gorm"
)

type UserRepoItf interface {
	Create(user *entity.User) error
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
