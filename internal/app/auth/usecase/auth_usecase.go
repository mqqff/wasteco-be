package usecase

import (
	"errors"

	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/app/user/repository"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
	"github.com/mqqff/wasteco-be/internal/domain/entity"
	"github.com/mqqff/wasteco-be/internal/domain/env"
	"github.com/mqqff/wasteco-be/pkg/jwt"
	"github.com/mqqff/wasteco-be/pkg/oauth"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecaseItf interface {
	HandleRedirect(state string) (string, error)
	HandleCallback(data dto.GoogleCallbackRequest) (string, error)
	Register(req dto.RegisterRequest) (dto.AuthResponse, error)
	Login(req dto.LoginRequest) (dto.AuthResponse, error)
}

type AuthUsecase struct {
	userRepo repository.UserRepoItf
	oauth    oauth.OAuthItf
	jwt      jwt.JWTIf
}

func NewAuthUsecase(r repository.UserRepoItf, o oauth.OAuthItf, j jwt.JWTIf) AuthUsecaseItf {
	return &AuthUsecase{
		userRepo: r,
		oauth:    o,
		jwt:      j,
	}
}

func (u *AuthUsecase) HandleRedirect(state string) (string, error) {
	return u.oauth.GenerateAuthLink(state), nil
}

func (u *AuthUsecase) HandleCallback(data dto.GoogleCallbackRequest) (string, error) {
	if data.State != env.AppEnv.OAuthState {
		return "", errors.New("invalid state")
	}

	token, err := u.oauth.ExchangeToken(data.Code)
	if err != nil {
		return "", err
	}

	userInfo, err := u.oauth.GetUserInfo(token)
	if err != nil {
		return "", err
	}

	user, err := u.userRepo.Show(dto.UserParam{Email: userInfo.Email})
	if err != nil {
		return "", err
	}

	if user.ID == uuid.Nil {
		user = entity.User{
			ID:              uuid.New(),
			Email:           userInfo.Email,
			Name:            userInfo.Name,
			SharingRadiusKM: 2,
		}

		err := u.userRepo.Create(&user)
		if err != nil {
			return "", err
		}
	}

	jwtToken, err := u.jwt.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}

	return jwtToken, nil
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
		return dto.AuthResponse{}, errors.New("this account was registered via Google, please login with Google")
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
