package rest

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/app/recipe/usecase"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
	"github.com/mqqff/wasteco-be/internal/middlewares"
)

type RecipeHandler struct {
	recipeUsecase usecase.RecipeUsecaseItf
	validator     *validator.Validate
}

func NewRecipeHandler(r fiber.Router, m middlewares.MiddlewareItf, u usecase.RecipeUsecaseItf, v *validator.Validate) {
	handler := &RecipeHandler{
		recipeUsecase: u,
		validator:     v,
	}

	recipes := r.Group("/recipes", m.RequireAuth)
	recipes.Post("/generate", handler.GenerateRecipes)
	recipes.Post("/cook-confirm", handler.ConfirmCook)
}

func (h *RecipeHandler) GenerateRecipes(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	req := dto.GenerateRecipeRequest{}
	_ = c.Bind().Body(&req)

	recipes, err := h.recipeUsecase.GenerateRecipes(userID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error_code": "recipe_generation_error",
			"error":      err.Error(),
			"message":    "Failed to generate zero-waste recipes with Gemini AI",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": fiber.Map{
			"recipes": recipes,
			"total":   len(recipes),
		},
	})
}

func (h *RecipeHandler) ConfirmCook(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	req := dto.CookConfirmRequest{}
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

	res, err := h.recipeUsecase.ConfirmCook(userID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error_code": "internal_server_error",
			"error":      err.Error(),
			"message":    "Failed to record cooking impact",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": res,
	})
}
