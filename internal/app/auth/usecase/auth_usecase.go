package usecase

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/app/user/repository"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
	"github.com/mqqff/wasteco-be/internal/domain/entity"
	"github.com/mqqff/wasteco-be/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailExists        = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
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
	email := strings.ToLower(strings.TrimSpace(req.Email))

	existingUser, err := u.userRepo.Show(dto.UserParam{Email: email})
	if err != nil {
		return dto.AuthResponse{}, err
	}

	if existingUser.ID != uuid.Nil {
		return dto.AuthResponse{}, ErrEmailExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	newUser := entity.User{
		ID:              uuid.New(),
		Email:           email,
		Name:            strings.TrimSpace(req.Name),
		PasswordHash:    string(hashedPassword),
		PhoneNumber:     strings.TrimSpace(req.PhoneNumber),
		AreaName:        strings.TrimSpace(req.AreaName),
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
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
			Latitude:        newUser.Latitude,
			Longitude:       newUser.Longitude,
			SharingRadiusKM: newUser.SharingRadiusKM,
		},
	}, nil
}

func (u *AuthUsecase) Login(req dto.LoginRequest) (dto.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	user, err := u.userRepo.Show(dto.UserParam{Email: email})
	if err != nil {
		return dto.AuthResponse{}, err
	}

	if user.ID == uuid.Nil || user.PasswordHash == "" {
		return dto.AuthResponse{}, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return dto.AuthResponse{}, ErrInvalidCredentials
	}

	jwtToken, err := u.jwt.GenerateToken(user.ID)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	totalPortions, _ := u.userRepo.GetContributionSummary(user.ID)

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
		},
	}, nil
}
