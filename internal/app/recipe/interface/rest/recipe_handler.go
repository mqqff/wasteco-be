package rest

import (
	"errors"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/app/recipe/usecase"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
	"github.com/mqqff/wasteco-be/internal/middlewares"
	"github.com/mqqff/wasteco-be/pkg/media"
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
	recipes.Post("/generate", m.Limit(15, time.Minute), handler.GenerateRecipes)
	recipes.Post("/cook-confirm", handler.ConfirmCook)
}

func (h *RecipeHandler) GenerateRecipes(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	if strings.HasPrefix(strings.ToLower(c.Get(fiber.HeaderContentType)), "multipart/form-data") {
		return h.generateRecipesFromImage(c, userID)
	}

	req := dto.GenerateRecipeRequest{}
	if len(c.Body()) > 0 {
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
	}

	recipes, err := h.recipeUsecase.GenerateRecipes(userID, req)
	if err != nil {
		return recipeGenerationError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": fiber.Map{
			"recipes": recipes,
			"total":   len(recipes),
		},
	})
}

func (h *RecipeHandler) generateRecipesFromImage(c fiber.Ctx, userID uuid.UUID) error {
	data, mimeType, err := media.ReadImage(c, "image")
	if err != nil {
		return media.WriteError(c, err)
	}

	req := dto.GenerateRecipeRequest{}
	if form, err := c.MultipartForm(); err == nil && form != nil {
		req.IngredientIDs = parseMultipartSlice(form.Value["ingredient_ids"])
		req.CustomIngredients = parseMultipartSlice(form.Value["custom_ingredients"])
		if err := h.validator.Struct(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error_code": "validation_error",
				"error":      err.Error(),
				"message":    "Validation failed",
			})
		}
	}

	res, err := h.recipeUsecase.GenerateRecipesFromImage(userID, data, mimeType, req)
	if err != nil {
		return recipeGenerationError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": fiber.Map{
			"detected_ingredients": res.DetectedIngredients,
			"recipes":              res.Recipes,
			"total":                len(res.Recipes),
		},
	})
}

func recipeGenerationError(c fiber.Ctx, err error) error {
	if errors.Is(err, usecase.ErrNoIngredients) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_code": "bad_request",
			"error":      err.Error(),
			"message":    "Please add or select at least one active ingredient",
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error_code": "recipe_generation_error",
		"error":      err.Error(),
		"message":    "Failed to generate zero-waste recipes with Gemini AI",
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

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"payload": res,
	})
}

func parseMultipartSlice(values []string) []string {
	var out []string
	for _, raw := range values {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			var parsed []string
			if err := sonic.Unmarshal([]byte(trimmed), &parsed); err == nil {
				for _, item := range parsed {
					if v := strings.TrimSpace(item); v != "" {
						out = append(out, v)
					}
				}
				continue
			}
		}
		for _, part := range strings.Split(trimmed, ",") {
			if v := strings.TrimSpace(part); v != "" {
				out = append(out, v)
			}
		}
	}
	return out
}
