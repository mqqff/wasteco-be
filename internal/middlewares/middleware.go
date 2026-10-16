package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mqqff/wasteco-be/pkg/jwt"
)

type MiddlewareItf interface {
	RequireAuth(c fiber.Ctx) error
}

type Middleware struct {
	jwt jwt.JWTIf
}

func NewMiddleware(jwt jwt.JWTIf) MiddlewareItf {
	return &Middleware{
		jwt: jwt,
	}
}
