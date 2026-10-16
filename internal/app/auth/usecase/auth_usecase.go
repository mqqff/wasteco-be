package usecase

import "github.com/mqqff/wasteco-be/pkg/oauth"

type AuthUsecaseItf interface {
	HandleRedirect(state string) (string, error)
}

type AuthUsecase struct {
	oauth oauth.OAuthItf
}

func NewAuthUsecase(o oauth.OAuthItf) AuthUsecaseItf {
	return &AuthUsecase{oauth: o}
}

func (u *AuthUsecase) HandleRedirect(state string) (string, error) {
	return u.oauth.GenerateAuthLink(state), nil
}
