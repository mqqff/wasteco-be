package rest

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/app/impact/usecase"
	"github.com/mqqff/wasteco-be/internal/middlewares"
)

type ImpactHandler struct {
	impactUsecase usecase.ImpactUsecaseItf
}

func NewImpactHandler(r fiber.Router, m middlewares.MiddlewareItf, u usecase.ImpactUsecaseItf) {
	handler := &ImpactHandler{
		impactUsecase: u,
	}

	impact := r.Group("/impact", m.RequireAuth)
	impact.Get("/summary", handler.GetSummary)
	impact.Get("/history", handler.GetHistory)
	impact.Get("/history/:id", handler.GetDetail)
}

func (h *ImpactHandler) GetSummary(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	month, err := parseOptionalInt(c.Query("month"))
	if err != nil || (c.Query("month") != "" && (month < 1 || month > 12)) {
		return invalidQueryError(c, "month must be an integer between 1 and 12")
	}
	year, err := parseOptionalInt(c.Query("year"))
	if err != nil || (c.Query("year") != "" && (year < 2000 || year > 2100)) {
		return invalidQueryError(c, "year must be an integer between 2000 and 2100")
	}

	summary, err := h.impactUsecase.GetMonthlySummary(userID, month, year)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error_code": "internal_server_error",
			"error":      err.Error(),
			"message":    "Failed to retrieve impact summary",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": summary,
	})
}

func (h *ImpactHandler) GetHistory(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	page, err := parseOptionalInt(c.Query("page"))
	if err != nil || (c.Query("page") != "" && page < 1) {
		return invalidQueryError(c, "page must be a positive integer")
	}
	limit, err := parseOptionalInt(c.Query("limit"))
	if err != nil || (c.Query("limit") != "" && limit < 1) {
		return invalidQueryError(c, "limit must be a positive integer")
	}

	history, err := h.impactUsecase.GetHistory(userID, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error_code": "internal_server_error",
			"error":      err.Error(),
			"message":    "Failed to retrieve impact history",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": history,
	})
}

func (h *ImpactHandler) GetDetail(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_code": "bad_request",
			"error":      "invalid activity id",
			"message":    "Please provide a valid activity ID",
		})
	}

	detail, err := h.impactUsecase.GetDetail(userID, id)
	if err != nil {
		if errors.Is(err, usecase.ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error_code": "forbidden",
				"error":      "forbidden",
				"message":    "You cannot view this activity",
			})
		}
		if errors.Is(err, usecase.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error_code": "not_found",
				"error":      "activity not found",
				"message":    "Impact activity not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error_code": "internal_server_error",
			"error":      err.Error(),
			"message":    "Failed to retrieve impact activity",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": fiber.Map{
			"activity": detail,
		},
	})
}

func parseOptionalInt(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	return strconv.Atoi(raw)
}

func invalidQueryError(c fiber.Ctx, errMsg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error_code": "bad_request",
		"error":      errMsg,
		"message":    "Invalid query parameter",
	})
}
