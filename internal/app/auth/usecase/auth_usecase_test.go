package usecase

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/app/user/repository"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
	"github.com/mqqff/wasteco-be/internal/domain/entity"
	"github.com/mqqff/wasteco-be/pkg/jwt"
)

type fakeUserRepo struct {
	repository.UserRepoItf
	users map[string]entity.User
}

func (f *fakeUserRepo) Show(param dto.UserParam) (entity.User, error) {
	if param.Email != "" {
		return f.users[param.Email], nil
	}
	for _, u := range f.users {
		if u.ID == param.ID {
			return u, nil
		}
	}
	return entity.User{}, nil
}

func (f *fakeUserRepo) Create(user *entity.User) error {
	f.users[user.Email] = *user
	return nil
}

func (f *fakeUserRepo) GetContributionSummary(uuid.UUID) (int, error) {
	return 0, nil
}

type fakeJWT struct{}

func (fakeJWT) GenerateToken(userID uuid.UUID) (string, error) {
	return "token-" + userID.String(), nil
}

func (fakeJWT) Decode(string, *jwt.Claims) error {
	return nil
}

func TestRegisterAndLoginNormalization(t *testing.T) {
	repo := &fakeUserRepo{users: map[string]entity.User{}}
	u := NewAuthUsecase(repo, fakeJWT{})

	res, err := u.Register(dto.RegisterRequest{
		Name:      "Affan",
		Email:     " Affan@WasteCo.id ",
		Password:  "password123",
		Latitude:  -7.9528,
		Longitude: 112.6074,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.User.Email != "affan@wasteco.id" || res.User.Latitude != -7.9528 || res.User.Longitude != 112.6074 {
		t.Fatalf("unexpected registered user: %+v", res.User)
	}

	if _, err := u.Register(dto.RegisterRequest{
		Name: "Affan", Email: "affan@wasteco.id", Password: "password123",
	}); !errors.Is(err, ErrEmailExists) {
		t.Fatalf("duplicate register err = %v, want ErrEmailExists", err)
	}

	loginRes, err := u.Login(dto.LoginRequest{Email: "AFFAN@wasteco.id", Password: "password123"})
	if err != nil || loginRes.Token == "" {
		t.Fatalf("login failed: %v", err)
	}
	if _, err := u.Login(dto.LoginRequest{Email: "affan@wasteco.id", Password: "wrong"}); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("wrong password err = %v, want ErrInvalidCredentials", err)
	}
}
