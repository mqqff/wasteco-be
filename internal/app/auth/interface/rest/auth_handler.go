package rest

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/mqqff/wasteco-be/internal/app/auth/usecase"
	"github.com/mqqff/wasteco-be/internal/domain/env"
)

type AuthHandler struct {
	authUsecase usecase.AuthUsecaseItf
}

func NewAuthHandler(r fiber.Router, u usecase.AuthUsecaseItf) {
	handler := AuthHandler{authUsecase: u}
	r = r.Group("/auth/google/oauth")
	r.Get("/redirect", handler.HandleRedirect)
}

func (h *AuthHandler) HandleRedirect(c fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name: "oauth2state", Value: env.AppEnv.OAuthState,
		MaxAge: 3600, Expires: time.Now().Add(5 * time.Minute),
		HTTPOnly: true, Secure: env.AppEnv.AppEnv != "development",
	})

	url, err := h.authUsecase.HandleRedirect(env.AppEnv.OAuthState)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"payload": fiber.Map{"redirect_url": url}})
}
