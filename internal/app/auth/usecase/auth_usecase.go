package usecase

import (
	"errors"

	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/app/user/repository"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
	"github.com/mqqff/wasteco-be/internal/domain/entity"
	"github.com/mqqff/wasteco-be/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecaseItf interface {
	Register(req dto.RegisterRequest) (dto.AuthResponse, error)
	Login(req dto.LoginRequest) (dto.AuthResponse, error)
}

type AuthUsecase struct {
	userRepo repository.UserRepoItf
	jwt      jwt.JWTIf
}

func NewAuthUsecase(r repository.UserRepoItf, j jwt.JWTIf) AuthUsecaseItf {
	return &AuthUsecase{
		userRepo: r,
		jwt:      j,
	}
}

func (u *AuthUsecase) Register(req dto.RegisterRequest) (dto.AuthResponse, error) {
	existingUser, err := u.userRepo.Show(dto.UserParam{Email: req.Email})
	if err != nil {
		return dto.AuthResponse{}, err
	}

	if existingUser.ID != uuid.Nil {
		return dto.AuthResponse{}, errors.New("email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	newUser := entity.User{
		ID:              uuid.New(),
		Email:           req.Email,
		Name:            req.Name,
		PasswordHash:    string(hashedPassword),
		PhoneNumber:     req.PhoneNumber,
		AreaName:        req.AreaName,
		SharingRadiusKM: 2,
	}

	err = u.userRepo.Create(&newUser)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	jwtToken, err := u.jwt.GenerateToken(newUser.ID)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	return dto.AuthResponse{
		Token: jwtToken,
		User: dto.UserProfile{
			ID:              newUser.ID.String(),
			Email:           newUser.Email,
			Name:            newUser.Name,
			PhoneNumber:     newUser.PhoneNumber,
			AreaName:        newUser.AreaName,
			SharingRadiusKM: newUser.SharingRadiusKM,
		},
	}, nil
}

func (u *AuthUsecase) Login(req dto.LoginRequest) (dto.AuthResponse, error) {
	user, err := u.userRepo.Show(dto.UserParam{Email: req.Email})
	if err != nil {
		return dto.AuthResponse{}, err
	}

	if user.ID == uuid.Nil {
		return dto.AuthResponse{}, errors.New("invalid email or password")
	}

	if user.PasswordHash == "" {
		return dto.AuthResponse{}, errors.New("account password not configured")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return dto.AuthResponse{}, errors.New("invalid email or password")
	}

	jwtToken, err := u.jwt.GenerateToken(user.ID)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	totalPortions, totalMoney, _ := u.userRepo.GetContributionSummary(user.ID)

	return dto.AuthResponse{
		Token: jwtToken,
		User: dto.UserProfile{
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
		},
	}, nil
}
