package rest

import (
	"errors"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/app/donation/usecase"
	"github.com/mqqff/wasteco-be/internal/domain/dto"
	"github.com/mqqff/wasteco-be/internal/middlewares"
)

type DonationHandler struct {
	donationUsecase usecase.DonationUsecaseItf
	validator       *validator.Validate
}

func NewDonationHandler(r fiber.Router, m middlewares.MiddlewareItf, u usecase.DonationUsecaseItf, v *validator.Validate) {
	handler := &DonationHandler{
		donationUsecase: u,
		validator:       v,
	}

	donations := r.Group("/donations", m.RequireAuth)
	donations.Get("/", handler.GetNearbyRequests)
	donations.Get("/my", handler.GetMyRequests)
	donations.Get("/:id", handler.GetRequestDetail)
	donations.Post("/", handler.CreateRequest)
	donations.Post("/:id/fulfill", handler.Fulfill)
	donations.Post("/:id/confirm-received", handler.ConfirmReceived)
	donations.Delete("/:id", handler.Cancel)
}

func (h *DonationHandler) CreateRequest(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	req := dto.CreateDonationRequest{}
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

	res, err := h.donationUsecase.CreateRequest(userID, req)
	if err != nil {
		return usecaseError(c, err, "Failed to create donation request")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"payload": fiber.Map{
			"donation": res,
		},
	})
}

func (h *DonationHandler) GetNearbyRequests(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	userLat, err := parseOptionalFloat(c.Query("lat"))
	if err != nil {
		return invalidQueryError(c, "invalid lat query parameter")
	}
	userLong, err := parseOptionalFloat(c.Query("long"))
	if err != nil {
		return invalidQueryError(c, "invalid long query parameter")
	}
	radiusKM, err := parseOptionalFloat(c.Query("radius"))
	if err != nil {
		return invalidQueryError(c, "invalid radius query parameter")
	}

	donations, err := h.donationUsecase.GetNearbyRequests(userID, c.Query("search"), userLat, userLong, radiusKM)
	if err != nil {
		return usecaseError(c, err, "Failed to retrieve donation requests")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": fiber.Map{
			"donations": donations,
			"total":     len(donations),
		},
	})
}

func (h *DonationHandler) GetRequestDetail(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return invalidIDError(c)
	}

	userLat, err := parseOptionalFloat(c.Query("lat"))
	if err != nil {
		return invalidQueryError(c, "invalid lat query parameter")
	}
	userLong, err := parseOptionalFloat(c.Query("long"))
	if err != nil {
		return invalidQueryError(c, "invalid long query parameter")
	}

	res, err := h.donationUsecase.GetRequestDetail(userID, id, userLat, userLong)
	if err != nil {
		return usecaseError(c, err, "Failed to retrieve donation request")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": fiber.Map{
			"donation": res,
		},
	})
}

func (h *DonationHandler) GetMyRequests(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	role := c.Query("role")
	switch role {
	case "", "requester", "donor", "all":
	default:
		return invalidQueryError(c, "invalid role query parameter")
	}

	res, err := h.donationUsecase.GetMyRequests(userID, role)
	if err != nil {
		return usecaseError(c, err, "Failed to retrieve your donation requests")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": res,
	})
}

func (h *DonationHandler) Fulfill(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return invalidIDError(c)
	}

	req := dto.FulfillDonationRequest{}
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

	res, err := h.donationUsecase.Fulfill(userID, id, req)
	if err != nil {
		return usecaseError(c, err, "Failed to fulfill donation request")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": fiber.Map{
			"donation": res,
		},
	})
}

func (h *DonationHandler) ConfirmReceived(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return invalidIDError(c)
	}

	res, err := h.donationUsecase.ConfirmReceived(userID, id)
	if err != nil {
		return usecaseError(c, err, "Failed to confirm donation receipt")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": fiber.Map{
			"donation": res,
		},
	})
}

func (h *DonationHandler) Cancel(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return invalidIDError(c)
	}

	if err := h.donationUsecase.Cancel(userID, id); err != nil {
		return usecaseError(c, err, "Failed to cancel donation request")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": fiber.Map{
			"cancelled": true,
		},
	})
}

func parseOptionalFloat(raw string) (float64, error) {
	if raw == "" {
		return 0, nil
	}
	return strconv.ParseFloat(raw, 64)
}

func invalidQueryError(c fiber.Ctx, errMsg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error_code": "bad_request",
		"error":      errMsg,
		"message":    "Invalid query parameter",
	})
}

func invalidIDError(c fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error_code": "bad_request",
		"error":      "invalid donation request id",
		"message":    "Please provide a valid donation request ID",
	})
}

func usecaseError(c fiber.Ctx, err error, message string) error {
	status, code := fiber.StatusInternalServerError, "internal_server_error"

	switch {
	case errors.Is(err, usecase.ErrNotFound):
		status, code, message = fiber.StatusNotFound, "not_found", "Donation request not found"
	case errors.Is(err, usecase.ErrForbidden):
		status, code, message = fiber.StatusForbidden, "forbidden", "You are not allowed to perform this action"
	case errors.Is(err, usecase.ErrInvalidState):
		status, code, message = fiber.StatusConflict, "invalid_state", "This donation request can no longer be updated"
	case errors.Is(err, usecase.ErrPhoneRequired):
		status, code, message = fiber.StatusBadRequest, "phone_number_required", "Please add your WhatsApp number in your profile first"
	case errors.Is(err, usecase.ErrInvalidItems):
		status, code, message = fiber.StatusBadRequest, "validation_error", "Selected items do not belong to this request"
	}

	return c.Status(status).JSON(fiber.Map{
		"error_code": code,
		"error":      err.Error(),
		"message":    message,
	})
}
