package middlewares

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/keyauth"
	"github.com/mqqff/wasteco-be/internal/domain/env"
	"github.com/mqqff/wasteco-be/pkg/jwt"
	"strings"
)

func (m *Middleware) RequireAuth(c fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if len(authHeader) < 1 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error_code": "missing_authentication_token",
			"error":      "missing authentication token",
			"message":    "You are not authorized",
		})
	}

	bearer := strings.Split(authHeader, " ")
	if len(bearer) < 2 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error_code": "invalid_authentication_format",
			"error":      "invalid authentication token format",
			"message":    "You are not authorized",
		})
	}

	token := bearer[1]

	claims := jwt.Claims{}
	err := m.jwt.Decode(token, &claims)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error_code": "invalid_authentication",
			"error":      err.Error(),
			"message":    "You are not authorized",
		})
	}

	c.Locals("userID", claims.UserID)

	return c.Next()
}

func RequireAPIKey() fiber.Handler {
	return keyauth.New(keyauth.Config{
		Extractor: extractors.FromHeader("x-api-key"),
		Validator: func(c fiber.Ctx, key string) (bool, error) {
			hashedAPIKey := sha256.Sum256([]byte(env.AppEnv.APIKey))
			hashedKey := sha256.Sum256([]byte(key))

			if subtle.ConstantTimeCompare(hashedAPIKey[:], hashedKey[:]) == 1 {
				return true, nil
			}

			return false, keyauth.ErrMissingOrMalformedAPIKey
		},
		ErrorHandler: func(c fiber.Ctx, err error) error {
			if errors.Is(err, keyauth.ErrMissingOrMalformedAPIKey) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error_code": "missing_or_malformed_api_key",
					"error":      err.Error(),
					"message":    "please provide valid api key",
				})
			}

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error_code": "invalid_api_key",
				"error":      "invalid or expired api key",
				"message":    "please provide valid api key",
			})
		},
	})
}
