package middlewares

import (
	fiberzerolog "github.com/gofiber/contrib/v3/zerolog"
	"github.com/gofiber/fiber/v3"
	"github.com/mqqff/wasteco-be/pkg/log"
)

func Logger() fiber.Handler {
	return fiberzerolog.New(fiberzerolog.Config{
		FieldsSnakeCase: true,
		Logger:          log.GetLogger(),
		Fields:          []string{"referrer", "ip", "url", "latency", "status", "method", "error"},
	})
}
