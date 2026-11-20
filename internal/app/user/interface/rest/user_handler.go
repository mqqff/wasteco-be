package rest

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/app/user/usecase"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
	"github.com/mqqff/wasteco-be/internal/middlewares"
)

type UserHandlerItf interface {
}

type UserHandler struct {
	userUsecase usecase.UserUsecaseItf
	validator   *validator.Validate
}

func NewUserHandler(r fiber.Router, m middlewares.MiddlewareItf, u usecase.UserUsecaseItf, v *validator.Validate) {
	handler := &UserHandler{
		userUsecase: u,
		validator:   v,
	}

	me := r.Group("/me", m.RequireAuth)
	me.Get("/", handler.GetProfile)
	me.Put("/", handler.UpdateProfile)
}

func (h *UserHandler) GetProfile(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	user, err := h.userUsecase.GetProfile(userID)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error_code": "not_found",
				"error":      err.Error(),
				"message":    "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error_code": "internal_server_error",
			"error":      err.Error(),
			"message":    "Failed to retrieve user profile",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": fiber.Map{
			"user": user,
		},
	})
}

func (h *UserHandler) UpdateProfile(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	req := dto.UpdateProfileRequest{}
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

	user, err := h.userUsecase.UpdateProfile(userID, req)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error_code": "not_found",
				"error":      err.Error(),
				"message":    "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error_code": "internal_server_error",
			"error":      err.Error(),
			"message":    "Failed to update profile",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": fiber.Map{
			"user": user,
		},
	})
}
