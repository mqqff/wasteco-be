package rest

import (
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/app/pantry/usecase"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
	"github.com/mqqff/wasteco-be/internal/middlewares"
)

type PantryHandler struct {
	pantryUsecase usecase.PantryUsecaseItf
	validator     *validator.Validate
}

func NewPantryHandler(r fiber.Router, m middlewares.MiddlewareItf, u usecase.PantryUsecaseItf, v *validator.Validate) {
	handler := &PantryHandler{
		pantryUsecase: u,
		validator:     v,
	}

	pantry := r.Group("/pantry", m.RequireAuth)
	pantry.Get("/", handler.GetIngredients)
	pantry.Post("/", handler.AddIngredient)
	pantry.Post("/bulk", handler.AddIngredients)
	pantry.Post("/scan", handler.ScanIngredients)
	pantry.Put("/:id", handler.UpdateIngredient)
	pantry.Delete("/:id", handler.DeleteIngredient)
}

func (h *PantryHandler) GetIngredients(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	category := c.Query("category")

	items, err := h.pantryUsecase.GetIngredients(userID, category)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error_code": "internal_server_error",
			"error":      err.Error(),
			"message":    "Failed to retrieve pantry ingredients",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": fiber.Map{
			"ingredients": items,
			"total":       len(items),
		},
	})
}

func (h *PantryHandler) AddIngredient(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	req := dto.CreatePantryItemRequest{}
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

	res, err := h.pantryUsecase.AddIngredient(userID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error_code": "internal_server_error",
			"error":      err.Error(),
			"message":    "Failed to add pantry ingredient",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"payload": fiber.Map{
			"ingredient": res,
		},
	})
}

func (h *PantryHandler) AddIngredients(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	req := dto.BulkCreatePantryItemRequest{}
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

	res, err := h.pantryUsecase.AddIngredients(userID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error_code": "internal_server_error",
			"error":      err.Error(),
			"message":    "Failed to add pantry ingredients",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"payload": fiber.Map{
			"ingredients": res,
			"total":       len(res),
		},
	})
}

var allowedScanTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

func (h *PantryHandler) ScanIngredients(c fiber.Ctx) error {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_code": "bad_request",
			"error":      err.Error(),
			"message":    "Please upload a photo in the 'image' field",
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_code": "bad_request",
			"error":      err.Error(),
			"message":    "Failed to read uploaded photo",
		})
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_code": "bad_request",
			"error":      err.Error(),
			"message":    "Failed to read uploaded photo",
		})
	}

	mimeType := http.DetectContentType(data)
	if !allowedScanTypes[mimeType] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_code": "validation_error",
			"error":      "unsupported file type: " + mimeType,
			"message":    "Photo must be JPEG, PNG, or WebP",
		})
	}

	items, err := h.pantryUsecase.ScanIngredients(data, mimeType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error_code": "internal_server_error",
			"error":      err.Error(),
			"message":    "Failed to detect ingredients from photo",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": fiber.Map{
			"ingredients": items,
			"total":       len(items),
		},
	})
}

func (h *PantryHandler) UpdateIngredient(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_code": "bad_request",
			"error":      "invalid ingredient id",
			"message":    "Please provide a valid ingredient ID",
		})
	}

	req := dto.UpdatePantryItemRequest{}
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

	res, err := h.pantryUsecase.UpdateIngredient(userID, id, req)
	if err != nil {
		if err.Error() == "unauthorized" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error_code": "forbidden",
				"error":      "forbidden",
				"message":    "You cannot edit this ingredient",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error_code": "internal_server_error",
			"error":      err.Error(),
			"message":    "Failed to update pantry ingredient",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": fiber.Map{
			"ingredient": res,
		},
	})
}

func (h *PantryHandler) DeleteIngredient(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_code": "bad_request",
			"error":      "invalid ingredient id",
			"message":    "Please provide a valid ingredient ID",
		})
	}

	err = h.pantryUsecase.DeleteIngredient(userID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error_code": "internal_server_error",
			"error":      err.Error(),
			"message":    "Failed to delete pantry ingredient",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": fiber.Map{
			"deleted": true,
		},
	})
}
