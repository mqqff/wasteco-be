package middlewares

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/utils/v2"
)

func (m *Middleware) Limit(max int, exp time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: exp,
		KeyGenerator: func(c fiber.Ctx) string {
			if ip := strings.TrimSpace(c.Get("Cf-Connecting-Ip")); ip != "" {
				return ip
			}
			if xff := c.Get("x-forwarded-for"); xff != "" {
				if first := strings.TrimSpace(strings.Split(xff, ",")[0]); first != "" {
					return first
				}
			}
			return c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error_code": "too_many_request",
				"error":      utils.StatusMessage(fiber.StatusTooManyRequests),
				"message":    "too many request, try again later",
			})
		},
	})
}
