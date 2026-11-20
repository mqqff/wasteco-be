package media

import (
	"errors"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v3"
)

var (
	ErrMissingImage         = errors.New("missing image file")
	ErrReadImage            = errors.New("failed to read uploaded image")
	ErrUnsupportedImageType = errors.New("unsupported image type")
)

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

type unsupportedError struct {
	mimeType string
}

func (e unsupportedError) Error() string {
	return "unsupported file type: " + e.mimeType
}

func (e unsupportedError) Unwrap() error {
	return ErrUnsupportedImageType
}

func ReadImage(c fiber.Ctx, field string) ([]byte, string, error) {
	fileHeader, err := c.FormFile(field)
	if err != nil {
		return nil, "", fmtError(ErrMissingImage, err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, "", fmtError(ErrReadImage, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, "", fmtError(ErrReadImage, err)
	}

	mimeType := http.DetectContentType(data)
	if !allowedImageTypes[mimeType] {
		return nil, "", unsupportedError{mimeType: mimeType}
	}

	return data, mimeType, nil
}

func WriteError(c fiber.Ctx, err error) error {
	code, msg := "bad_request", "Failed to read uploaded photo"

	switch {
	case errors.Is(err, ErrMissingImage):
		msg = "Please upload a photo in the 'image' field"
	case errors.Is(err, ErrUnsupportedImageType):
		code, msg = "validation_error", "Photo must be JPEG, PNG, or WebP"
	}

	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error_code": code,
		"error":      err.Error(),
		"message":    msg,
	})
}

type wrappedError struct {
	sentinel error
	cause    error
}

func (e wrappedError) Error() string {
	return e.cause.Error()
}

func (e wrappedError) Unwrap() error {
	return e.sentinel
}

func fmtError(sentinel, cause error) error {
	return wrappedError{sentinel: sentinel, cause: cause}
}
