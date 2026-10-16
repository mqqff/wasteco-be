package usecase

import "github.com/mqqff/wasteco-be/internal/app/user/repository"

type UserUsecase struct {
	userRepo repository.UserRepoItf
}

func NewUserUsecase(r repository.UserRepoItf) *UserUsecase {
	return &UserUsecase{userRepo: r}
}
