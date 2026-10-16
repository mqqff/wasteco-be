package oauth

import (
	"context"

	"github.com/mqqff/wasteco-be/internal/domain/env"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuthItf interface {
	GenerateAuthLink(state string) string
	ExchangeToken(code string) (*oauth2.Token, error)
}

type OAuthStruct struct {
	config *oauth2.Config
}

var OAuth = getOAuth()

func getOAuth() OAuthItf {
	return &OAuthStruct{config: &oauth2.Config{
		ClientID: env.AppEnv.GoogleClientID, ClientSecret: env.AppEnv.GoogleClientSecret,
		Endpoint: google.Endpoint, RedirectURL: env.AppEnv.GoogleRedirectURL,
		Scopes: []string{"openid", "https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	}}
}

func (o *OAuthStruct) GenerateAuthLink(state string) string {
	return o.config.AuthCodeURL(state)
}

func (o *OAuthStruct) ExchangeToken(code string) (*oauth2.Token, error) {
	return o.config.Exchange(context.Background(), code)
}
