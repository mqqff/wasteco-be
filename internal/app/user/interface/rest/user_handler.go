package rest

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/app/user/usecase"
	"github.com/mqqff/wasteco-be/internal/middlewares"
)

type UserHandler struct {
	userUsecase usecase.UserUsecaseItf
}

func NewUserHandler(r fiber.Router, m middlewares.MiddlewareItf, u usecase.UserUsecaseItf) {
	handler := &UserHandler{userUsecase: u}
	r.Get("/me", m.RequireAuth, handler.GetProfile)
}
func (h *UserHandler) GetProfile(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	user, err := h.userUsecase.GetProfile(userID)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"payload": fiber.Map{"user": user}})
}
