package middlewares

import (
	fiberzerolog "github.com/gofiber/contrib/v3/zerolog"
	"github.com/gofiber/fiber/v3"
	"github.com/mqqff/wasteco-be/internal/domain/env"
	"github.com/mqqff/wasteco-be/pkg/log"
)

func Logger() fiber.Handler {
	fields := []string{
		"referrer",
		"ip",
		"url",
		"latency",
		"status",
		"method",
		"error",
	}

	if env.AppEnv.AppEnv != "production" {
		fields = append(fields, "body")
		fields = append(fields, "reqHeaders")
		fields = append(fields, "resHeaders")
	}

	logger := log.GetLogger()
	config := fiberzerolog.Config{
		FieldsSnakeCase: true,
		Logger:          logger,
		Fields:          fields,
		Messages: []string{
			"[LoggerMiddleware.LoggerConfig] Server error",
			"[LoggerMiddleware.LoggerConfig] Client error",
			"[LoggerMiddleware.LoggerConfig] Success",
		},
	}

	return fiberzerolog.New(config)
}
