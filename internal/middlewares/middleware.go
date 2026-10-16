package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mqqff/wasteco-be/pkg/jwt"
	"time"
)

type MiddlewareItf interface {
	RequireAuth(c fiber.Ctx) error
	Limit(max int, exp time.Duration) fiber.Handler
}

type Middleware struct {
	jwt jwt.JWTIf
}

func NewMiddleware(jwt jwt.JWTIf) MiddlewareItf {
	return &Middleware{
		jwt: jwt,
	}
}
