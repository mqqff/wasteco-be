package usecase

import (
	"errors"

	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/app/user/repository"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
)

type UserUsecaseItf interface {
	GetProfile(userID uuid.UUID) (dto.UserProfile, error)
	UpdateProfile(userID uuid.UUID, req dto.UpdateProfileRequest) (dto.UserProfile, error)
}

type UserUsecase struct {
	userRepo repository.UserRepoItf
}

func NewUserUsecase(r repository.UserRepoItf) UserUsecaseItf {
	return &UserUsecase{
		userRepo: r,
	}
}

func (u *UserUsecase) GetProfile(userID uuid.UUID) (dto.UserProfile, error) {
	user, err := u.userRepo.Show(dto.UserParam{ID: userID})
	if err != nil {
		return dto.UserProfile{}, err
	}
	if user.ID == uuid.Nil {
		return dto.UserProfile{}, errors.New("user not found")
	}

	totalPortions, totalMoney, err := u.userRepo.GetContributionSummary(userID)
	if err != nil {
		totalPortions = 0
		totalMoney = 0
	}

	userProfile := dto.UserProfile{
		ID:                 user.ID.String(),
		Email:              user.Email,
		Name:               user.Name,
		PhoneNumber:        user.PhoneNumber,
		AreaName:           user.AreaName,
		Latitude:           user.Latitude,
		Longitude:          user.Longitude,
		SharingRadiusKM:    user.SharingRadiusKM,
		TotalPortionsSaved: totalPortions,
		TotalMoneySaved:    totalMoney,
	}

	return userProfile, nil
}

func (u *UserUsecase) UpdateProfile(userID uuid.UUID, req dto.UpdateProfileRequest) (dto.UserProfile, error) {
	user, err := u.userRepo.Show(dto.UserParam{ID: userID})
	if err != nil {
		return dto.UserProfile{}, err
	}
	if user.ID == uuid.Nil {
		return dto.UserProfile{}, errors.New("user not found")
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.PhoneNumber != "" {
		user.PhoneNumber = req.PhoneNumber
	}
	if req.AreaName != "" {
		user.AreaName = req.AreaName
	}
	if req.Latitude != 0 {
		user.Latitude = req.Latitude
	}
	if req.Longitude != 0 {
		user.Longitude = req.Longitude
	}
	if req.SharingRadiusKM > 0 {
		user.SharingRadiusKM = req.SharingRadiusKM
	}

	err = u.userRepo.Update(&user)
	if err != nil {
		return dto.UserProfile{}, err
	}

	return u.GetProfile(userID)
}
