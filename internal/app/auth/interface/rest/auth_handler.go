package rest

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/mqqff/wasteco-be/internal/app/auth/usecase"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
	"github.com/mqqff/wasteco-be/internal/domain/env"
)

type AuthHandler struct {
	authUsecase usecase.AuthUsecaseItf
	validator   *validator.Validate
}

func NewAuthHandler(r fiber.Router, u usecase.AuthUsecaseItf, v *validator.Validate) {
	handler := &AuthHandler{
		authUsecase: u,
		validator:   v,
	}

	auth := r.Group("/auth")
	auth.Post("/register", handler.HandleRegister)
	auth.Post("/login", handler.HandleLogin)

	google := auth.Group("/google/oauth")
	google.Get("/redirect", handler.HandleRedirect)
	google.Post("/callback", handler.HandleCallback)
}

func (h *AuthHandler) HandleRegister(c fiber.Ctx) error {
	req := dto.RegisterRequest{}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_code": "bad_request",
			"error":      err.Error(),
			"message":    "Invalid request body",
		})
	}

	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_code": "validation_error",
			"error":      err.Error(),
			"message":    "Validation failed",
		})
	}

	res, err := h.authUsecase.Register(req)
	if err != nil {
		if err.Error() == "email already registered" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error_code": "email_already_registered",
				"error":      err.Error(),
				"message":    "Email is already registered",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error_code": "internal_server_error",
			"error":      err.Error(),
			"message":    "Failed to register user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"payload": res,
	})
}

func (h *AuthHandler) HandleLogin(c fiber.Ctx) error {
	req := dto.LoginRequest{}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_code": "bad_request",
			"error":      err.Error(),
			"message":    "Invalid request body",
		})
	}

	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_code": "validation_error",
			"error":      err.Error(),
			"message":    "Validation failed",
		})
	}

	res, err := h.authUsecase.Login(req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error_code": "invalid_credentials",
			"error":      err.Error(),
			"message":    "Invalid email or password",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": res,
	})
}

func (h *AuthHandler) HandleRedirect(c fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "oauth2state",
		Value:    env.AppEnv.OAuthState,
		MaxAge:   3600,
		Expires:  time.Now().Add(5 * time.Minute),
		HTTPOnly: true,
		Secure:   env.AppEnv.AppEnv != "development",
	})

	url, _ := h.authUsecase.HandleRedirect(env.AppEnv.OAuthState)

	payload := fiber.Map{
		"redirect_url": url,
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": payload,
	})
}

func (h *AuthHandler) HandleCallback(c fiber.Ctx) error {
	req := dto.GoogleCallbackRequest{}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_code": "bad_request",
			"error":      err.Error(),
			"message":    "Invalid body request",
		})
	}

	err := h.validator.Struct(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_code": "validation_error",
			"error":      err.Error(),
			"message":    "Validation error",
		})
	}

	token, err := h.authUsecase.HandleCallback(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error_code": "internal_server_error",
			"error":      err.Error(),
			"message":    "Something happened with our end. Please try again later",
		})
	}

	payload := fiber.Map{
		"token": token,
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": payload,
	})
}
