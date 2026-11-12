package rest

import (
	"mime/multipart"
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

	donations := r.Group("/donations")
	donations.Get("/", handler.GetAvailableDonations)
	donations.Get("/my", m.RequireAuth, handler.GetMyDonations)
	donations.Get("/:id", handler.GetDonationDetail)
	donations.Post("/", m.RequireAuth, handler.CreateDonation)
	donations.Patch("/:id/status", m.RequireAuth, handler.UpdateStatus)
	donations.Post("/:id/confirm-handed", m.RequireAuth, handler.ConfirmHandedOver)
	donations.Delete("/:id", m.RequireAuth, handler.CancelDonation)
}

func (h *DonationHandler) CreateDonation(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	req := dto.CreateDonationRequest{}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_code": "bad_request",
			"error":      err.Error(),
			"message":    "Invalid form data",
		})
	}

	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_code": "validation_error",
			"error":      err.Error(),
			"message":    "Validation failed",
		})
	}

	var files []*multipart.FileHeader
	form, err := c.MultipartForm()
	if err == nil && form != nil {
		files = form.File["photos"]
		if len(files) == 0 {
			files = form.File["image"]
		}
	}

	res, err := h.donationUsecase.CreateDonation(userID, req, files)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error_code": "internal_server_error",
			"error":      err.Error(),
			"message":    "Failed to create donation",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"payload": fiber.Map{
			"donation": res,
		},
	})
}

func (h *DonationHandler) GetAvailableDonations(c fiber.Ctx) error {
	category := c.Query("category")
	search := c.Query("search")

	userLat, _ := strconv.ParseFloat(c.Query("lat"), 64)
	userLong, _ := strconv.ParseFloat(c.Query("long"), 64)
	radiusKM, _ := strconv.ParseFloat(c.Query("radius"), 64)
	if radiusKM == 0 {
		radiusKM = 5.0
	}

	donations, err := h.donationUsecase.GetAvailableDonations(category, search, userLat, userLong, radiusKM)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error_code": "internal_server_error",
			"error":      err.Error(),
			"message":    "Failed to retrieve donations",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": fiber.Map{
			"donations": donations,
			"total":     len(donations),
		},
	})
}

func (h *DonationHandler) GetDonationDetail(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_code": "bad_request",
			"error":      "invalid donation id",
			"message":    "Please provide a valid donation ID",
		})
	}

	userLat, _ := strconv.ParseFloat(c.Query("lat"), 64)
	userLong, _ := strconv.ParseFloat(c.Query("long"), 64)

	res, err := h.donationUsecase.GetDonationDetail(id, userLat, userLong)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error_code": "not_found",
			"error":      "donation not found",
			"message":    "Donation not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": fiber.Map{
			"donation": res,
		},
	})
}

func (h *DonationHandler) GetMyDonations(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	res, err := h.donationUsecase.GetMyDonations(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error_code": "internal_server_error",
			"error":      err.Error(),
			"message":    "Failed to retrieve user donations",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": res,
	})
}

func (h *DonationHandler) UpdateStatus(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_code": "bad_request",
			"error":      "invalid donation id",
			"message":    "Please provide a valid donation ID",
		})
	}

	req := dto.UpdateDonationStatusRequest{}
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

	res, err := h.donationUsecase.UpdateStatus(userID, id, req.Status)
	if err != nil {
		if err.Error() == "unauthorized" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error_code": "forbidden",
				"error":      "forbidden",
				"message":    "You cannot edit this donation",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error_code": "internal_server_error",
			"error":      err.Error(),
			"message":    "Failed to update donation status",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": fiber.Map{
			"donation": res,
		},
	})
}

func (h *DonationHandler) ConfirmHandedOver(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_code": "bad_request",
			"error":      "invalid donation id",
			"message":    "Please provide a valid donation ID",
		})
	}

	res, err := h.donationUsecase.ConfirmHandedOver(userID, id)
	if err != nil {
		if err.Error() == "unauthorized" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error_code": "forbidden",
				"error":      "forbidden",
				"message":    "You cannot confirm this donation",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error_code": "internal_server_error",
			"error":      err.Error(),
			"message":    "Failed to confirm donation handover",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": fiber.Map{
			"confirmed": true,
			"donation":  res,
			"message":   "Donasi berhasil diserahkan dan dampak penyelamatan makanan telah dicatat!",
		},
	})
}

func (h *DonationHandler) CancelDonation(c fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_code": "bad_request",
			"error":      "invalid donation id",
			"message":    "Please provide a valid donation ID",
		})
	}

	err = h.donationUsecase.CancelDonation(userID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error_code": "internal_server_error",
			"error":      err.Error(),
			"message":    "Failed to cancel donation",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"payload": fiber.Map{
			"cancelled": true,
		},
	})
}
