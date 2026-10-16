package middlewares

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/keyauth"
	"github.com/mqqff/wasteco-be/internal/domain/env"
)

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
