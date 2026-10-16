package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"time"
)

func (m *Middleware) Limit(max int, exp time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: exp,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.Get("x-forwarded-for")
		},
	})
}
