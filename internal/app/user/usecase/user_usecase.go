package usecase

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/app/user/repository"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
)

var ErrUserNotFound = errors.New("user not found")

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
		return dto.UserProfile{}, ErrUserNotFound
	}

	totalPortions, err := u.userRepo.GetContributionSummary(userID)
	if err != nil {
		totalPortions = 0
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
	}

	return userProfile, nil
}

func (u *UserUsecase) UpdateProfile(userID uuid.UUID, req dto.UpdateProfileRequest) (dto.UserProfile, error) {
	user, err := u.userRepo.Show(dto.UserParam{ID: userID})
	if err != nil {
		return dto.UserProfile{}, err
	}
	if user.ID == uuid.Nil {
		return dto.UserProfile{}, ErrUserNotFound
	}

	if name := strings.TrimSpace(req.Name); name != "" {
		user.Name = name
	}
	if req.PhoneNumber != nil {
		user.PhoneNumber = strings.TrimSpace(*req.PhoneNumber)
	}
	if area := strings.TrimSpace(req.AreaName); area != "" {
		user.AreaName = area
	}
	if req.Latitude != nil {
		user.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		user.Longitude = *req.Longitude
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
